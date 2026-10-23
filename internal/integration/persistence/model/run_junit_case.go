package model

import "github.com/brienze1/xenon/internal/domain/entity"

const junitErrorCase = "request"

type RunJunitCase struct {
	Failure   *RunJunitMessage `xml:"failure,omitempty"`
	Error     *RunJunitMessage `xml:"error,omitempty"`
	Skipped   *struct{}        `xml:"skipped,omitempty"`
	Name      string           `xml:"name,attr"`
	Classname string           `xml:"classname,attr"`
	Time      string           `xml:"time,attr"`
}

func NewRunJunitCase(test entity.TestResult, classname string) RunJunitCase {
	testCase := RunJunitCase{Name: test.Name, Classname: classname, Time: junitSeconds(0)}
	switch {
	case test.IsFailed():
		testCase.Failure = NewRunJunitMessage(test.ErrorMessage)
	case test.IsSkipped():
		testCase.Skipped = &struct{}{}
	}
	return testCase
}

func NewRunJunitErrorCase(row entity.RunRowResult) RunJunitCase {
	return RunJunitCase{Name: junitErrorCase, Classname: row.Name, Time: junitSeconds(row.DurationMs), Error: NewRunJunitMessage(row.BreakMessage())}
}
