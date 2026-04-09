package runner

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/alexdempster44/phpunit-parallel/internal/output"
)

// FailureTracker intercepts WorkerLine calls to track failed test names
// from TeamCity protocol output.
type FailureTracker struct {
	mu            sync.Mutex
	failedTests   map[string]bool
	failedWorkers map[int]bool
}

func NewFailureTracker() *FailureTracker {
	return &FailureTracker{
		failedTests:   make(map[string]bool),
		failedWorkers: make(map[int]bool),
	}
}

// ProcessLine parses a worker output line looking for testFailed messages.
// It extracts the test name and strips data set suffixes.
func (ft *FailureTracker) ProcessLine(workerID int, line string) {
	if !strings.HasPrefix(line, "##teamcity[testFailed ") {
		return
	}

	name := output.ParseTeamCityAttr(line, "name")
	if name == "" {
		return
	}

	// Strip data set suffix: "testFoo with data set ..." -> "testFoo"
	if idx := strings.Index(name, " with data set "); idx >= 0 {
		name = name[:idx]
	}

	ft.mu.Lock()
	ft.failedTests[name] = true
	ft.failedWorkers[workerID] = true
	ft.mu.Unlock()
}

// HasWorkerFailures returns true if the given worker had any test failures.
func (ft *FailureTracker) HasWorkerFailures(workerID int) bool {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	return ft.failedWorkers[workerID]
}

// HasFailures returns true if any test failures were tracked.
func (ft *FailureTracker) HasFailures() bool {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	return len(ft.failedTests) > 0
}

// Reset clears all tracked failures.
func (ft *FailureTracker) Reset() {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.failedTests = make(map[string]bool)
	ft.failedWorkers = make(map[int]bool)
}

// BuildFilter creates a PHPUnit --filter regex from the failed test names.
// It extracts the method name (after ::) and joins with | using regexp.QuoteMeta.
func (ft *FailureTracker) BuildFilter() string {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	if len(ft.failedTests) == 0 {
		return ""
	}

	methods := make(map[string]bool)
	for name := range ft.failedTests {
		method := name
		if idx := strings.LastIndex(name, "::"); idx >= 0 {
			method = name[idx+2:]
		}
		methods[method] = true
	}

	parts := make([]string, 0, len(methods))
	for method := range methods {
		parts = append(parts, regexp.QuoteMeta(method))
	}
	sort.Strings(parts)

	return strings.Join(parts, "|")
}

// trackedOutput wraps an output.Output and intercepts WorkerLine calls
// to feed them to a FailureTracker.
type trackedOutput struct {
	output.Output
	tracker *FailureTracker
}

func (t *trackedOutput) WorkerLine(workerID int, line string) {
	t.tracker.ProcessLine(workerID, line)
	t.Output.WorkerLine(workerID, line)
}
