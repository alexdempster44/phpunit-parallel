package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexdempster44/phpunit-parallel/internal/output"
)

const tickInterval = 100 * time.Millisecond

func (m *Model) Init() tea.Cmd {
	return tick()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		if m.phase == PhaseRunning || m.phase == PhaseCleanup {
			return m, tick()
		}
		return m, nil

	case WorkerStartMsg:
		m.handleWorkerStart(msg)
		return m, nil

	case TestStartMsg:
		m.handleTestStart(msg)
		return m, nil

	case TestPassMsg:
		m.handleTestPass(msg)
		return m, nil

	case TestFailMsg:
		m.handleTestFail(msg)
		return m, nil

	case TestDeprecationMsg:
		m.handleTestDeprecation(msg)
		return m, nil

	case TestSkipMsg:
		m.handleTestSkip(msg)
		return m, nil

	case WorkerCompleteMsg:
		m.handleWorkerComplete(msg)
		return m, nil

	case TestCountMsg:
		m.handleTestCount(msg)
		return m, nil

	case CleanupProgressMsg:
		if m.phase == PhaseRunning {
			m.endTime = time.Now()
		}
		m.phase = PhaseCleanup
		m.cleanupCompleted = msg.Completed
		m.cleanupTotal = msg.Total
		return m, tick()

	case CopyNoticeExpiredMsg:
		m.copyNotice = ""
		return m, nil

	case FinishMsg:
		m.phase = PhaseComplete
		if m.endTime.IsZero() {
			m.endTime = time.Now()
		}
		m.activePanel = PanelErrors
		return m, nil

	case RetryStartMsg:
		m.retryAttempt = msg.Attempt
		m.phase = PhaseRunning
		m.workers = make(map[int]*WorkerNode)
		m.workerOrder = make([]int, 0, msg.WorkerCount)
		for _, id := range msg.WorkerIDs {
			m.workers[id] = &WorkerNode{
				ID:    id,
				Tests: make([]*TestNode, 0),
			}
			m.workerOrder = append(m.workerOrder, id)
		}
		m.errors = make([]ErrorEntry, 0)
		m.deprecations = make([]DeprecationEntry, 0)
		m.totalComplete = 0
		m.totalFailed = 0
		m.totalSkipped = 0
		m.totalDeprecations = 0
		m.failedWorkers = 0
		m.testCount = msg.TestCount
		m.workerCount = msg.WorkerCount
		m.hasTestCount = false
		m.startTime = time.Now()
		m.endTime = time.Time{}
		m.errorCursor = 0
		m.errorOffset = 0
		m.runningCursor = 0
		m.runningOffset = 0
		m.workersOffset = 0
		m.cleanupCompleted = 0
		m.cleanupTotal = 0
		m.activePanel = PanelErrors
		return m, tick()

	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := DefaultKeyMap()

	if m.showCopyModal {
		return m.handleCopyModalKey(msg, keys)
	}

	switch {
	case key.Matches(msg, keys.Retry):
		if m.phase == PhaseComplete && m.hasFailed() && m.actionCh != nil {
			m.actionCh <- output.ActionRetry
			m.phase = PhaseRunning
			return m, tick()
		}
		return m, nil

	case key.Matches(msg, keys.RerunAll):
		if m.phase == PhaseComplete && m.actionCh != nil {
			m.actionCh <- output.ActionRerunAll
			m.phase = PhaseRunning
			return m, tick()
		}
		return m, nil

	case key.Matches(msg, keys.Quit):
		if m.phase == PhaseComplete || m.phase == PhaseExploring || msg.String() == "ctrl+c" {
			if m.phase == PhaseComplete && m.actionCh != nil {
				m.actionCh <- output.ActionQuit
			}
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case key.Matches(msg, keys.Tab):
		switch m.activePanel {
		case PanelWorkers:
			m.activePanel = PanelErrors
		case PanelErrors:
			m.activePanel = PanelWorkers
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		m.moveCursor(-1)
		return m, nil

	case key.Matches(msg, keys.Down):
		m.moveCursor(1)
		return m, nil

	case key.Matches(msg, keys.Enter):
		m.toggle()
		return m, nil

	case key.Matches(msg, keys.Right):
		m.expandError(true)
		return m, nil

	case key.Matches(msg, keys.Left):
		m.expandError(false)
		return m, nil

	case key.Matches(msg, keys.PageUp):
		m.moveCursor(-10)
		return m, nil

	case key.Matches(msg, keys.PageDown):
		m.moveCursor(10)
		return m, nil

	case key.Matches(msg, keys.CopyAll):
		if len(m.errors)+len(m.deprecations) > 0 {
			m.showCopyModal = true
			m.copyModalCursor = 0
			m.copyModalErrors = len(m.errors) > 0
			m.copyModalDeprecations = len(m.deprecations) > 0
		}
		return m, nil

	case key.Matches(msg, keys.Copy):
		return m.copyError()

	}

	return m, nil
}

func (m *Model) copyModalItemCount() int {
	n := 0
	if len(m.errors) > 0 {
		n++
	}
	if len(m.deprecations) > 0 {
		n++
	}
	return n + 2 // checkboxes + 2 buttons
}

func (m *Model) handleCopyModalKey(msg tea.KeyMsg, keys KeyMap) (tea.Model, tea.Cmd) {
	itemCount := m.copyModalItemCount()
	if m.copyModalCursor >= itemCount {
		m.copyModalCursor = itemCount - 1
	}

	switch {
	case key.Matches(msg, keys.Up):
		if m.copyModalCursor > 0 {
			m.copyModalCursor--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.copyModalCursor < itemCount-1 {
			m.copyModalCursor++
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		checkboxCount := itemCount - 2
		if m.copyModalCursor < checkboxCount {
			// Toggle checkbox
			m.toggleCopyModalCheckbox(m.copyModalCursor)
		} else {
			// Button press
			m.showCopyModal = false
			namesOnly := m.copyModalCursor == checkboxCount
			return m.copyAllEntries(namesOnly)
		}
		return m, nil

	case key.Matches(msg, keys.Quit), key.Matches(msg, keys.CopyAll):
		m.showCopyModal = false
		return m, nil

	default:
		if msg.String() == "esc" {
			m.showCopyModal = false
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) toggleCopyModalCheckbox(index int) {
	// Map index to the correct checkbox based on which are visible
	i := 0
	if len(m.errors) > 0 {
		if i == index {
			m.copyModalErrors = !m.copyModalErrors
			return
		}
		i++
	}
	if len(m.deprecations) > 0 {
		if i == index {
			m.copyModalDeprecations = !m.copyModalDeprecations
			return
		}
	}
}

func (m *Model) copyAllEntries(namesOnly bool) (tea.Model, tea.Cmd) {
	var entries []string
	var count int

	formatEntry := func(name, message, details string) string {
		var parts []string
		parts = append(parts, name)
		if message != "" {
			parts = append(parts, message)
		}
		if details != "" {
			parts = append(parts, details)
		}
		return strings.Join(parts, "\n\n")
	}

	if m.copyModalErrors {
		for _, e := range m.errors {
			if namesOnly {
				entries = append(entries, e.TestName)
			} else {
				entries = append(entries, formatEntry(e.TestName, e.Message, e.Details))
			}
		}
		count += len(m.errors)
	}

	if m.copyModalDeprecations {
		for _, d := range m.deprecations {
			if namesOnly {
				entries = append(entries, d.TestName)
			} else {
				entries = append(entries, formatEntry(d.TestName, d.Message, d.Details))
			}
		}
		count += len(m.deprecations)
	}

	if len(entries) == 0 {
		m.copyNotice = "Nothing to copy"
		return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
			return CopyNoticeExpiredMsg{}
		})
	}

	var text string
	if namesOnly {
		text = strings.Join(entries, "\n")
	} else {
		text = strings.Join(entries, "\n\n---\n\n")
	}

	if err := clipboard.WriteAll(text); err != nil {
		m.copyNotice = fmt.Sprintf("Copy failed: %s", err)
	} else {
		m.copyNotice = fmt.Sprintf("Copied %d entries to clipboard!", count)
	}

	return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return CopyNoticeExpiredMsg{}
	})
}

func (m *Model) moveCursor(delta int) {
	switch m.activePanel {
	case PanelWorkers:
		m.workersOffset += delta

	case PanelErrors:
		maxCursor := len(m.errors) + len(m.deprecations) - 1
		m.errorCursor += delta
		if m.errorCursor < 0 {
			m.errorCursor = 0
		}
		if m.errorCursor > maxCursor && maxCursor >= 0 {
			m.errorCursor = maxCursor
		}
	}
}

func (m *Model) copyError() (tea.Model, tea.Cmd) {
	totalEntries := len(m.errors) + len(m.deprecations)
	if m.activePanel != PanelErrors || totalEntries == 0 {
		return m, nil
	}

	if m.errorCursor < 0 || m.errorCursor >= totalEntries {
		return m, nil
	}

	var testName, message, details string
	if m.errorCursor < len(m.errors) {
		e := m.errors[m.errorCursor]
		testName, message, details = e.TestName, e.Message, e.Details
	} else {
		d := m.deprecations[m.errorCursor-len(m.errors)]
		testName, message, details = d.TestName, d.Message, d.Details
	}

	var parts []string
	parts = append(parts, testName)
	if message != "" {
		parts = append(parts, message)
	}
	if details != "" {
		parts = append(parts, details)
	}
	text := strings.Join(parts, "\n\n")

	if err := clipboard.WriteAll(text); err != nil {
		m.copyNotice = fmt.Sprintf("Copy failed: %s", err)
	} else {
		m.copyNotice = "Copied to clipboard!"
	}

	return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return CopyNoticeExpiredMsg{}
	})
}

func (m *Model) toggle() {
	if m.activePanel == PanelErrors {
		if m.errorCursor >= 0 && m.errorCursor < len(m.errors) {
			m.errors[m.errorCursor].Expanded = !m.errors[m.errorCursor].Expanded
		} else if idx := m.errorCursor - len(m.errors); idx >= 0 && idx < len(m.deprecations) {
			m.deprecations[idx].Expanded = !m.deprecations[idx].Expanded
		}
	}
}

func (m *Model) expandError(expand bool) {
	if m.activePanel == PanelErrors {
		if m.errorCursor >= 0 && m.errorCursor < len(m.errors) {
			m.errors[m.errorCursor].Expanded = expand
		} else if idx := m.errorCursor - len(m.errors); idx >= 0 && idx < len(m.deprecations) {
			m.deprecations[idx].Expanded = expand
		}
	}
}

func (m *Model) handleWorkerStart(msg WorkerStartMsg) {
	if w, ok := m.workers[msg.WorkerID]; ok {
		w.TestFiles = msg.TestCount
		w.Total = msg.TestCount
	}
}

func (m *Model) handleTestStart(msg TestStartMsg) {
	w := m.workers[msg.WorkerID]
	if w == nil {
		return
	}

	for _, t := range w.Tests {
		if t.Key == msg.TestKey {
			t.Status = StatusRunning
			return
		}
	}

	w.Tests = append(w.Tests, &TestNode{
		Key:      msg.TestKey,
		Name:     msg.DisplayName,
		Status:   StatusRunning,
		WorkerID: msg.WorkerID,
	})
}

func (m *Model) handleTestPass(msg TestPassMsg) {
	w := m.workers[msg.WorkerID]
	if w == nil {
		return
	}

	for _, t := range w.Tests {
		if t.Key == msg.TestName {
			if t.Status != StatusFailed {
				t.Status = StatusPassed
				w.Completed++
				m.totalComplete++
			}
			return
		}
	}

	w.Tests = append(w.Tests, &TestNode{
		Key:      msg.TestName,
		Name:     msg.TestName,
		Status:   StatusPassed,
		WorkerID: msg.WorkerID,
	})
	w.Completed++
	m.totalComplete++
}

func (m *Model) handleTestFail(msg TestFailMsg) {
	w := m.workers[msg.WorkerID]
	if w == nil {
		return
	}

	firstError := len(m.errors) == 0

	for _, t := range w.Tests {
		if t.Key == msg.TestName {
			t.Status = StatusFailed
			t.ErrorMessage = msg.Message
			t.ErrorDetails = msg.Details
			w.Completed++
			w.Failed++
			m.totalComplete++
			m.totalFailed++
			m.errors = append(m.errors, ErrorEntry{
				TestName: t.Name,
				Message:  msg.Message,
				Details:  msg.Details,
				WorkerID: msg.WorkerID,
				Expanded: false,
			})
			if m.showCopyModal && firstError {
				m.copyModalErrors = true
			}
			return
		}
	}

	w.Tests = append(w.Tests, &TestNode{
		Key:          msg.TestName,
		Name:         msg.TestName,
		Status:       StatusFailed,
		ErrorMessage: msg.Message,
		ErrorDetails: msg.Details,
		WorkerID:     msg.WorkerID,
	})
	w.Completed++
	w.Failed++
	m.totalComplete++
	m.totalFailed++
	m.errors = append(m.errors, ErrorEntry{
		TestName: msg.TestName,
		Message:  msg.Message,
		Details:  msg.Details,
		WorkerID: msg.WorkerID,
		Expanded: false,
	})
	if m.showCopyModal && firstError {
		m.copyModalErrors = true
	}
}

func (m *Model) handleTestDeprecation(msg TestDeprecationMsg) {
	firstDeprecation := len(m.deprecations) == 0
	m.totalDeprecations++
	m.deprecations = append(m.deprecations, DeprecationEntry{
		TestName: msg.TestName,
		Message:  msg.Message,
		Details:  msg.Details,
		WorkerID: msg.WorkerID,
		Expanded: false,
	})
	if m.showCopyModal && firstDeprecation {
		m.copyModalDeprecations = true
	}
}

func (m *Model) handleTestSkip(msg TestSkipMsg) {
	w := m.workers[msg.WorkerID]
	if w == nil {
		return
	}

	for _, t := range w.Tests {
		if t.Key == msg.TestName {
			t.Status = StatusSkipped
			t.ErrorMessage = msg.Message
			w.Completed++
			m.totalComplete++
			m.totalSkipped++
			return
		}
	}

	w.Tests = append(w.Tests, &TestNode{
		Key:          msg.TestName,
		Name:         msg.TestName,
		Status:       StatusSkipped,
		ErrorMessage: msg.Message,
		WorkerID:     msg.WorkerID,
	})
	w.Completed++
	m.totalComplete++
	m.totalSkipped++
}

func (m *Model) handleWorkerComplete(msg WorkerCompleteMsg) {
	w := m.workers[msg.WorkerID]
	if w == nil {
		return
	}

	if msg.Error != nil {
		m.failedWorkers++
		w.Error = msg.Error
	}

	if !w.HasTestCount {
		m.testCount -= w.TestFiles
		w.Total = 0
		w.HasTestCount = true
	}
}

func (m *Model) handleTestCount(msg TestCountMsg) {
	w := m.workers[msg.WorkerID]
	if w == nil {
		return
	}

	if !w.HasTestCount {
		m.testCount = m.testCount - w.TestFiles + msg.Count
	} else {
		m.testCount = m.testCount - w.Total + msg.Count
	}
	w.Total = msg.Count
	w.HasTestCount = true
	m.hasTestCount = true
}

func (m *Model) getRunningTests() []*TestNode {
	var running []*TestNode
	for _, id := range m.workerOrder {
		w := m.workers[id]
		for _, t := range w.Tests {
			if t.Status == StatusRunning {
				running = append(running, t)
			}
		}
	}
	return running
}
