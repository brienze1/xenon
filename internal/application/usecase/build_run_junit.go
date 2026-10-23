package usecase

import (
	"encoding/xml"
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	junitErrorCase   = "request"
	millisPerSecond  = 1000
	junitTimeDecimal = 3
)

type junitSuites struct {
	XMLName  xml.Name     `xml:"testsuites"`
	Name     string       `xml:"name,attr"`
	Time     string       `xml:"time,attr"`
	Suites   []junitSuite `xml:"testsuite"`
	Tests    int          `xml:"tests,attr"`
	Failures int          `xml:"failures,attr"`
	Errors   int          `xml:"errors,attr"`
	Skipped  int          `xml:"skipped,attr"`
}

type junitSuite struct {
	Name      string      `xml:"name,attr"`
	Timestamp string      `xml:"timestamp,attr"`
	Time      string      `xml:"time,attr"`
	Cases     []junitCase `xml:"testcase"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
	Errors    int         `xml:"errors,attr"`
	Skipped   int         `xml:"skipped,attr"`
}

type junitCase struct {
	Failure   *junitMessage `xml:"failure,omitempty"`
	Error     *junitMessage `xml:"error,omitempty"`
	Skipped   *struct{}     `xml:"skipped,omitempty"`
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
}

type junitMessage struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

// BuildRunJUnit writes a functional run as JUnit XML: a testsuite per executed row and iteration, a testcase per test.
func BuildRunJUnit(run entity.Run) ([]byte, error) {
	suites := junitSuites{Name: run.Target.Name, Time: junitSeconds(run.Summary.DurationMs), Suites: make([]junitSuite, 0, len(run.Rows))}
	for _, row := range run.Rows {
		suite := newJunitSuite(row, run)
		suites.Tests += suite.Tests
		suites.Failures += suite.Failures
		suites.Errors += suite.Errors
		suites.Skipped += suite.Skipped
		suites.Suites = append(suites.Suites, suite)
	}
	encoded, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(append([]byte(xml.Header), encoded...), '\n'), nil
}

func newJunitSuite(row entity.RunRowResult, run entity.Run) junitSuite {
	suite := junitSuite{
		Name:      row.Path() + " [iteration " + strconv.Itoa(row.Iteration+1) + "]",
		Timestamp: run.StartedAt.UTC().Format("2006-01-02T15:04:05"),
		Time:      junitSeconds(row.DurationMs),
		Cases:     make([]junitCase, 0, len(row.Tests)+1),
	}
	for _, test := range row.Tests {
		testCase := junitCase{Name: test.Name, Classname: row.Name, Time: junitSeconds(0)}
		switch {
		case test.IsFailed():
			testCase.Failure = &junitMessage{Message: test.ErrorMessage, Text: test.ErrorMessage}
			suite.Failures++
		case test.IsSkipped():
			testCase.Skipped = &struct{}{}
			suite.Skipped++
		}
		suite.Cases = append(suite.Cases, testCase)
	}
	if row.Breaks() {
		message := row.BreakMessage()
		suite.Cases = append(suite.Cases, junitCase{Name: junitErrorCase, Classname: row.Name, Time: suite.Time, Error: &junitMessage{Message: message, Text: message}})
		suite.Errors++
	}
	suite.Tests = len(suite.Cases)
	return suite
}

func junitSeconds(milliseconds int64) string {
	return strconv.FormatFloat(float64(milliseconds)/millisPerSecond, 'f', junitTimeDecimal, 64)
}
