package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type WorkerStartMsg struct {
	WorkerID  int
	TestCount int
}

type TestStartMsg struct {
	WorkerID    int
	TestKey     string
	DisplayName string
}

type TestPassMsg struct {
	WorkerID int
	TestName string
}

type TestFailMsg struct {
	WorkerID int
	TestName string
	Message  string
	Details  string
}

type TestSkipMsg struct {
	WorkerID int
	TestName string
	Message  string
}

type TestCountMsg struct {
	WorkerID int
	Count    int
}

type WorkerCompleteMsg struct {
	WorkerID int
	Error    error
}

type FinishMsg struct{}

type TickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}
