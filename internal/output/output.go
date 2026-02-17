package output

import (
	"fmt"
	"strings"
)

type StartOptions struct {
	TestCount    int
	WorkerCount  int
	Filter       string
	Group        string
	ExcludeGroup string
}

type Output interface {
	Start(opts StartOptions)
	WorkerStart(workerID, testCount int)
	WorkerLine(workerID int, line string)
	WorkerComplete(workerID int, err error)
	CleanupProgress(completed, total int)
	Finish()
	SetOnCancel(fn func())
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
