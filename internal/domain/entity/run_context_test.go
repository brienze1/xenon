package entity

import "testing"

func TestRunContextInEnvironmentOnlySetsTheEnvironment(t *testing.T) {
	context := NoRunContext().InEnvironment("prod")
	if context.EnvironmentID != "prod" || context.IterationCount != 1 || context.IsRun() {
		t.Errorf("context = %+v", context)
	}
}
