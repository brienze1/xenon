package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const junitTimestampLayout = "2006-01-02T15:04:05"

type RunJunitSuite struct {
	Name      string         `xml:"name,attr"`
	Timestamp string         `xml:"timestamp,attr"`
	Time      string         `xml:"time,attr"`
	Cases     []RunJunitCase `xml:"testcase"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Skipped   int            `xml:"skipped,attr"`
}

func NewRunJunitSuite(row entity.RunRowResult, startedAt time.Time) RunJunitSuite {
	cases := make([]RunJunitCase, 0, len(row.Tests)+1)
	for _, test := range row.Tests {
		cases = append(cases, NewRunJunitCase(test, row.Name))
	}
	errored := 0
	if row.Breaks() {
		cases = append(cases, NewRunJunitErrorCase(row))
		errored++
	}
	return RunJunitSuite{
		Name:      row.IterationLabel(),
		Timestamp: startedAt.UTC().Format(junitTimestampLayout),
		Time:      junitSeconds(row.DurationMs),
		Cases:     cases,
		Tests:     len(cases),
		Failures:  row.FailedTests(),
		Errors:    errored,
		Skipped:   row.SkippedTests(),
	}
}
