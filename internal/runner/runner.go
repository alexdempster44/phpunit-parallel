package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/alexdempster44/phpunit-parallel/internal/config"
	"github.com/alexdempster44/phpunit-parallel/internal/distributor"
	"github.com/alexdempster44/phpunit-parallel/internal/output"
)

type Runner struct {
	PHPUnitConfig *config.PHPUnit
	RunnerConfig  *config.Runner
	BaseDir       string
	Output        output.Output
	Version       string
}

func New(phpunitConfig *config.PHPUnit, runnerConfig *config.Runner, baseDir string, out output.Output) *Runner {
	return &Runner{
		PHPUnitConfig: phpunitConfig,
		RunnerConfig:  runnerConfig,
		BaseDir:       baseDir,
		Output:        out,
	}
}

func (r *Runner) Run() error {
	tests, err := r.discoverTests()
	if err != nil {
		return fmt.Errorf("failed to discover tests: %w", err)
	}

	if r.RunnerConfig.ShardTotal > 1 {
		sort.Slice(tests, func(i, j int) bool { return tests[i].Path < tests[j].Path })
		tests = distributor.Shard(tests, r.RunnerConfig.ShardIndex, r.RunnerConfig.ShardTotal)
	}

	dist := distributor.RoundRobin(tests, r.RunnerConfig.Workers)
	workers := r.createWorkers(dist)
	workerCount := len(workers)
	for _, w := range workers {
		w.WorkerCount = workerCount
	}

	if r.RunnerConfig.Before != "" {
		cmd := exec.Command("sh", "-c", r.RunnerConfig.Before)
		cmd.Dir = r.BaseDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = r.env(workerCount)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("before command failed: %w", err)
		}
	}

	tracker := NewFailureTracker()
	tracked := &trackedOutput{Output: r.Output, tracker: tracker}

	r.Output.Start(output.StartOptions{
		TestCount:    len(tests),
		WorkerCount:  len(workers),
		Filter:       r.RunnerConfig.Filter,
		Group:        r.RunnerConfig.Group,
		ExcludeGroup: r.RunnerConfig.ExcludeGroup,
		Version:      r.Version,
	})

	// uncleanedWorkers tracks workers whose environments persist (failed workers).
	uncleanedWorkers := make(map[int]*Worker)

	// Track current workers for cancel callback
	var currentWorkersMu sync.Mutex
	var currentWorkers []*Worker

	r.Output.SetOnCancel(func() {
		currentWorkersMu.Lock()
		cw := currentWorkers
		currentWorkersMu.Unlock()
		r.cleanupWorkers(cw)
	})

	retryAttempt := 0
	allWorkers := make(map[int]*Worker)
	var lastWorkerErrors map[int]error

	for {
		// Set workers to use tracked output so failure tracker sees all lines
		for _, w := range workers {
			w.Output = tracked
		}

		currentWorkersMu.Lock()
		currentWorkers = workers
		currentWorkersMu.Unlock()

		workerErrors := r.runWorkers(workers, tracked, tracker)
		lastWorkerErrors = workerErrors

		// Keep all workers open (don't run after-worker hooks yet)
		for _, w := range workers {
			uncleanedWorkers[w.ID] = w
			allWorkers[w.ID] = w
			_ = workerErrors[w.ID] // tracked by failure tracker
		}

		r.Output.Finish()

		action := r.Output.AwaitRetry()
		if action == output.ActionQuit {
			break
		}

		retryAttempt++
		filter := tracker.BuildFilter()
		tracker.Reset()

		if action == output.ActionRerunAll {
			// Rerun all workers without a filter
			workers = r.createRetryWorkers(allWorkers, "", tracked)
		} else {
			// Workers that crashed (exit code 2+) get all their files rerun (no filter)
			crashedWorkers := make(map[int]*Worker)
			// Workers with test failures (exit code 1) get only the failed tests (with filter)
			testFailedWorkers := make(map[int]*Worker)
			for id, w := range allWorkers {
				if workerErrors[id] != nil {
					crashedWorkers[id] = w
				} else if filter != "" && tracker.HasWorkerFailures(id) {
					testFailedWorkers[id] = w
				}
			}

			workers = append(
				r.createRetryWorkers(crashedWorkers, "", tracked),
				r.createRetryWorkers(testFailedWorkers, filter, tracked)...,
			)
		}
		if len(workers) == 0 {
			break
		}

		workerIDs := make([]int, len(workers))
		retryTestCount := 0
		for i, w := range workers {
			workerIDs[i] = w.ID
			retryTestCount += len(w.Tests)
		}

		r.Output.RetryStart(output.RetryStartOptions{
			Attempt:     retryAttempt,
			TestCount:   retryTestCount,
			WorkerCount: len(workers),
			WorkerIDs:   workerIDs,
		})
	}

	// Final cleanup: clean up all remaining uncleaned workers
	remaining := make([]*Worker, 0, len(uncleanedWorkers))
	for _, w := range uncleanedWorkers {
		remaining = append(remaining, w)
	}
	r.cleanupWorkers(remaining)

	if r.RunnerConfig.After != "" {
		cmd := exec.Command("sh", "-c", r.RunnerConfig.After)
		cmd.Dir = r.BaseDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = r.env(workerCount)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("after command failed: %w", err)
		}
	}

	// Collect worker errors in deterministic order
	var errs []error
	var workerIDs []int
	for id := range lastWorkerErrors {
		if lastWorkerErrors[id] != nil {
			workerIDs = append(workerIDs, id)
		}
	}
	if len(workerIDs) > 0 {
		sort.Ints(workerIDs)
		for _, id := range workerIDs {
			errs = append(errs, fmt.Errorf("worker %d: %w", id, lastWorkerErrors[id]))
		}
	}

	if tracker.HasFailures() {
		errs = append(errs, fmt.Errorf("some tests failed"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// runWorkers launches all workers in parallel and returns a map of workerID -> error (nil if succeeded).
func (r *Runner) runWorkers(workers []*Worker, out output.Output, tracker *FailureTracker) map[int]error {
	workerErrors := make(map[int]error)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, worker := range workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()

			out.WorkerStart(w.ID, w.TestCount())
			err := w.Run()
			// If the failure tracker captured test failures for this worker,
			// the non-zero exit is from PHPUnit reporting failures, not a crash.
			if err != nil && tracker.HasWorkerFailures(w.ID) {
				err = nil
			}
			out.WorkerComplete(w.ID, err)

			mu.Lock()
			workerErrors[w.ID] = err
			mu.Unlock()
		}(worker)
	}

	wg.Wait()
	return workerErrors
}

// cleanupWorkers runs after-worker hooks for all provided workers.
func (r *Runner) cleanupWorkers(workers []*Worker) {
	if r.RunnerConfig.AfterWorker == "" || len(workers) == 0 {
		return
	}
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	total := len(workers)
	var completed atomic.Int32
	r.Output.CleanupProgress(0, total)
	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			w.runAfterWorker()
			done := int(completed.Add(1))
			r.Output.CleanupProgress(done, total)
		}(w)
	}
	wg.Wait()
}

// createRetryWorkers creates new Worker structs from uncleaned workers with IsRetry=true and the retry filter.
func (r *Runner) createRetryWorkers(uncleaned map[int]*Worker, filter string, out output.Output) []*Worker {
	workers := make([]*Worker, 0, len(uncleaned))
	for _, orig := range uncleaned {
		w := NewWorker(
			orig.ID,
			orig.Tests,
			orig.BeforeWorker,
			orig.RunWorker,
			orig.AfterWorker,
			orig.BaseDir,
			orig.ConfigBuildDir,
			orig.Bootstrap,
			orig.RawConfigXML,
			out,
			filter,
			orig.Group,
			orig.ExcludeGroup,
		)
		w.WorkerCount = orig.WorkerCount
		w.IsRetry = true
		workers = append(workers, w)
	}
	return workers
}

func (r *Runner) env(workerCount int) []string {
	return append(os.Environ(),
		"PARALLEL=1",
		fmt.Sprintf("PROJECT=%s", filepath.Base(r.BaseDir)),
		fmt.Sprintf("RUNNER_PID=%d", os.Getpid()),
		fmt.Sprintf("WORKER_COUNT=%d", workerCount),
	)
}

func (r *Runner) createWorkers(dist distributor.Distribution) []*Worker {
	var workers []*Worker
	for _, bucket := range dist.Workers {
		if len(bucket.Tests) == 0 {
			continue
		}
		workers = append(workers, NewWorker(
			bucket.WorkerID,
			bucket.Tests,
			r.RunnerConfig.BeforeWorker,
			r.RunnerConfig.RunWorker,
			r.RunnerConfig.AfterWorker,
			r.BaseDir,
			r.RunnerConfig.ConfigBuildDir,
			r.PHPUnitConfig.Bootstrap,
			r.PHPUnitConfig.RawXML,
			r.Output,
			r.RunnerConfig.Filter,
			r.RunnerConfig.Group,
			r.RunnerConfig.ExcludeGroup,
		))
	}
	return workers
}

func (r *Runner) discoverTests() ([]distributor.TestFile, error) {
	var tests []distributor.TestFile

	for _, suite := range r.PHPUnitConfig.TestSuites.TestSuites {
		for _, dir := range suite.Directories {
			dirPath := filepath.Join(r.BaseDir, dir)
			files, err := r.findTestFiles(dirPath, suite.Name, suite.Exclude)
			if err != nil {
				return nil, fmt.Errorf("failed to scan directory %s: %w", dir, err)
			}
			tests = append(tests, files...)
		}

		for _, file := range suite.Files {
			filePath := filepath.Join(r.BaseDir, file)
			if _, err := os.Stat(filePath); err == nil {
				tests = append(tests, distributor.TestFile{
					Path:  filePath,
					Suite: suite.Name,
				})
			}
		}
	}

	return tests, nil
}

func (r *Runner) findTestFiles(dir, suiteName string, excludes []string) ([]distributor.TestFile, error) {
	var tests []distributor.TestFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, r.RunnerConfig.TestSuffix) {
			return nil
		}

		for _, exclude := range excludes {
			excludePath := filepath.Join(r.BaseDir, exclude)
			if matched, _ := filepath.Match(excludePath, path); matched {
				return nil
			}
			if strings.HasPrefix(path, excludePath) {
				return nil
			}
		}

		tests = append(tests, distributor.TestFile{
			Path:  path,
			Suite: suiteName,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tests, nil
}
