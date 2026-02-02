package config

import (
	"encoding/xml"
	"os"
	"runtime"
)

type Runner struct {
	XMLName        xml.Name `xml:"runner"`
	Workers        int      `xml:"workers"`
	ConfigBuildDir string   `xml:"config-build-dir"`
	RunCommand     string   `xml:"run-command"`
}

func DefaultRunner() *Runner {
	return &Runner{
		Workers:        runtime.NumCPU(),
		ConfigBuildDir: ".phpunit-parallel",
		RunCommand:     "vendor/bin/phpunit",
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
