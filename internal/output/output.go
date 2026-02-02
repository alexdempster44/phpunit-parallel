package output

import (
	"fmt"
	"strings"
)

type Output interface {
	Start(testCount, workerCount int)
	WorkerStart(workerID, testCount int)
	WorkerLine(workerID int, line string)
	WorkerComplete(workerID int, err error)
	Finish()
}

func parseTeamCityAttr(line, attr string) string {
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

func parseTeamCityCount(line string) *int {
	countStr := parseTeamCityAttr(line, "count")
	if countStr == "" {
		return nil
	}
	var count int
	if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
		return nil
	}
	return &count
}

func parseTeamCityError(line string) (name, message, details string) {
	return parseTeamCityAttr(line, "name"), parseTeamCityAttr(line, "message"), parseTeamCityAttr(line, "details")
}
