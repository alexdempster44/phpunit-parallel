package config

import (
	"encoding/xml"
	"os"
)

type PHPUnit struct {
	XMLName    xml.Name   `xml:"phpunit"`
	Bootstrap  string     `xml:"bootstrap,attr"`
	TestSuites TestSuites `xml:"testsuites"`
}

type TestSuites struct {
	TestSuites []TestSuite `xml:"testsuite"`
}

type TestSuite struct {
	Name        string   `xml:"name,attr"`
	Directories []string `xml:"directory"`
	Files       []string `xml:"file"`
	Exclude     []string `xml:"exclude"`
}

func ParsePHPUnit(path string) (*PHPUnit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config PHPUnit
	if err := xml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
