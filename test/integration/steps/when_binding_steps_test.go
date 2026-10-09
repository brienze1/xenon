package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/cucumber/godog"

	"github.com/brienze1/xenon/internal/infra/server"
)

func (t *testUtils) registerWhenBindingSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I call "(\w+)\.(\w+)" with$`, t.iCallWith)
	ctx.Step(`^I call "(\w+)\.(\w+)"$`, t.iCall)
	ctx.When(`^I call "(\w+)\.(\w+)" (\d+) times cycling through the payloads$`, t.iCallTimesCyclingThroughThePayloads)
	ctx.When(`^I call "(\w+)\.(\w+)" and cancel it after (\d+)ms with$`, t.iCallAndCancelItAfterWith)
	ctx.Step(`^I save the response field "([^"]*)" as "([^"]*)"$`, t.iSaveTheResponseFieldAs)
}

func (t *testUtils) iCallWith(service, method string, payload *godog.DocString) error {
	return t.call(context.Background(), service, method, payload.Content)
}

func (t *testUtils) iCall(service, method string) error {
	return t.call(context.Background(), service, method, "")
}

func (t *testUtils) iCallAndCancelItAfterWith(service, method string, milliseconds int, payload *godog.DocString) error {
	ctx, cancel := context.WithCancel(context.Background())
	timer := time.AfterFunc(time.Duration(milliseconds)*time.Millisecond, cancel)
	defer timer.Stop()
	defer cancel()
	return t.call(ctx, service, method, payload.Content)
}

func (t *testUtils) iCallTimesCyclingThroughThePayloads(service, method string, times int, payloads *godog.DocString) error {
	var cycle []json.RawMessage
	if err := json.Unmarshal([]byte(payloads.Content), &cycle); err != nil || len(cycle) == 0 {
		return fmt.Errorf("expected a non-empty JSON array of payloads: %v", err)
	}
	for call := 0; call < times; call++ {
		if err := t.call(context.Background(), service, method, string(cycle[call%len(cycle)])); err != nil {
			return err
		}
		if t.callError != nil {
			return fmt.Errorf("call %d of %s.%s failed: %s", call+1, service, method, t.callError)
		}
	}
	return nil
}

func (t *testUtils) iSaveTheResponseFieldAs(path, name string) error {
	value, err := fieldOf(t.response, path)
	if err != nil {
		return err
	}
	t.vars[name] = stringify(value)
	return nil
}

func (t *testUtils) call(ctx context.Context, service, method, payload string) error {
	t.response, t.callError = nil, nil
	target, err := t.service(service)
	if err != nil {
		return err
	}
	function := target.MethodByName(method)
	if !function.IsValid() {
		return fmt.Errorf("binding %s has no method %s", service, method)
	}
	arguments := []reflect.Value{reflect.ValueOf(ctx)}
	if function.Type().NumIn() == 2 {
		input := reflect.New(function.Type().In(1))
		if payload != "" {
			if err := json.Unmarshal([]byte(t.substituteJson(payload)), input.Interface()); err != nil {
				return fmt.Errorf("invalid payload for %s.%s: %w", service, method, err)
			}
		}
		arguments = append(arguments, input.Elem())
	}
	started := time.Now()
	results := function.Call(arguments)
	t.elapsed = time.Since(started)
	if failure, isError := results[len(results)-1].Interface().(error); isError && failure != nil {
		t.callError = server.MarshalError(failure)
		return nil
	}
	if len(results) == 2 {
		t.response, err = json.Marshal(results[0].Interface())
	}
	return err
}

func (t *testUtils) service(name string) (reflect.Value, error) {
	bindings := reflect.ValueOf(t.bindings).Elem()
	for index := 0; index < bindings.NumField(); index++ {
		field := bindings.Field(index)
		if field.Type().Elem().Name() == name {
			return field, nil
		}
	}
	return reflect.Value{}, fmt.Errorf("no binding named %s", name)
}
