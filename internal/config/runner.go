package config

import (
	"encoding/xml"
	"os"
	"runtime"
)

type Runner struct {
	XMLName        xml.Name `xml:"runner"`
	Workers        int      `xml:"workers"`
	Configuration  string   `xml:"configuration"`
	ConfigBuildDir string   `xml:"config-build-dir"`
	TestSuffix     string   `xml:"test-suffix"`
	Before         string   `xml:"before"`
	BeforeWorker   string   `xml:"before-worker"`
	RunWorker      string   `xml:"run-worker"`
	AfterWorker    string   `xml:"after-worker"`
	After          string   `xml:"after"`
	Filter         string   `xml:"-"` // CLI-only, not in XML config
	Group          string   `xml:"-"` // CLI-only, not in XML config
	ExcludeGroup   string   `xml:"-"` // CLI-only, not in XML config
}

func DefaultRunner() *Runner {
	return &Runner{
		Workers:        max(runtime.NumCPU()-2, 1),
		ConfigBuildDir: ".phpunit-parallel",
		RunWorker:      "vendor/bin/phpunit",
		TestSuffix:     "Test.php",
	}
}

func ParseRunner(path string) (*Runner, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultRunner()
	if err := xml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
