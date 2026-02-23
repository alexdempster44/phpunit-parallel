package output

import (
	"fmt"
	"regexp"
	"strings"
)

// DeprecationInfo represents a single deprecation extracted from PHPUnit output.
type DeprecationInfo struct {
	TestName string // FQCN::method (from "Triggered by:" section) or source location
	Message  string // The deprecation message
	Source   string // Source file:line where the deprecation was triggered
}

type deprecationState int

const (
	depStateIdle deprecationState = iota
	depStateAwaitingItem
	depStateReadingMessage
	depStateAwaitingTriggeredBy
	depStateReadingTestName
)

var deprecationHeaderRegex = regexp.MustCompile(`^\d+ tests? triggered \d+ deprecations?:`)
var deprecationItemRegex = regexp.MustCompile(`^(\d+)\) (.+)$`)
var deprecationTestNameRegex = regexp.MustCompile(`^\* (.+)$`)

// DeprecationParser parses PHPUnit plain-text deprecation output line by line.
type DeprecationParser struct {
	state         deprecationState
	source        string
	message       strings.Builder
	testName      string
	OnDeprecation func(info DeprecationInfo)
}

// NewDeprecationParser creates a parser that calls onDeprecation for each deprecation found.
func NewDeprecationParser(onDeprecation func(info DeprecationInfo)) *DeprecationParser {
	return &DeprecationParser{
		OnDeprecation: onDeprecation,
	}
}

// ParseLine processes a single line of PHPUnit output.
func (p *DeprecationParser) ParseLine(line string) {
	switch p.state {
	case depStateIdle:
		if deprecationHeaderRegex.MatchString(line) {
			p.state = depStateAwaitingItem
		}

	case depStateAwaitingItem:
		if m := deprecationItemRegex.FindStringSubmatch(line); m != nil {
			p.source = m[2]
			p.message.Reset()
			p.testName = ""
			p.state = depStateReadingMessage
		}

	case depStateReadingMessage:
		if line == "" {
			p.state = depStateAwaitingTriggeredBy
		} else {
			if p.message.Len() > 0 {
				p.message.WriteByte('\n')
			}
			p.message.WriteString(line)
		}

	case depStateAwaitingTriggeredBy:
		switch {
		case line == "Triggered by:":
			p.state = depStateReadingTestName
		case deprecationItemRegex.MatchString(line):
			// New item without "Triggered by:" section - emit current and start new
			p.emit()
			m := deprecationItemRegex.FindStringSubmatch(line)
			p.source = m[2]
			p.message.Reset()
			p.testName = ""
			p.state = depStateReadingMessage
		case strings.HasPrefix(line, "OK") || strings.HasPrefix(line, "ERRORS") ||
			strings.HasPrefix(line, "FAILURES") || strings.HasPrefix(line, "Tests:"):
			p.emit()
			p.state = depStateIdle
		}

	case depStateReadingTestName:
		if m := deprecationTestNameRegex.FindStringSubmatch(line); m != nil {
			p.testName = m[1]
			p.emit()
			p.state = depStateAwaitingItem
		}
	}
}

// Flush emits any pending deprecation (called when output ends).
func (p *DeprecationParser) Flush() {
	if p.state != depStateIdle && p.message.Len() > 0 {
		p.emit()
	}
	p.state = depStateIdle
}

func (p *DeprecationParser) emit() {
	if p.OnDeprecation == nil || p.message.Len() == 0 {
		return
	}
	info := DeprecationInfo{
		TestName: p.testName,
		Message:  p.message.String(),
		Source:   p.source,
	}
	if info.TestName == "" {
		info.TestName = p.source
	}
	p.OnDeprecation(info)
	p.message.Reset()
	p.source = ""
	p.testName = ""
}

// Reset clears the parser state for reuse (e.g., on retry).
func (p *DeprecationParser) Reset() {
	p.state = depStateIdle
	p.source = ""
	p.message.Reset()
	p.testName = ""
}

type StartOptions struct {
	TestCount    int
	WorkerCount  int
	Filter       string
	Group        string
	ExcludeGroup string
	Version      string
}

type RetryAction int

const (
	ActionRetry RetryAction = iota
	ActionRerunAll
	ActionQuit
)

type RetryStartOptions struct {
	Attempt     int
	TestCount   int
	WorkerCount int
	WorkerIDs   []int
}

type Output interface {
	Start(opts StartOptions)
	WorkerStart(workerID, testCount int)
	WorkerLine(workerID int, line string)
	WorkerComplete(workerID int, err error)
	CleanupProgress(completed, total int)
	Finish()
	SetOnCancel(fn func())
	AwaitRetry() RetryAction
	RetryStart(opts RetryStartOptions)
}

func ParseTeamCityAttr(line, attr string) string {
	prefix := attr + "='"
	start := strings.Index(line, prefix)
	if start < 0 {
		return ""
	}
	start += len(prefix)

	end := start
	for end < len(line) {
		if line[end] == '\'' && (end == start || line[end-1] != '|') {
			break
		}
		end++
	}

	value := line[start:end]
	value = strings.ReplaceAll(value, "|'", "'")
	value = strings.ReplaceAll(value, "|n", "\n")
	value = strings.ReplaceAll(value, "|r", "\r")
	value = strings.ReplaceAll(value, "||", "|")
	value = strings.ReplaceAll(value, "|[", "[")
	value = strings.ReplaceAll(value, "|]", "]")
	return value
}

func ParseTeamCityCount(line string) *int {
	countStr := ParseTeamCityAttr(line, "count")
	if countStr == "" {
		return nil
	}
	var count int
	if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
		return nil
	}
	return &count
}

func ParseTeamCityError(line string) (name, message, details string) {
	return ParseTeamCityAttr(line, "name"), ParseTeamCityAttr(line, "message"), ParseTeamCityAttr(line, "details")
}

func ParseTeamCityTestName(line string) string {
	locationHint := ParseTeamCityAttr(line, "locationHint")
	if locationHint != "" {
		if _, after, found := strings.Cut(locationHint, "::"); found {
			return after
		}
	}
	return ParseTeamCityAttr(line, "name")
}
