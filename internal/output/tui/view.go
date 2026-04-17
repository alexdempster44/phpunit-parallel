package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var styles = DefaultStyles()

func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	b.WriteString(m.renderOverallProgress())
	b.WriteString("\n\n")

	contentHeight := max(m.height-8, 8)

	leftWidth := max((m.width-5)/2, 20)
	rightWidth := max(m.width-leftWidth-5, 20)

	leftInnerWidth := leftWidth - 4

	maxWorkersHeight := contentHeight / 2
	maxDisplayWorkers := max((maxWorkersHeight-3)/2, 1)
	displayWorkers := min(m.workerCount, maxDisplayWorkers)
	workersHeight := (displayWorkers * 2) + 3
	if m.workerCount > maxDisplayWorkers {
		workersHeight++
	}
	workersHeight = max(workersHeight, 5)
	runningHeight := contentHeight - workersHeight - 1

	var topLeftPanel string
	if m.phase != PhaseRunning {
		topLeftPanel = m.renderSummaryPanel(leftInnerWidth, runningHeight-2)
	} else {
		topLeftPanel = m.renderRunningPanel(runningHeight, leftInnerWidth)
	}
	bottomLeftPanel := m.renderWorkersPanel(leftInnerWidth, workersHeight-2)

	topLeftStyle := styles.Panel.Width(leftWidth).Height(runningHeight)
	bottomLeftStyle := styles.Panel.Width(leftWidth).Height(workersHeight)

	if m.activePanel == PanelRunning && m.phase == PhaseRunning {
		topLeftStyle = styles.ActivePanel.Width(leftWidth).Height(runningHeight)
	}
	if m.activePanel == PanelWorkers {
		bottomLeftStyle = styles.ActivePanel.Width(leftWidth).Height(workersHeight)
	}

	topLeft := topLeftStyle.Render(topLeftPanel)
	bottomLeft := bottomLeftStyle.Render(bottomLeftPanel)
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, topLeft, bottomLeft)

	rightInnerWidth := rightWidth - 4
	errorsPanelHeight := contentHeight + 1
	errorsPanel := m.renderErrorsPanel(errorsPanelHeight, rightInnerWidth)
	errorsStyle := styles.Panel.Width(rightWidth).Height(errorsPanelHeight)
	if m.activePanel == PanelErrors {
		errorsStyle = styles.ActivePanel.Width(rightWidth).Height(errorsPanelHeight)
	}
	rightColumn := errorsStyle.Render(errorsPanel)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, " ", rightColumn))
	b.WriteString("\n")

	b.WriteString(m.renderHelpBar())

	output := b.String()

	if m.showCopyModal {
		output = m.overlayModal(output, m.renderCopyModal())
	}

	return output
}

func (m *Model) getElapsed() time.Duration {
	if !m.endTime.IsZero() {
		return m.endTime.Sub(m.startTime)
	}
	return time.Since(m.startTime)
}

func (m *Model) renderHeader() string {
	elapsed := m.getElapsed().Round(time.Second)

	var status string
	switch m.phase {
	case PhaseRunning:
		if m.retryAttempt > 0 {
			status = styles.TestRunning.Render(fmt.Sprintf("Retry #%d", m.retryAttempt))
		} else {
			status = styles.TestRunning.Render("Running")
		}
	case PhaseCleanup:
		status = styles.TestRunning.Render(fmt.Sprintf("Cleaning up workers... %d/%d", m.cleanupCompleted, m.cleanupTotal))
	case PhaseComplete, PhaseExploring:
		if m.hasFailed() {
			status = styles.TestFailed.Render("Complete - FAILED")
		} else if m.retryAttempt > 0 {
			status = styles.TestPassed.Render("Complete - PASSED (after retry)")
		} else {
			status = styles.TestPassed.Render("Complete - PASSED")
		}
	}

	titleText := "PHPUnit Parallel"
	if m.version != "" {
		titleText += " v" + m.version
	}
	title := styles.Title.Render(titleText)
	header := fmt.Sprintf("%s - %s (%s elapsed)", title, status, elapsed)

	if args := m.renderArgs(); args != "" {
		header += "  " + styles.Dim.Render(args)
	}

	return header
}

func (m *Model) renderArgs() string {
	var parts []string
	if m.filter != "" {
		parts = append(parts, "--filter "+m.filter)
	}
	if m.group != "" {
		parts = append(parts, "--group "+m.group)
	}
	if m.excludeGroup != "" {
		parts = append(parts, "--exclude-group "+m.excludeGroup)
	}
	return strings.Join(parts, " ")
}

func (m *Model) renderOverallProgress() string {
	total := m.testCount
	completed := m.totalComplete
	failed := m.totalFailed
	elapsed := m.getElapsed()

	var statsLine string
	if m.hasTestCount {
		percent := 100
		if total > 0 {
			percent = (completed * 100) / total
		}
		statsLine = fmt.Sprintf("Overall: %d/%d (%d%%)", completed, total, percent)
	} else {
		statsLine = fmt.Sprintf("Overall: %d test files", m.testCount)
	}

	if failed > 0 {
		statsLine += styles.TestFailed.Render(fmt.Sprintf(" %d failed", failed))
	}
	if m.failedWorkers > 0 {
		statsLine += styles.TestFailed.Render(fmt.Sprintf(" %d workers failed", m.failedWorkers))
	}
	if m.totalDeprecations > 0 {
		statsLine += styles.TestSkipped.Render(fmt.Sprintf(" %d deprecated", m.totalDeprecations))
	}

	var etaLine string
	if m.phase == PhaseRunning && m.hasTestCount && completed > 0 && total > 0 {
		estimatedTotal := time.Duration(float64(elapsed) * float64(total) / float64(completed))
		remaining := max(estimatedTotal-elapsed, 0)
		etaLine = styles.Dim.Render(fmt.Sprintf("  ETA: %s remaining (est. %s total)", formatDuration(remaining), formatDuration(estimatedTotal)))
	} else if m.phase != PhaseRunning {
		etaLine = styles.Dim.Render(fmt.Sprintf("  Completed in %s", formatDuration(elapsed)))
	}

	barWidth := max(m.width-2, 20)
	crashedFraction := 0
	if m.failedWorkers > 0 && m.workerCount > 0 {
		crashedFraction = (m.failedWorkers * total) / m.workerCount
		if crashedFraction == 0 {
			crashedFraction = 1
		}
	}
	bar := m.buildProgressBar(completed, failed, crashedFraction, total, barWidth, false)

	return statsLine + etaLine + "\n" + bar
}

func (m *Model) buildProgressBar(completed, failed, crashed, total, width int, dimmed bool) string {
	if total == 0 {
		if dimmed {
			return styles.Dim.Render("[" + strings.Repeat("█", width) + "]")
		}
		return styles.Dim.Render("[" + strings.Repeat("░", width) + "]")
	}

	crashedWidth := (crashed * width) / total
	if crashed > 0 && crashedWidth == 0 {
		crashedWidth = 1
	}

	usableWidth := width - crashedWidth
	filledWidth := (completed * usableWidth) / total
	if completed >= total {
		filledWidth = usableWidth
	}

	failedWidth := 0
	if completed > 0 {
		failedWidth = (failed * filledWidth) / completed
	}
	if failed > 0 && failedWidth == 0 && filledWidth > 0 {
		failedWidth = 1
	}
	passedWidth := filledWidth - failedWidth
	remaining := max(usableWidth-filledWidth, 0)

	if dimmed {
		return styles.Dim.Render("["+strings.Repeat("█", passedWidth)) +
			styles.TestFailed.Render(strings.Repeat("█", failedWidth)) +
			styles.Dim.Render(strings.Repeat("░", remaining)) +
			styles.WorkerCrashed.Render(crashedBar(crashedWidth)) +
			styles.Dim.Render("]")
	}

	return styles.Dim.Render("[") +
		styles.TestPassed.Render(strings.Repeat("█", passedWidth)) +
		styles.TestFailed.Render(strings.Repeat("█", failedWidth)) +
		styles.Dim.Render(strings.Repeat("░", remaining)) +
		styles.WorkerCrashed.Render(crashedBar(crashedWidth)) +
		styles.Dim.Render("]")
}

func crashedBar(width int) string {
	b := make([]byte, 0, width*3)
	for i := range width {
		if i%2 == 0 {
			b = append(b, "▀"...)
		} else {
			b = append(b, "▄"...)
		}
	}
	return string(b)
}

func (m *Model) renderWorkersPanel(panelWidth int, panelHeight int) string {
	var lines []string
	lines = append(lines, styles.Bold.Render("Workers"))
	lines = append(lines, "")

	barWidth := max(panelWidth-2, 10)

	var erroredWorkers []int
	var activeWorkers []int
	var finishedWorkers []int
	for _, id := range m.workerOrder {
		w := m.workers[id]
		if w.Error != nil {
			erroredWorkers = append(erroredWorkers, id)
		} else if w.HasTestCount && w.Completed >= w.Total {
			finishedWorkers = append(finishedWorkers, id)
		} else {
			activeWorkers = append(activeWorkers, id)
		}
	}
	sortedWorkers := make([]int, 0, len(m.workerOrder))
	sortedWorkers = append(sortedWorkers, erroredWorkers...)
	sortedWorkers = append(sortedWorkers, activeWorkers...)
	sortedWorkers = append(sortedWorkers, finishedWorkers...)

	var workerLines []string
	for _, id := range sortedWorkers {
		w := m.workers[id]
		isComplete := w.HasTestCount && w.Completed >= w.Total

		var statsLine string
		var workerBar string

		if w.Error != nil {
			statsLine = styles.TestFailed.Render(fmt.Sprintf("Worker %d: %s", id+1, w.Error))
			workerBar = styles.Dim.Render("[") +
				styles.WorkerCrashed.Render(crashedBar(barWidth)) +
				styles.Dim.Render("]")
		} else if w.HasTestCount {
			percent := 100
			if w.Total > 0 {
				percent = (w.Completed * 100) / w.Total
			}
			baseLine := fmt.Sprintf("Worker %d: %d/%d (%d%%)", id+1, w.Completed, w.Total, percent)
			if isComplete {
				baseLine = styles.Dim.Render(baseLine)
			}
			statsLine = baseLine
			if w.Failed > 0 {
				statsLine += styles.TestFailed.Render(fmt.Sprintf(" %d failed", w.Failed))
			}
			workerBar = m.buildProgressBar(w.Completed, w.Failed, 0, w.Total, barWidth, isComplete)
		} else {
			statsLine = fmt.Sprintf("Worker %d: %d files", id+1, w.TestFiles)
			if isComplete {
				statsLine = styles.Dim.Render(statsLine)
			}
			workerBar = m.buildProgressBar(0, 0, 0, 0, barWidth, isComplete)
		}

		workerLines = append(workerLines, statsLine)
		workerLines = append(workerLines, workerBar)
	}

	visibleLines := max(panelHeight, 2)

	workersPerPage := max(visibleLines/2, 1)
	totalWorkers := len(m.workerOrder)
	totalPages := max((totalWorkers+workersPerPage-1)/workersPerPage, 1)

	needsPagination := totalPages > 1
	if needsPagination {
		visibleLines--
		workersPerPage = max(visibleLines/2, 1)
		totalPages = (totalWorkers + workersPerPage - 1) / workersPerPage
	}

	m.workersOffset = max(m.workersOffset, 0)
	m.workersOffset = min(m.workersOffset, totalPages-1)
	currentPage := m.workersOffset

	startWorker := currentPage * workersPerPage
	endWorker := min(startWorker+workersPerPage, totalWorkers)

	startLine := startWorker * 2
	endLine := min(endWorker*2, len(workerLines))
	lines = append(lines, workerLines[startLine:endLine]...)

	if needsPagination {
		pageInfo := styles.Dim.Render(fmt.Sprintf("Page %d/%d (↑↓)", currentPage+1, totalPages))
		lines = append(lines, pageInfo)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderSummaryPanel(panelWidth int, _ int) string {
	var lines []string

	elapsed := m.getElapsed()
	elapsedSec := elapsed.Seconds()
	var testsPerSec float64
	if elapsedSec > 0 {
		testsPerSec = float64(m.totalComplete) / elapsedSec
	}

	cumulativeTime := elapsed * time.Duration(m.workerCount)

	var resultText string
	if m.hasFailed() {
		resultText = styles.TestFailed.Render("  FAILED  ")
	} else {
		resultText = styles.TestPassed.Render("  PASSED  ")
	}
	resultPadding := max((panelWidth-visibleLength(resultText))/2, 0)

	lines = append(lines, "")
	lines = append(lines, strings.Repeat(" ", resultPadding)+resultText)
	lines = append(lines, "")

	formatRow := func(label, value string, valueStyle ...lipgloss.Style) string {
		styledValue := value
		if len(valueStyle) > 0 {
			styledValue = valueStyle[0].Render(value)
		}
		valueLen := len(value)
		spacing := max(panelWidth-len(label)-valueLen, 1)
		return label + strings.Repeat(" ", spacing) + styledValue
	}

	lines = append(lines, formatRow("Duration:", formatDuration(elapsed)))
	lines = append(lines, formatRow("Cumulative:", formatDuration(cumulativeTime)))
	lines = append(lines, formatRow("Rate:", fmt.Sprintf("%.1f tests/sec", testsPerSec)))
	lines = append(lines, "")

	passed := m.totalComplete - m.totalFailed - m.totalSkipped
	lines = append(lines, formatRow("Total:", fmt.Sprintf("%d tests", m.totalComplete)))
	lines = append(lines, formatRow("Passed:", fmt.Sprintf("%d", passed), styles.TestPassed))

	if m.totalFailed > 0 {
		lines = append(lines, formatRow("Failed:", fmt.Sprintf("%d", m.totalFailed), styles.TestFailed))
	}
	if m.totalSkipped > 0 {
		lines = append(lines, formatRow("Skipped:", fmt.Sprintf("%d", m.totalSkipped), styles.TestSkipped))
	}

	lines = append(lines, "")
	lines = append(lines, formatRow("Deprecations:", fmt.Sprintf("%d", m.totalDeprecations), styles.TestSkipped))

	lines = append(lines, "")
	workerValue := fmt.Sprintf("%d", m.workerCount)
	if m.failedWorkers > 0 {
		plainLen := len(fmt.Sprintf("%d (%d failed)", m.workerCount, m.failedWorkers))
		failedPart := styles.TestFailed.Render(fmt.Sprintf(" (%d failed)", m.failedWorkers))
		styledValue := styles.Dim.Render(fmt.Sprintf("%d", m.workerCount)) + failedPart
		spacing := max(panelWidth-len("Workers:")-plainLen, 1)
		lines = append(lines, "Workers:"+strings.Repeat(" ", spacing)+styledValue)
	} else {
		lines = append(lines, formatRow("Workers:", workerValue, styles.Dim))
	}
	if m.retryAttempt > 0 {
		lines = append(lines, formatRow("Retry:", fmt.Sprintf("#%d", m.retryAttempt), styles.Dim))
	}

	return strings.Join(lines, "\n")
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)
	return fmt.Sprintf("%dm %.1fs", minutes, seconds)
}

func (m *Model) renderRunningPanel(height int, panelWidth int) string {
	var lines []string

	runningTests := m.getRunningTests()
	title := fmt.Sprintf("Running (%d)", len(runningTests))
	lines = append(lines, styles.Bold.Render(title))
	lines = append(lines, "")

	if len(runningTests) == 0 {
		if m.phase == PhaseRunning {
			lines = append(lines, styles.Dim.Render("Waiting..."))
		} else {
			lines = append(lines, styles.Dim.Render("Complete"))
		}
		return strings.Join(lines, "\n")
	}

	maxNameLen := max(panelWidth-4, 10)

	for i, t := range runningTests {
		icon := styles.IconRunning
		line := fmt.Sprintf("%s %s", styles.TestRunning.Render(icon), truncateName(t.Name, maxNameLen))

		if m.activePanel == PanelRunning && i == m.runningCursor {
			line = styles.Cursor.Render(line)
		}
		lines = append(lines, line)
	}

	visibleLines := max(height-3, 1)
	if len(lines) > visibleLines+2 {
		start := m.runningOffset
		if m.runningCursor+2 < start {
			start = m.runningCursor
		}
		if m.runningCursor >= start+visibleLines {
			start = m.runningCursor - visibleLines + 1
		}
		start = max(start, 0)
		headerLines := lines[:2]
		contentLines := lines[2:]
		if start+visibleLines > len(contentLines) {
			start = max(len(contentLines)-visibleLines, 0)
		}
		m.runningOffset = start
		lines = append(headerLines, contentLines[start:start+min(visibleLines, len(contentLines))]...)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderErrorsPanel(height int, panelWidth int) string {
	var lines []string

	title := fmt.Sprintf("Errors (%d)", len(m.errors))
	if m.totalDeprecations > 0 {
		title += fmt.Sprintf(" | Deprecations (%d)", len(m.deprecations))
	}
	lines = append(lines, styles.Bold.Render(title))
	lines = append(lines, "")

	if len(m.errors) == 0 && len(m.deprecations) == 0 {
		lines = append(lines, styles.Dim.Render("No errors"))
		return strings.Join(lines, "\n")
	}

	maxNameLen := max(panelWidth-4, 10)
	cursorStart := 0
	cursorEnd := 0

	// Render error entries
	for i, e := range m.errors {
		expandIcon := styles.IconCollaps
		if e.Expanded {
			expandIcon = styles.IconExpand
		}

		if i == m.errorCursor {
			cursorStart = len(lines) - 2
		}

		line := fmt.Sprintf("%s %s", expandIcon, styles.TestFailed.Render(truncateName(e.TestName, maxNameLen)))
		if m.activePanel == PanelErrors && i == m.errorCursor {
			line = styles.Cursor.Render(line)
		}
		lines = append(lines, line)

		if e.Expanded {
			detailWidth := max(panelWidth-4, 10)
			if e.Message != "" {
				msgLines := wrapText(e.Message, detailWidth)
				for _, ml := range msgLines {
					lines = append(lines, "  "+styles.ErrorMsg.Render(ml))
				}
			}
			if e.Details != "" {
				detailLines := strings.Split(e.Details, "\n")
				for _, d := range detailLines {
					if d != "" {
						lines = append(lines, "  "+styles.ErrorDetail.Render(truncateName(d, detailWidth)))
					}
				}
			}
		}

		if i == m.errorCursor {
			cursorEnd = len(lines) - 2
		}
	}

	// Render deprecation entries
	for i, d := range m.deprecations {
		globalIdx := len(m.errors) + i
		expandIcon := styles.IconCollaps
		if d.Expanded {
			expandIcon = styles.IconExpand
		}

		if globalIdx == m.errorCursor {
			cursorStart = len(lines) - 2
		}

		line := fmt.Sprintf("%s %s", expandIcon, styles.TestSkipped.Render(truncateName(d.TestName, maxNameLen)))
		if m.activePanel == PanelErrors && globalIdx == m.errorCursor {
			line = styles.Cursor.Render(line)
		}
		lines = append(lines, line)

		if d.Expanded {
			detailWidth := max(panelWidth-4, 10)
			if d.Message != "" {
				msgLines := wrapText(d.Message, detailWidth)
				for _, ml := range msgLines {
					lines = append(lines, "  "+styles.ErrorMsg.Render(ml))
				}
			}
			if d.Details != "" {
				detailLines := strings.Split(d.Details, "\n")
				for _, dl := range detailLines {
					if dl != "" {
						lines = append(lines, "  "+styles.ErrorDetail.Render(truncateName(dl, detailWidth)))
					}
				}
			}
		}

		if globalIdx == m.errorCursor {
			cursorEnd = len(lines) - 2
		}
	}

	visibleLines := max(height-2, 1)
	if len(lines) > visibleLines+2 {
		start := m.errorOffset
		// Ensure cursor title line is visible (scroll up if needed)
		if cursorStart < start {
			start = cursorStart
		}
		// Ensure cursor end is visible (scroll down if needed)
		if cursorEnd >= start+visibleLines {
			start = cursorEnd - visibleLines + 1
		}
		// But always keep the title line visible even if expanded content is tall
		if cursorStart < start {
			start = cursorStart
		}
		start = max(start, 0)
		headerLines := lines[:2]
		contentLines := lines[2:]
		if start+visibleLines > len(contentLines) {
			start = max(len(contentLines)-visibleLines, 0)
		}
		m.errorOffset = start
		lines = append(headerLines, contentLines[start:start+min(visibleLines, len(contentLines))]...)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderHelpBar() string {
	if m.copyNotice != "" {
		return padToWidth(styles.TestPassed.Render(m.copyNotice), m.width)
	}

	var segments []string
	if m.phase == PhaseRunning {
		segments = []string{"[Tab] Panel", "[↑↓] Navigate", "[←→] Expand", "[c] Copy", "[C] Copy all", "[Ctrl+C] Quit"}
	} else if m.phase == PhaseComplete && m.totalFailed > 0 {
		segments = []string{"[Tab] Panel", "[↑↓] Navigate", "[←→] Expand", "[c] Copy", "[C] Copy all", "[r] Rerun failed", "[a] Rerun all", "[q] Quit"}
	} else {
		segments = []string{"[Tab] Panel", "[↑↓] Navigate", "[←→] Expand", "[c] Copy", "[C] Copy all", "[a] Rerun all", "[q] Quit"}
	}

	help := fitSegments(segments, m.width)
	return padToWidth(styles.HelpBar.Render(help), m.width)
}

// fitSegments joins help segments separated by two spaces, dropping segments
// from the right until the line fits within width.
func fitSegments(segments []string, width int) string {
	if width <= 0 {
		return ""
	}
	for n := len(segments); n > 0; n-- {
		joined := strings.Join(segments[:n], "  ")
		if visibleLength(joined) <= width {
			return joined
		}
	}
	first := segments[0]
	if visibleLength(first) <= width {
		return first
	}
	runes := []rune(first)
	if width >= len(runes) {
		return first
	}
	return string(runes[:width])
}

// padToWidth pads a styled string with trailing spaces so the visible width
// matches the target. This prevents leftover characters from a prior, longer
// render from leaking through bubbletea's diff-based renderer (e.g. when the
// help bar is replaced by a shorter copy notice).
func padToWidth(s string, width int) string {
	visible := visibleLength(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

func (m *Model) renderCopyModal() string {
	modalWidth := 30
	var lines []string

	lines = append(lines, styles.Bold.Render("Copy to clipboard"))
	lines = append(lines, "")

	lines = append(lines, styles.Dim.Render("Include:"))

	index := 0

	if len(m.errors) > 0 {
		check := "[ ]"
		if m.copyModalErrors {
			check = "[x]"
		}
		label := fmt.Sprintf("%s Errors (%d)", check, len(m.errors))
		if m.copyModalCursor == index {
			lines = append(lines, styles.Cursor.Render("  > "+label))
		} else {
			lines = append(lines, "    "+label)
		}
		index++
	}

	if len(m.deprecations) > 0 {
		check := "[ ]"
		if m.copyModalDeprecations {
			check = "[x]"
		}
		label := fmt.Sprintf("%s Deprecations (%d)", check, len(m.deprecations))
		if m.copyModalCursor == index {
			lines = append(lines, styles.Cursor.Render("  > "+label))
		} else {
			lines = append(lines, "    "+label)
		}
		index++
	}

	lines = append(lines, "")

	namesLabel := "Copy names"
	if m.copyModalCursor == index {
		lines = append(lines, styles.Cursor.Render("  > "+namesLabel))
	} else {
		lines = append(lines, "    "+namesLabel)
	}
	index++

	detailsLabel := "Copy details"
	if m.copyModalCursor == index {
		lines = append(lines, styles.Cursor.Render("  > "+detailsLabel))
	} else {
		lines = append(lines, "    "+detailsLabel)
	}

	lines = append(lines, "")
	lines = append(lines, styles.Dim.Render("Esc to cancel"))

	content := strings.Join(lines, "\n")

	return styles.ActivePanel.
		Width(modalWidth).
		Render(content)
}

func (m *Model) overlayModal(background, modal string) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal, "\n")

	modalHeight := len(modalLines)
	modalWidth := 0
	for _, line := range modalLines {
		if w := visibleLength(line); w > modalWidth {
			modalWidth = w
		}
	}

	startRow := max((m.height-modalHeight)/2, 0)
	startCol := max((m.width-modalWidth)/2, 0)

	// Pad background to fill height if needed
	for len(bgLines) < m.height {
		bgLines = append(bgLines, "")
	}

	for i, modalLine := range modalLines {
		row := startRow + i
		if row >= len(bgLines) {
			break
		}
		bgRunes := []rune(expandAnsiToRunes(bgLines[row], m.width))
		modalRunes := []rune(modalLine)

		// Build: background prefix + modal line + background suffix
		prefix := string(bgRunes[:min(startCol, len(bgRunes))])
		suffixStart := startCol + visibleLength(modalLine)
		var suffix string
		if suffixStart < len(bgRunes) {
			suffix = string(bgRunes[suffixStart:])
		}

		_ = modalRunes
		bgLines[row] = prefix + modalLine + suffix
	}

	return strings.Join(bgLines, "\n")
}

func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	if maxLen <= 3 {
		return name[:maxLen]
	}
	return name[:maxLen-3] + "..."
}

func wrapText(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var lines []string
	for len(text) > maxLen {
		lines = append(lines, text[:maxLen])
		text = text[maxLen:]
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}

// expandAnsiToRunes returns a slice of runes of exactly targetWidth visible characters,
// where each rune position maps to one visible column. ANSI sequences are stripped.
func expandAnsiToRunes(s string, targetWidth int) string {
	var result []rune
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result = append(result, r)
	}
	for len(result) < targetWidth {
		result = append(result, ' ')
	}
	return string(result)
}

func visibleLength(s string) int {
	length := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		length++
	}
	return length
}
