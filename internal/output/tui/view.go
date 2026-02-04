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

	return b.String()
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
		status = styles.TestRunning.Render("Running")
	case PhaseComplete, PhaseExploring:
		if m.totalFailed > 0 {
			status = styles.TestFailed.Render("Complete - FAILED")
		} else {
			status = styles.TestPassed.Render("Complete - PASSED")
		}
	}

	title := styles.Title.Render("PHPUnit Parallel")
	return fmt.Sprintf("%s - %s (%s elapsed)", title, status, elapsed)
}

func (m *Model) renderOverallProgress() string {
	total := m.testCount
	completed := m.totalComplete
	failed := m.totalFailed
	elapsed := m.getElapsed()

	var statsLine string
	if m.hasTestCount {
		percent := 0
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

	var etaLine string
	if m.phase == PhaseRunning && m.hasTestCount && completed > 0 && total > 0 {
		estimatedTotal := time.Duration(float64(elapsed) * float64(total) / float64(completed))
		remaining := max(estimatedTotal-elapsed, 0)
		etaLine = styles.Dim.Render(fmt.Sprintf("  ETA: %s remaining (est. %s total)", formatDuration(remaining), formatDuration(estimatedTotal)))
	} else if m.phase != PhaseRunning {
		etaLine = styles.Dim.Render(fmt.Sprintf("  Completed in %s", formatDuration(elapsed)))
	}

	barWidth := max(m.width-2, 20)
	bar := m.buildProgressBar(completed, failed, total, barWidth, false)

	return statsLine + etaLine + "\n" + bar
}

func (m *Model) buildProgressBar(completed, failed, total, width int, dimmed bool) string {
	if total == 0 {
		return styles.Dim.Render("[" + strings.Repeat("░", width) + "]")
	}

	filledWidth := (completed * width) / total
	if completed >= total {
		filledWidth = width
	}

	failedWidth := 0
	if completed > 0 {
		failedWidth = (failed * filledWidth) / completed
	}
	if failed > 0 && failedWidth == 0 && filledWidth > 0 {
		failedWidth = 1
	}
	passedWidth := filledWidth - failedWidth
	remaining := width - filledWidth

	if dimmed {
		return styles.Dim.Render("["+strings.Repeat("█", passedWidth)) +
			styles.TestFailed.Render(strings.Repeat("█", failedWidth)) +
			styles.Dim.Render(strings.Repeat("░", remaining)+"]")
	}

	return styles.Dim.Render("[") +
		styles.TestPassed.Render(strings.Repeat("█", passedWidth)) +
		styles.TestFailed.Render(strings.Repeat("█", failedWidth)) +
		styles.Dim.Render(strings.Repeat("░", remaining)+"]")
}

func (m *Model) renderWorkersPanel(panelWidth int, panelHeight int) string {
	var lines []string
	lines = append(lines, styles.Bold.Render("Workers"))
	lines = append(lines, "")

	barWidth := max(panelWidth-2, 10)

	sortedWorkers := make([]int, 0, len(m.workerOrder))
	var finishedWorkers []int
	for _, id := range m.workerOrder {
		w := m.workers[id]
		if w.HasTestCount && w.Completed >= w.Total {
			finishedWorkers = append(finishedWorkers, id)
		} else {
			sortedWorkers = append(sortedWorkers, id)
		}
	}
	sortedWorkers = append(sortedWorkers, finishedWorkers...)

	var workerLines []string
	for _, id := range sortedWorkers {
		w := m.workers[id]
		isComplete := w.HasTestCount && w.Completed >= w.Total

		var statsLine string
		if w.HasTestCount {
			percent := 0
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
		} else {
			statsLine = fmt.Sprintf("Worker %d: %d files", id+1, w.TestFiles)
			if isComplete {
				statsLine = styles.Dim.Render(statsLine)
			}
		}
		workerLines = append(workerLines, statsLine)

		var workerBar string
		if w.HasTestCount {
			workerBar = m.buildProgressBar(w.Completed, w.Failed, w.Total, barWidth, isComplete)
		} else {
			workerBar = m.buildProgressBar(0, 0, 0, barWidth, isComplete)
		}
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
	if m.totalFailed > 0 {
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
	lines = append(lines, formatRow("Workers:", fmt.Sprintf("%d", m.workerCount), styles.Dim))

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
	lines = append(lines, styles.Bold.Render(title))
	lines = append(lines, "")

	if len(m.errors) == 0 {
		lines = append(lines, styles.Dim.Render("No errors"))
		return strings.Join(lines, "\n")
	}

	maxNameLen := max(panelWidth-4, 10)

	for i, e := range m.errors {
		expandIcon := styles.IconCollaps
		if e.Expanded {
			expandIcon = styles.IconExpand
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
	}

	visibleLines := max(height-2, 1)
	if len(lines) > visibleLines+2 {
		start := m.errorOffset
		if m.errorCursor+2 < start {
			start = m.errorCursor
		}
		if m.errorCursor >= start+visibleLines {
			start = m.errorCursor - visibleLines + 1
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
	var help string
	if m.phase == PhaseRunning {
		help = "[Tab] Panel  [↑↓] Navigate  [Enter] Expand  [Ctrl+C] Quit"
	} else {
		help = "[Tab] Panel  [↑↓] Navigate  [Enter] Expand  [q] Quit"
	}
	return styles.HelpBar.Render(help)
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
