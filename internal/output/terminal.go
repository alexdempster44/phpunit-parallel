package output

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim     = "\033[2m"
	colorBold    = "\033[1m"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type terminalWorker struct {
	testCount       int
	testsCompleted  int
	testsFailed     int
	completed       bool
	err             error
	failedTestNames map[string]bool
}

type terminalError struct {
	testName string
	message  string
	details  string
}

type TerminalOutput struct {
	mu            sync.Mutex
	testCount     int
	workerCount   int
	workers       map[int]*terminalWorker
	errors        []terminalError
	spinnerFrame  int
	renderedLines int
	done          chan struct{}
}

func NewTerminalOutput() *TerminalOutput {
	return &TerminalOutput{
		workers: make(map[int]*terminalWorker),
		done:    make(chan struct{}),
	}
}

func (t *TerminalOutput) Start(testCount, workerCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.testCount = testCount
	t.workerCount = workerCount

	fmt.Printf("Running %d tests across %d workers\n\n", testCount, workerCount)

	go t.runSpinner()
}

func (t *TerminalOutput) runSpinner() {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-t.done:
			return
		case <-ticker.C:
			t.mu.Lock()
			t.spinnerFrame++
			t.render()
			t.mu.Unlock()
		}
	}
}

func (t *TerminalOutput) WorkerStart(workerID, testCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.workers[workerID] = &terminalWorker{
		testCount:       testCount,
		failedTestNames: make(map[string]bool),
	}
	t.render()
}

func (t *TerminalOutput) WorkerLine(workerID int, line string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	w := t.workers[workerID]
	if w == nil {
		return
	}

	switch {
	case strings.HasPrefix(line, "##teamcity[testCount "):
		if count := parseTeamCityCount(line); count != nil {
			t.testCount = t.testCount - w.testCount + *count
			w.testCount = *count
		}

	case strings.HasPrefix(line, "##teamcity[testFailed "):
		w.testsFailed++
		w.testsCompleted++
		name, message, details := parseTeamCityError(line)
		w.failedTestNames[name] = true
		t.errors = append(t.errors, terminalError{testName: name, message: message, details: details})

	case strings.HasPrefix(line, "##teamcity[testFinished "):
		name := parseTeamCityAttr(line, "name")
		if !w.failedTestNames[name] {
			w.testsCompleted++
		}
	}
}

func (t *TerminalOutput) WorkerComplete(workerID int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if w, ok := t.workers[workerID]; ok {
		w.completed = true
		w.err = err
	}

	t.render()
}

func (t *TerminalOutput) Finish() {
	close(t.done)

	t.mu.Lock()
	defer t.mu.Unlock()

	totalTests := 0
	totalFailed := 0
	for _, w := range t.workers {
		totalTests += w.testsCompleted
		totalFailed += w.testsFailed
	}

	t.testCount = totalTests

	t.render()

	fmt.Println()
	if totalFailed > 0 {
		fmt.Printf("%s%sFAILED:%s %d tests, %s%d failed%s\n",
			colorBold, colorRed, colorReset, totalTests, colorRed, totalFailed, colorReset)
	} else {
		fmt.Printf("%s%sOK:%s %d tests passed\n", colorBold, colorGreen, colorReset, totalTests)
	}
}

func (t *TerminalOutput) render() {
	t.clearLines()

	totalCompleted := 0
	totalFailed := 0
	for _, w := range t.workers {
		totalCompleted += w.testsCompleted
		totalFailed += w.testsFailed
	}

	progressBar := t.buildProgressBar(totalCompleted, totalFailed, t.testCount, 30)
	progressText := fmt.Sprintf("%s%d/%d tests%s", colorBold, totalCompleted, t.testCount, colorReset)
	if totalFailed > 0 {
		progressText += fmt.Sprintf(" %s(%d failed)%s", colorRed, totalFailed, colorReset)
	}
	fmt.Printf("  %s %s\n\n", progressBar, progressText)

	var lines []string
	for i := 0; i < t.workerCount; i++ {
		w, ok := t.workers[i]
		if !ok {
			lines = append(lines, fmt.Sprintf("  %sWorker %d:%s %spending%s", colorDim, i, colorReset, colorYellow, colorReset))
			continue
		}

		var status string
		if w.completed {
			if w.testsFailed > 0 {
				status = fmt.Sprintf("%s✗ failed%s %s(%d tests, %s%d failed%s)%s", colorRed, colorReset, colorDim, w.testsCompleted, colorRed, w.testsFailed, colorDim, colorReset)
			} else {
				status = fmt.Sprintf("%s✓ passed%s %s(%d tests)%s", colorGreen, colorReset, colorDim, w.testsCompleted, colorReset)
			}
		} else {
			spinner := spinnerFrames[t.spinnerFrame%len(spinnerFrames)]
			workerBar := t.buildProgressBar(w.testsCompleted, w.testsFailed, w.testCount, 15)
			countText := fmt.Sprintf("%s%d/%d%s", colorDim, w.testsCompleted, w.testCount, colorReset)
			if w.testsFailed > 0 {
				countText += fmt.Sprintf(" %s(%d failed)%s", colorRed, w.testsFailed, colorReset)
			}
			status = fmt.Sprintf("%s%s%s %s %s", colorCyan, spinner, colorReset, workerBar, countText)
		}

		lines = append(lines, fmt.Sprintf("  Worker %d: %s", i, status))
	}

	for _, line := range lines {
		fmt.Println(line)
	}
	lineCount := 2 + len(lines)

	if len(t.errors) > 0 {
		fmt.Println()
		lineCount++
		for i, e := range t.errors {
			fmt.Printf("  %s%d) %s%s\n", colorRed, i+1, e.testName, colorReset)
			lineCount++
			if e.message != "" {
				fmt.Printf("     %s%s%s\n", colorYellow, e.message, colorReset)
				lineCount++
			}
			if e.details != "" {
				detailLines := strings.Split(e.details, "\n")
				for _, detail := range detailLines {
					if detail != "" {
						fmt.Printf("     %s%s%s\n", colorDim, detail, colorReset)
						lineCount++
					}
				}
			}
		}
	}

	t.spinnerFrame++
	t.renderedLines = lineCount
}

func (t *TerminalOutput) buildProgressBar(completed, failed, total, width int) string {
	if total == 0 {
		return colorDim + "[" + strings.Repeat("░", width) + "]" + colorReset
	}

	filledWidth := min((completed*width)/total, width)
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

	return colorDim + "[" + colorReset +
		colorGreen + strings.Repeat("█", passedWidth) + colorReset +
		colorRed + strings.Repeat("█", failedWidth) + colorReset +
		colorDim + strings.Repeat("░", remaining) + "]" + colorReset
}

func (t *TerminalOutput) clearLines() {
	if t.renderedLines > 0 {
		fmt.Print(strings.Repeat("\033[A\033[K", t.renderedLines))
	}
}
