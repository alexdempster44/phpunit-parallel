package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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

	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			if r.RunnerConfig.AfterWorker == "" {
				return
			}
			// Ignore signals during cleanup so a second Ctrl+C doesn't kill the process
			signal.Ignore(syscall.SIGINT, syscall.SIGTERM)
			defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)
			total := len(workers)
			var completed atomic.Int32
			r.Output.CleanupProgress(0, total)
			var cwg sync.WaitGroup
			for _, w := range workers {
				cwg.Add(1)
				go func(w *Worker) {
					defer cwg.Done()
					w.runAfterWorker()
					done := int(completed.Add(1))
					r.Output.CleanupProgress(done, total)
				}(w)
			}
			cwg.Wait()
		})
	}

	r.Output.SetOnCancel(cleanup)
	r.Output.Start(output.StartOptions{
		TestCount:    len(tests),
		WorkerCount:  len(workers),
		Filter:       r.RunnerConfig.Filter,
		Group:        r.RunnerConfig.Group,
		ExcludeGroup: r.RunnerConfig.ExcludeGroup,
	})

	var wg sync.WaitGroup

	for _, worker := range workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()

			r.Output.WorkerStart(w.ID, w.TestCount())
			err := w.Run()
			r.Output.WorkerComplete(w.ID, err)
		}(worker)
	}

	wg.Wait()
	cleanup()
	r.Output.Finish()

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

	return nil
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
