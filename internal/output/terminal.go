package output

import (
	"fmt"
	"strings"
	"sync"
)

type terminalWorkerState struct {
	testFileCount  int
	testCount      int
	hasTestCount   bool
	testsCompleted int
	testsFailed    int
	testsSkipped   int
	completed      bool
	err            error
	failedTests    map[string]bool
}

type TerminalOutput struct {
	mu       sync.Mutex
	workers  map[int]*terminalWorkerState
	onCancel func()
}

func NewTerminalOutput() *TerminalOutput {
	return &TerminalOutput{
		workers: make(map[int]*terminalWorkerState),
	}
}

func (t *TerminalOutput) Start(opts StartOptions) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if opts.Version != "" {
		fmt.Printf("[start] PHPUnit Parallel v%s\n", opts.Version)
	}
	fmt.Printf("[start] Running %d test files across %d workers\n", opts.TestCount, opts.WorkerCount)
	if opts.Filter != "" {
		fmt.Printf("[start] Filter: %s\n", opts.Filter)
	}
	if opts.Group != "" {
		fmt.Printf("[start] Group: %s\n", opts.Group)
	}
	if opts.ExcludeGroup != "" {
		fmt.Printf("[start] Exclude group: %s\n", opts.ExcludeGroup)
	}
}

func (t *TerminalOutput) WorkerStart(workerID, testCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.workers[workerID] = &terminalWorkerState{
		testFileCount: testCount,
		failedTests:   make(map[string]bool),
	}
	fmt.Printf("[worker %d] Started with %d test files\n", workerID, testCount)
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
		if count := ParseTeamCityCount(line); count != nil {
			w.testCount = *count
			w.hasTestCount = true
			fmt.Printf("[worker %d] Test count: %d\n", workerID, *count)
		}

	case strings.HasPrefix(line, "##teamcity[testFailed "):
		name, message, details := ParseTeamCityError(line)
		w.testsFailed++
		w.testsCompleted++
		w.failedTests[name] = true
		fmt.Printf("[fail] %s\n", name)
		if message != "" {
			fmt.Printf("[fail]   Message: %s\n", message)
		}
		if details != "" {
			fmt.Printf("[fail]   Details:\n")
			for _, dl := range strings.Split(details, "\n") {
				if dl != "" {
					fmt.Printf("[fail]     %s\n", dl)
				}
			}
		}

	case strings.HasPrefix(line, "##teamcity[testIgnored "):
		w.testsSkipped++
		w.testsCompleted++

	case strings.HasPrefix(line, "##teamcity[testFinished "):
		name := ParseTeamCityAttr(line, "name")
		if !w.failedTests[name] {
			w.testsCompleted++
		}
	}
}

func (t *TerminalOutput) WorkerComplete(workerID int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	w := t.workers[workerID]
	if w == nil {
		return
	}

	w.completed = true
	w.err = err

	msg := fmt.Sprintf("[worker %d] Completed: %d tests, %d failed", workerID, w.testsCompleted, w.testsFailed)
	if err != nil {
		msg += fmt.Sprintf(" (error: %s)", err)
	}
	fmt.Println(msg)
}

func (t *TerminalOutput) CleanupProgress(completed, total int) {
	fmt.Printf("[cleanup] %d/%d workers cleaned up\n", completed, total)
}

func (t *TerminalOutput) Finish() {
	t.mu.Lock()
	defer t.mu.Unlock()

	totalTests := 0
	totalFailed := 0
	totalSkipped := 0
	for _, w := range t.workers {
		totalTests += w.testsCompleted
		totalFailed += w.testsFailed
		totalSkipped += w.testsSkipped
	}

	totalPassed := totalTests - totalFailed - totalSkipped
	fmt.Printf("[summary] Total: %d tests, %d passed, %d failed, %d skipped\n", totalTests, totalPassed, totalFailed, totalSkipped)

	if totalFailed > 0 {
		fmt.Println("[result] FAILED")
	} else {
		fmt.Println("[result] OK")
	}
}

func (t *TerminalOutput) SetOnCancel(fn func()) {
	t.onCancel = fn
}

func (t *TerminalOutput) AwaitRetry() RetryAction {
	return ActionQuit
}

func (t *TerminalOutput) RetryStart(_ RetryStartOptions) {}
