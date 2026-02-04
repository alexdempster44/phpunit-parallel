package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
		if m.phase == PhaseRunning {
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

	case TestSkipMsg:
		m.handleTestSkip(msg)
		return m, nil

	case TestCountMsg:
		m.handleTestCount(msg)
		return m, nil

	case FinishMsg:
		m.phase = PhaseComplete
		m.endTime = time.Now()
		m.activePanel = PanelErrors
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := DefaultKeyMap()

	switch {
	case key.Matches(msg, keys.Quit):
		if m.phase == PhaseComplete || m.phase == PhaseExploring {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.String() == "ctrl+c" {
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

	case key.Matches(msg, keys.PageUp):
		m.moveCursor(-10)
		return m, nil

	case key.Matches(msg, keys.PageDown):
		m.moveCursor(10)
		return m, nil

	}

	return m, nil
}

func (m *Model) moveCursor(delta int) {
	switch m.activePanel {
	case PanelWorkers:
		m.workersOffset += delta

	case PanelErrors:
		maxCursor := len(m.errors) - 1
		m.errorCursor += delta
		if m.errorCursor < 0 {
			m.errorCursor = 0
		}
		if m.errorCursor > maxCursor && maxCursor >= 0 {
			m.errorCursor = maxCursor
		}
	}
}

func (m *Model) toggle() {
	if m.activePanel == PanelErrors {
		if m.errorCursor >= 0 && m.errorCursor < len(m.errors) {
			m.errors[m.errorCursor].Expanded = !m.errors[m.errorCursor].Expanded
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
