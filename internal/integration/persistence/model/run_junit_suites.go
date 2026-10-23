package model

import (
	"encoding/xml"
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	millisPerSecond  = 1000
	junitTimeDecimal = 3
)

type RunJunitSuites struct {
	XMLName  xml.Name        `xml:"testsuites"`
	Name     string          `xml:"name,attr"`
	Time     string          `xml:"time,attr"`
	Suites   []RunJunitSuite `xml:"testsuite"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Skipped  int             `xml:"skipped,attr"`
}

func NewRunJunitSuites(run entity.Run) RunJunitSuites {
	suites := make([]RunJunitSuite, 0, len(run.Rows))
	tests := 0
	for _, row := range run.Rows {
		suite := NewRunJunitSuite(row, run.StartedAt)
		tests += suite.Tests
		suites = append(suites, suite)
	}
	return RunJunitSuites{
		Name:     run.Target.Name,
		Time:     junitSeconds(run.Summary.DurationMs),
		Suites:   suites,
		Tests:    tests,
		Failures: run.FailedTests(),
		Errors:   run.BrokenRows(),
		Skipped:  run.SkippedTests(),
	}
}

func junitSeconds(milliseconds int64) string {
	return strconv.FormatFloat(float64(milliseconds)/millisPerSecond, 'f', junitTimeDecimal, 64)
}
