package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexdempster44/phpunit-parallel/internal/output"
)

type TUIOutput struct {
	program  *tea.Program
	model    *Model
	mu       sync.Mutex
	onCancel func()
	stopped  bool
}

func New() *TUIOutput {
	return &TUIOutput{}
}

func (t *TUIOutput) Start(opts output.StartOptions) {
	t.model = NewModel(opts)
	t.program = tea.NewProgram(t.model, tea.WithAltScreen())

	go func() {
		_, _ = t.program.Run()
		if t.model.quitting && t.model.phase != PhaseComplete {
			t.mu.Lock()
			t.stopped = true
			t.mu.Unlock()
			if t.onCancel != nil {
				t.onCancel()
			}
			os.Exit(130)
		}
	}()
}

func (t *TUIOutput) WorkerStart(workerID, testCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.program != nil {
		t.program.Send(WorkerStartMsg{
			WorkerID:  workerID,
			TestCount: testCount,
		})
	}
}

func (t *TUIOutput) WorkerLine(workerID int, line string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.program == nil {
		return
	}

	switch {
	case strings.HasPrefix(line, "##teamcity[testCount "):
		count := output.ParseTeamCityCount(line)
		if count != nil {
			t.program.Send(TestCountMsg{
				WorkerID: workerID,
				Count:    *count,
			})
		}

	case strings.HasPrefix(line, "##teamcity[testStarted "):
		key := output.ParseTeamCityAttr(line, "name")
		displayName := output.ParseTeamCityTestName(line)
		t.program.Send(TestStartMsg{
			WorkerID:    workerID,
			TestKey:     key,
			DisplayName: displayName,
		})

	case strings.HasPrefix(line, "##teamcity[testFailed "):
		name, message, details := output.ParseTeamCityError(line)
		t.program.Send(TestFailMsg{
			WorkerID: workerID,
			TestName: name,
			Message:  message,
			Details:  details,
		})

	case strings.HasPrefix(line, "##teamcity[testIgnored "):
		name := output.ParseTeamCityAttr(line, "name")
		message := output.ParseTeamCityAttr(line, "message")
		t.program.Send(TestSkipMsg{
			WorkerID: workerID,
			TestName: name,
			Message:  message,
		})

	case strings.HasPrefix(line, "##teamcity[testFinished "):
		name := output.ParseTeamCityAttr(line, "name")
		t.program.Send(TestPassMsg{
			WorkerID: workerID,
			TestName: name,
		})
	}
}

func (t *TUIOutput) WorkerComplete(workerID int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.program != nil {
		t.program.Send(WorkerCompleteMsg{
			WorkerID: workerID,
			Error:    err,
		})
	}
}

func (t *TUIOutput) CleanupProgress(completed, total int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		fmt.Fprintf(os.Stderr, "\rCleaning up workers... %d/%d", completed, total)
		if completed >= total {
			fmt.Fprintln(os.Stderr)
		}
		return
	}

	if t.program != nil {
		t.program.Send(CleanupProgressMsg{
			Completed: completed,
			Total:     total,
		})
	}
}

func (t *TUIOutput) SetOnCancel(fn func()) {
	t.onCancel = fn
}

func (t *TUIOutput) Finish() {
	t.mu.Lock()
	if t.program != nil {
		t.program.Send(FinishMsg{})
	}
	t.mu.Unlock()

	if t.program != nil {
		t.program.Wait()
	}
}
