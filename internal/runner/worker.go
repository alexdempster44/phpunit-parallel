package runner

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexdempster44/phpunit-parallel/internal/distributor"
	"github.com/alexdempster44/phpunit-parallel/internal/output"
)

type Worker struct {
	ID             int
	Tests          []distributor.TestFile
	RunCommand     string
	BaseDir        string
	ConfigBuildDir string
	Bootstrap      string
	Output         output.Output
}

func NewWorker(id int, tests []distributor.TestFile, runCommand, baseDir, configBuildDir, bootstrap string, out output.Output) *Worker {
	return &Worker{
		ID:             id,
		Tests:          tests,
		RunCommand:     runCommand,
		BaseDir:        baseDir,
		ConfigBuildDir: configBuildDir,
		Bootstrap:      bootstrap,
		Output:         out,
	}
}

func (w *Worker) Run() error {
	configPath, err := w.buildConfig()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	defer os.Remove(configPath)

	args := []string{"--configuration", configPath, "--teamcity"}
	var cmd *exec.Cmd
	if strings.Contains(w.RunCommand, " ") {
		shellArgs := w.RunCommand + " " + strings.Join(args, " ")
		cmd = exec.Command("sh", "-c", shellArgs)
	} else {
		cmd = exec.Command(w.RunCommand, args...)
	}
	cmd.Dir = w.BaseDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		w.Output.WorkerLine(w.ID, scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (w *Worker) TestCount() int {
	return len(w.Tests)
}

func (w *Worker) buildConfig() (string, error) {
	type testFile struct {
		XMLName xml.Name `xml:"file"`
		Path    string   `xml:",chardata"`
	}

	type testSuite struct {
		XMLName xml.Name   `xml:"testsuite"`
		Name    string     `xml:"name,attr"`
		Files   []testFile `xml:"file"`
	}

	type testSuites struct {
		XMLName    xml.Name    `xml:"testsuites"`
		TestSuites []testSuite `xml:"testsuite"`
	}

	type phpunit struct {
		XMLName    xml.Name   `xml:"phpunit"`
		Bootstrap  string     `xml:"bootstrap,attr,omitempty"`
		TestSuites testSuites `xml:"testsuites"`
	}

	suiteMap := make(map[string][]testFile)
	for _, test := range w.Tests {
		relPath, _ := filepath.Rel(w.BaseDir, test.Path)
		pathFromConfig := filepath.Join("..", relPath)
		suiteMap[test.Suite] = append(suiteMap[test.Suite], testFile{Path: pathFromConfig})
	}

	var suites []testSuite
	for name, files := range suiteMap {
		suites = append(suites, testSuite{Name: name, Files: files})
	}

	bootstrap := w.Bootstrap
	if bootstrap != "" {
		bootstrap = filepath.Join("..", bootstrap)
	}

	cfg := phpunit{
		Bootstrap:  bootstrap,
		TestSuites: testSuites{TestSuites: suites},
	}

	data, err := xml.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(w.ConfigBuildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(w.ConfigBuildDir, fmt.Sprintf("phpunit-worker-%d.xml", w.ID))
	if err := os.WriteFile(configPath, append([]byte(xml.Header), data...), 0644); err != nil {
		return "", fmt.Errorf("failed to write config: %w", err)
	}

	return configPath, nil
}
