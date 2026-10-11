package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/brienze1/xenon/internal/infra/server/errorpayload"
)

func (t *testUtils) registerWhenBindingSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I call "(\w+)\.(\w+)" with$`, t.iCallWith)
	ctx.Step(`^I call "(\w+)\.(\w+)"$`, t.iCall)
	ctx.Step(`^I call "(\w+)\.(\w+)" with resource fields$`, t.iCallWithResourceFields)
	ctx.When(`^I call "(\w+)\.(\w+)" (\d+) times cycling through the payloads$`, t.iCallTimesCyclingThroughThePayloads)
	ctx.When(`^I call "(\w+)\.(\w+)" and cancel it after (\d+)ms with$`, t.iCallAndCancelItAfterWith)
	ctx.When(`^I call "(\w+)\.(\w+)" already cancelled with$`, t.iCallAlreadyCancelledWith)
	ctx.Step(`^I save the response field "([^"]*)" as "([^"]*)"$`, t.iSaveTheResponseFieldAs)
	ctx.Step(`^I save every "([^"]*)" found in the response as "([^"]*)"$`, t.iSaveEveryFoundInTheResponseAs)
	ctx.When(`^I start calling "(\w+)\.(\w+)" in the background with$`, t.iStartCallingInTheBackgroundWith)
	ctx.Then(`^the background call should finish within (\d+) ms$`, t.theBackgroundCallShouldFinishWithin)
}

type backgroundCall struct {
	done    chan struct{}
	started time.Time
	result  callResult
	err     error
}

func (t *testUtils) iStartCallingInTheBackgroundWith(service, method string, payload *godog.DocString) error {
	invoke, err := t.prepareCall(context.Background(), service, method, payload.Content)
	if err != nil {
		return err
	}
	call := &backgroundCall{done: make(chan struct{}), started: time.Now()}
	t.background = call
	go func() {
		call.result, call.err = invoke()
		close(call.done)
	}()
	return nil
}

func (t *testUtils) theBackgroundCallShouldFinishWithin(milliseconds int) error {
	call := t.background
	if call == nil {
		return fmt.Errorf("no call was started in the background")
	}
	limit := time.Duration(milliseconds) * time.Millisecond
	select {
	case <-call.done:
	case <-time.After(time.Until(call.started.Add(limit))):
		return fmt.Errorf("the background call did not finish within %s", limit)
	}
	t.response, t.callError, t.elapsed = call.result.response, call.result.callError, time.Since(call.started)
	if t.elapsed > limit {
		return fmt.Errorf("the background call took %s, expected at most %s", t.elapsed, limit)
	}
	return call.err
}

func (t *testUtils) iCallWith(service, method string, payload *godog.DocString) error {
	return t.call(context.Background(), service, method, payload.Content)
}

var resourceField = regexp.MustCompile(`\{\{resource:([^#}]+)#([^}]*)\}\}`)

func (t *testUtils) iCallWithResourceFields(service, method string, payload *godog.DocString) error {
	var failure error
	expanded := resourceField.ReplaceAllStringFunc(payload.Content, func(match string) string {
		parts := resourceField.FindStringSubmatch(match)
		value, err := resourceFileField(parts[1], parts[2])
		if err != nil {
			failure = err
			return match
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			failure = err
		}
		return string(encoded)
	})
	if failure != nil {
		return failure
	}
	return t.call(context.Background(), service, method, expanded)
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

func (t *testUtils) iCallAlreadyCancelledWith(service, method string, payload *godog.DocString) error {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
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

func (t *testUtils) iSaveEveryFoundInTheResponseAs(key, name string) error {
	data, err := decodeJson(t.response)
	if err != nil {
		return err
	}
	t.vars[name] = strings.Join(collectValues(data, key), ",")
	return nil
}

func collectValues(data any, key string) []string {
	values := make([]string, 0)
	switch node := data.(type) {
	case map[string]any:
		for field, value := range node {
			if field == key {
				values = append(values, stringify(value))
			}
			values = append(values, collectValues(value, key)...)
		}
	case []any:
		for _, item := range node {
			values = append(values, collectValues(item, key)...)
		}
	}
	return values
}

type callResult struct {
	response  []byte
	callError []byte
	elapsed   time.Duration
}

func (t *testUtils) call(ctx context.Context, service, method, payload string) error {
	t.response, t.callError = nil, nil
	invoke, err := t.prepareCall(ctx, service, method, payload)
	if err != nil {
		return err
	}
	result, err := invoke()
	t.response, t.callError, t.elapsed = result.response, result.callError, result.elapsed
	return err
}

func (t *testUtils) prepareCall(ctx context.Context, service, method, payload string) (func() (callResult, error), error) {
	target, err := t.service(service)
	if err != nil {
		return nil, err
	}
	function := target.MethodByName(method)
	if !function.IsValid() {
		return nil, fmt.Errorf("binding %s has no method %s", service, method)
	}
	arguments := []reflect.Value{reflect.ValueOf(ctx)}
	if function.Type().NumIn() == 2 {
		input := reflect.New(function.Type().In(1))
		if payload != "" {
			if err := json.Unmarshal([]byte(t.substituteJson(payload)), input.Interface()); err != nil {
				return nil, fmt.Errorf("invalid payload for %s.%s: %w", service, method, err)
			}
		}
		arguments = append(arguments, input.Elem())
	}
	return func() (callResult, error) {
		started := time.Now()
		results := function.Call(arguments)
		result := callResult{elapsed: time.Since(started)}
		if failure, isError := results[len(results)-1].Interface().(error); isError && failure != nil {
			result.callError = errorpayload.MarshalError(failure)
			return result, nil
		}
		var err error
		if len(results) == 2 {
			result.response, err = json.Marshal(results[0].Interface())
		}
		return result, err
	}, nil
}

func (t *testUtils) service(name string) (reflect.Value, error) {
	for _, candidate := range t.bindings {
		target := reflect.ValueOf(candidate)
		if target.Type().Elem().Name() == name {
			return target, nil
		}
	}
	return reflect.Value{}, fmt.Errorf("no binding named %s", name)
}
