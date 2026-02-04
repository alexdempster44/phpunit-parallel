package output

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var flowIdPattern = regexp.MustCompile(` flowId='[^']*'`)

type teamCitySuite struct {
	name     string
	lines    []string
	hasTests bool
}

type teamCityWorker struct {
	suites        []teamCitySuite
	skippedSuites map[string]bool
}

type TeamCityOutput struct {
	mu            sync.Mutex
	workers       map[int]*teamCityWorker
	startedSuites map[string]bool
}

func NewTeamCityOutput() *TeamCityOutput {
	return &TeamCityOutput{
		workers:       make(map[int]*teamCityWorker),
		startedSuites: make(map[string]bool),
	}
}

func (t *TeamCityOutput) Start(testCount, workerCount int) {}

func (t *TeamCityOutput) WorkerStart(workerID, testCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.workers[workerID] = &teamCityWorker{
		skippedSuites: make(map[string]bool),
	}
}

func (t *TeamCityOutput) WorkerLine(workerID int, line string) {
	if !strings.HasPrefix(line, "##teamcity") {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	w := t.workers[workerID]
	if w == nil {
		return
	}

	cleanLine := flowIdPattern.ReplaceAllString(line, "")

	switch {
	case strings.HasPrefix(line, "##teamcity[testSuiteStarted "):
		t.handleSuiteStarted(w, line, cleanLine)

	case strings.HasPrefix(line, "##teamcity[testSuiteFinished "):
		t.handleSuiteFinished(w, line, cleanLine)

	case strings.HasPrefix(line, "##teamcity[testStarted "):
		if len(w.suites) > 0 {
			w.suites[len(w.suites)-1].hasTests = true
		}
		t.bufferLine(w, cleanLine)

	default:
		t.bufferLine(w, cleanLine)
	}
}

func (t *TeamCityOutput) handleSuiteStarted(w *teamCityWorker, line, cleanLine string) {
	name := ParseTeamCityAttr(line, "name")

	if strings.HasSuffix(name, ".xml") || t.startedSuites[name] {
		w.skippedSuites[name] = true
		return
	}

	t.startedSuites[name] = true
	w.suites = append(w.suites, teamCitySuite{
		name:  name,
		lines: []string{cleanLine},
	})
}

func (t *TeamCityOutput) handleSuiteFinished(w *teamCityWorker, line, cleanLine string) {
	name := ParseTeamCityAttr(line, "name")

	if w.skippedSuites[name] {
		delete(w.skippedSuites, name)
		return
	}

	if len(w.suites) == 0 {
		return
	}

	idx := len(w.suites) - 1
	suite := w.suites[idx]
	w.suites = w.suites[:idx]

	if suite.hasTests {
		for _, bufferedLine := range append(suite.lines, cleanLine) {
			fmt.Println(bufferedLine)
		}
	}
}

func (t *TeamCityOutput) bufferLine(w *teamCityWorker, line string) {
	if len(w.suites) > 0 {
		w.suites[len(w.suites)-1].lines = append(w.suites[len(w.suites)-1].lines, line)
	} else {
		fmt.Println(line)
	}
}

func (t *TeamCityOutput) WorkerComplete(workerID int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	w := t.workers[workerID]
	if w == nil {
		return
	}

	for _, suite := range w.suites {
		if suite.hasTests {
			for _, line := range suite.lines {
				fmt.Println(line)
			}
		}
	}
	w.suites = nil

}

func (t *TeamCityOutput) Finish() {}
