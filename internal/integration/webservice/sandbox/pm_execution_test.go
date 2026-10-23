package sandbox

import (
	"context"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestSetNextRequestRecordsTheLastCall(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		target string
		stop   bool
	}{
		{name: "by name", text: `pm.execution.setNextRequest("B"); pm.execution.setNextRequest("D")`, target: "D"},
		{name: "null", text: `pm.execution.setNextRequest(null)`, stop: true},
		{name: "no argument", text: `pm.execution.setNextRequest()`, stop: true},
		{name: "legacy", text: `postman.setNextRequest("C")`, target: "C"},
		{name: "before a skip", text: `pm.execution.setNextRequest("D"); pm.execution.skipRequest()`, target: "D"},
	}
	for _, test := range cases {
		result := run(t, test.text)
		if result.HasFailed() || result.NextRequest == nil || result.NextRequest.Target != test.target || result.NextRequest.Stop != test.stop {
			t.Errorf("%s: run = %+v, next = %+v", test.name, result.Error, result.NextRequest)
		}
	}
	if result := run(t, `pm.test("x", function () {})`); result.NextRequest != nil {
		t.Errorf("no call means no next request: %+v", result.NextRequest)
	}
}

func TestIterationInfoAndDataReachPm(t *testing.T) {
	scope := entity.ScriptContext{
		Info:          entity.NewScriptInfo("01R", "Login", "Login", nil).ForIteration(2, 3),
		IterationData: map[string]any{"email": "a@x.test", "n": 1.5, "obj": map[string]any{"a": 1}},
	}
	text := `pm.test(pm.info.iteration + "/" + pm.info.iterationCount + " " + pm.iterationData.get("email") + " " + typeof pm.iterationData.get("n") + " " + pm.iterationData.get("obj").a + " " + pm.variables.get("email"), function () {})`
	result := testRunner(30*time.Second, memoryGrowthLimit).Run(context.Background(), script(text), scope, nil, nil, entity.DefaultAppSettings())
	if result.HasFailed() || len(result.Tests) != 1 || result.Tests[0].Name != "2/3 a@x.test number 1 a@x.test" {
		t.Fatalf("run = %+v, tests = %+v", result.Error, result.Tests)
	}
}
