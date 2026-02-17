package runner

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/alexdempster44/phpunit-parallel/internal/distributor"
	"github.com/alexdempster44/phpunit-parallel/internal/output"
)

type Worker struct {
	ID             int
	Tests          []distributor.TestFile
	BeforeWorker   string
	RunWorker      string
	AfterWorker    string
	BaseDir        string
	ConfigBuildDir string
	Bootstrap      string
	RawConfigXML   []byte
	Output         output.Output
	Filter         string
	Group          string
	ExcludeGroup   string
	WorkerCount    int
}

func NewWorker(id int, tests []distributor.TestFile, beforeWorker, runWorker, afterWorker, baseDir, configBuildDir, bootstrap string, rawConfigXML []byte, out output.Output, filter, group, excludeGroup string) *Worker {
	return &Worker{
		ID:             id,
		Tests:          tests,
		BeforeWorker:   beforeWorker,
		RunWorker:      runWorker,
		AfterWorker:    afterWorker,
		BaseDir:        baseDir,
		ConfigBuildDir: configBuildDir,
		Bootstrap:      bootstrap,
		RawConfigXML:   rawConfigXML,
		Output:         out,
		Filter:         filter,
		Group:          group,
		ExcludeGroup:   excludeGroup,
	}
}

func (w *Worker) Run() error {
	if w.BeforeWorker != "" {
		if err := w.runHook(w.BeforeWorker); err != nil {
			return fmt.Errorf("before-worker failed: %w", err)
		}
	}

	configPath, err := w.buildConfig()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	defer func() { _ = os.Remove(configPath) }()

	args := []string{"--configuration", configPath, "--teamcity"}
	if w.Filter != "" {
		args = append(args, "--filter", w.Filter)
	}
	if w.Group != "" {
		args = append(args, "--group", w.Group)
	}
	if w.ExcludeGroup != "" {
		args = append(args, "--exclude-group", w.ExcludeGroup)
	}
	joinedArgs := strings.Join(args, " ")
	var cmd *exec.Cmd
	if strings.Contains(w.RunWorker, "{}") {
		shellArgs := strings.ReplaceAll(w.RunWorker, "{}", joinedArgs)
		cmd = exec.Command("sh", "-c", shellArgs)
	} else if strings.Contains(w.RunWorker, " ") {
		cmd = exec.Command("sh", "-c", w.RunWorker+" "+joinedArgs)
	} else {
		cmd = exec.Command(w.RunWorker, args...)
	}
	cmd.Dir = w.BaseDir
	cmd.Env = w.env()

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

func (w *Worker) runHook(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = w.BaseDir
	cmd.Env = w.env()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd.Run()
}

func (w *Worker) env() []string {
	return append(os.Environ(),
		"PARALLEL=1",
		fmt.Sprintf("PROJECT=%s", filepath.Base(w.BaseDir)),
		fmt.Sprintf("RUNNER_PID=%d", os.Getpid()),
		fmt.Sprintf("WORKER_ID=%d", w.ID),
		fmt.Sprintf("WORKER_COUNT=%d", w.WorkerCount),
	)
}

func (w *Worker) runAfterWorker() {
	if w.AfterWorker == "" {
		return
	}
	_ = w.runHook(w.AfterWorker)
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

	newTestSuites := testSuites{TestSuites: suites}
	testSuitesXML, err := xml.MarshalIndent(newTestSuites, "  ", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal testsuites: %w", err)
	}

	configXML := string(w.RawConfigXML)

	testSuitesRegex := regexp.MustCompile(`(?s)<testsuites[^>]*>.*?</testsuites>`)
	configXML = testSuitesRegex.ReplaceAllString(configXML, string(testSuitesXML))

	if w.Bootstrap != "" {
		bootstrapRegex := regexp.MustCompile(`bootstrap="([^"]*)"`)
		configXML = bootstrapRegex.ReplaceAllStringFunc(configXML, func(match string) string {
			return fmt.Sprintf(`bootstrap="%s"`, filepath.Join("..", w.Bootstrap))
		})
	}

	if err := os.MkdirAll(w.ConfigBuildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(w.ConfigBuildDir, fmt.Sprintf("phpunit-worker-%d.xml", w.ID))
	if err := os.WriteFile(configPath, []byte(configXML), 0644); err != nil {
		return "", fmt.Errorf("failed to write config: %w", err)
	}

	return configPath, nil
}
