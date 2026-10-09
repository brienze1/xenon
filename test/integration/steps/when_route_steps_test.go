package steps

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerWhenRouteSteps(ctx *godog.ScenarioContext) {
	ctx.When(`^I request "([^"]*)" from the "(\w+)" route$`, t.iRequestFromTheRoute)
}

func (t *testUtils) iRequestFromTheRoute(path, name string) error {
	target, err := t.service(name)
	if err != nil {
		return err
	}
	handler, isHandler := target.Interface().(http.Handler)
	if !isHandler {
		return fmt.Errorf("binding %s is not an HTTP route", name)
	}
	recorder := httptest.NewRecorder()
	started := time.Now()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, t.substitute(path), nil))
	t.elapsed = time.Since(started)
	t.routeResponse = recorder
	return nil
}
