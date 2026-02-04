package tui

import "time"

type TestStatus int

const (
	StatusPending TestStatus = iota
	StatusRunning
	StatusPassed
	StatusFailed
	StatusSkipped
)

type TestNode struct {
	Key          string
	Name         string
	Status       TestStatus
	ErrorMessage string
	ErrorDetails string
	WorkerID     int
}

type WorkerNode struct {
	ID           int
	Tests        []*TestNode
	Completed    int
	Total        int
	Failed       int
	HasTestCount bool
	TestFiles    int
}

type ErrorEntry struct {
	TestName string
	Message  string
	Details  string
	WorkerID int
	Expanded bool
}

type RunPhase int

const (
	PhaseRunning RunPhase = iota
	PhaseComplete
	PhaseExploring
)

type Panel int

const (
	PanelWorkers Panel = iota
	PanelRunning
	PanelErrors
)

type Model struct {
	workers       map[int]*WorkerNode
	workerOrder   []int
	errors        []ErrorEntry
	phase         RunPhase
	activePanel   Panel
	runningCursor int
	errorCursor   int
	runningOffset int
	errorOffset   int
	workersOffset int
	testCount     int
	workerCount   int
	startTime     time.Time
	endTime       time.Time
	width         int
	height        int
	quitting      bool
	hasTestCount  bool
	totalComplete int
	totalFailed   int
	totalSkipped  int
}

func NewModel(testCount, workerCount int) *Model {
	m := &Model{
		workers:     make(map[int]*WorkerNode),
		workerOrder: make([]int, 0, workerCount),
		errors:      make([]ErrorEntry, 0),
		phase:       PhaseRunning,
		activePanel: PanelErrors,
		testCount:   testCount,
		workerCount: workerCount,
		startTime:   time.Now(),
		width:       80,
		height:      24,
	}

	for i := range workerCount {
		m.workers[i] = &WorkerNode{
			ID:        i,
			Tests:     make([]*TestNode, 0),
			TestFiles: 0,
		}
		m.workerOrder = append(m.workerOrder, i)
	}

	return m
}
