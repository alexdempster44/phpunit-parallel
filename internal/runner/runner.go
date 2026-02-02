package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

	r.Output.Start(len(tests), len(workers))

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
	r.Output.Finish()

	return nil
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
			r.RunnerConfig.RunCommand,
			r.BaseDir,
			r.RunnerConfig.ConfigBuildDir,
			r.PHPUnitConfig.Bootstrap,
			r.Output,
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

		if !strings.HasSuffix(path, "Test.php") {
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
