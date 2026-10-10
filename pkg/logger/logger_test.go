package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type syncBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (s *syncBuffer) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.Write(data)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.String()
}

type secretRequest struct {
	Id    string
	Token string
}

func (s secretRequest) LogValue() any {
	return map[string]any{"id": s.Id}
}

type codedErr struct{}

func (codedErr) Error() string       { return "coded" }
func (codedErr) Code() string        { return "TST-00400" }
func (codedErr) Description() string { return "described" }
func (codedErr) StatusCode() int     { return 400 }

func capture(t *testing.T) *syncBuffer {
	t.Helper()
	buffer := &syncBuffer{}
	SetOutput(buffer)
	t.Cleanup(func() { SetOutput(os.Stdout) })
	return buffer
}

func decodeLines(t *testing.T, output string) []logPattern {
	t.Helper()
	var lines []logPattern
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var pattern logPattern
		if err := json.Unmarshal([]byte(line), &pattern); err != nil {
			t.Fatalf("invalid json line %q: %v", line, err)
		}
		lines = append(lines, pattern)
	}
	return lines
}

func TestInfoWritesJsonLine(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	buffer := capture(t)
	ctx := GetContext(context.Background())
	Info(ctx, "Started", map[string]any{"source": "test"})
	lines := decodeLines(t, buffer.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	line := lines[0]
	if line.Level != "INFO" || line.Detail != "Started" || line.Service != "xenon" {
		t.Errorf("unexpected line %+v", line)
	}
	if line.TransactionId != ctx.GetTransactionId() || line.RequestId != ctx.GetRequestId() || line.CorrelationId == "" {
		t.Errorf("ids not propagated: %+v", line)
	}
	if !strings.Contains(line.Operation, "TestInfoWritesJsonLine") {
		t.Errorf("operation = %q", line.Operation)
	}
}

func TestLevelFiltering(t *testing.T) {
	cases := []struct {
		level string
		want  []string
	}{
		{level: "ERROR", want: []string{"ERROR"}},
		{level: "WARN", want: []string{"ERROR", "WARN"}},
		{level: "", want: []string{"ERROR", "WARN", "INFO"}},
		{level: "debug", want: []string{"ERROR", "WARN", "INFO", "DEBUG"}},
		{level: "TRACE", want: []string{"ERROR", "WARN", "INFO", "DEBUG", "TRACE"}},
		{level: "OFF", want: nil},
	}
	for _, tc := range cases {
		t.Run(tc.level, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", tc.level)
			buffer := capture(t)
			wrapper := LoggerWrapper()
			wrapper.Error(context.Background(), errors.New("e"), "error")
			wrapper.Warn(context.Background(), errors.New("w"), "warn")
			wrapper.Info(context.Background(), "info")
			wrapper.Debug(context.Background(), "debug")
			wrapper.Trace(context.Background(), "trace")
			lines := decodeLines(t, buffer.String())
			if len(lines) != len(tc.want) {
				t.Fatalf("got %d lines, want %d", len(lines), len(tc.want))
			}
			for index, line := range lines {
				if line.Level != tc.want[index] {
					t.Errorf("line %d level = %s, want %s", index, line.Level, tc.want[index])
				}
			}
		})
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	t.Setenv("LOG_LEVEL", "TRACE")
	buffer := capture(t)
	Error(nil, codedErr{}, "error")
	Warn(context.Background(), nil, "warn")
	Debug(context.Background(), "debug")
	Trace(context.Background(), "trace")
	lines := decodeLines(t, buffer.String())
	if len(lines) != 4 {
		t.Fatalf("got %d lines", len(lines))
	}
	exception, ok := lines[0].Exceptions.(map[string]any)
	if !ok || exception["code"] != "TST-00400" || exception["description"] != "described" {
		t.Errorf("exceptions = %v", lines[0].Exceptions)
	}
	if lines[0].TransactionId != "" {
		t.Error("nil context should not carry ids")
	}
	if lines[1].Exceptions != nil {
		t.Error("nil error should not produce exceptions")
	}
}

func TestRedactableValuesAreLoggedThroughLogValue(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	buffer := capture(t)
	request := secretRequest{Id: "req-1", Token: "Bearer super-secret-token"}
	Info(context.Background(), "Started", request, &request, map[string]any{"nested": request, "plain": "ok"})
	output := buffer.String()
	if strings.Contains(output, "super-secret-token") {
		t.Fatalf("secret leaked into log: %s", output)
	}
	if !strings.Contains(output, "req-1") || !strings.Contains(output, `"plain":"ok"`) {
		t.Errorf("expected redacted values in log: %s", output)
	}
}

func TestSlicesOfRedactableValuesAreLoggedThroughLogValue(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	buffer := capture(t)
	requests := []secretRequest{{Id: "req-1", Token: "Bearer super-secret-token"}}
	Info(context.Background(), "Started", requests, map[string]any{"requests": requests, "bytes": []byte("ok"), "names": []string{"plain"}})
	output := buffer.String()
	if strings.Contains(output, "super-secret-token") {
		t.Fatalf("secret leaked into log: %s", output)
	}
	if !strings.Contains(output, "req-1") || !strings.Contains(output, `"names":["plain"]`) || !strings.Contains(output, `"bytes":"b2s="`) {
		t.Errorf("expected redacted values in log: %s", output)
	}
}

func TestUnserializableMetadata(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	buffer := capture(t)
	Info(context.Background(), "chan", make(chan int))
	if !strings.Contains(buffer.String(), "unserializable metadata") {
		t.Errorf("output = %s", buffer.String())
	}
}

func TestGetContextReusesIds(t *testing.T) {
	first := GetContext(context.Background())
	if GetContext(first) != first {
		t.Error("an existing Ctx should be returned as is")
	}
	derived, cancel := context.WithCancel(first)
	defer cancel()
	again := GetContext("ignored", derived)
	if again.GetTransactionId() != first.GetTransactionId() {
		t.Error("derived context should keep the transaction id")
	}
	again.SetCorrelationId("corr-1")
	if first.GetCorrelationId() != "corr-1" {
		t.Error("correlation id should be shared")
	}
	if GetContext().GetTransactionId() == "" {
		t.Error("empty call should create ids")
	}
}

func TestSetServiceName(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	buffer := capture(t)
	SetServiceName("other")
	t.Cleanup(func() { SetServiceName("xenon") })
	Info(context.Background(), "x")
	if decodeLines(t, buffer.String())[0].Service != "other" {
		t.Error("service name not applied")
	}
}

func TestSetFileSinkRotates(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	path := filepath.Join(t.TempDir(), "logs", "xenon.log")
	sink, err := SetFileSink(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { SetOutput(os.Stdout) })
	Info(context.Background(), "to file")
	if err := sink.Close(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(content), "to file") {
		t.Fatalf("file content %q, err %v", content, err)
	}

	rotating, err := newRotatingFile(path, 10, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, chunk := range []string{"aaaaaaaa\n", "bbbbbbbb\n", "cccccccc\n", "dddddddd\n"} {
		if _, err := rotating.Write([]byte(chunk)); err != nil {
			t.Fatal(err)
		}
	}
	if err := rotating.Close(); err != nil {
		t.Fatal(err)
	}
	assertFile(t, path, "dddddddd\n")
	assertFile(t, path+".1", "cccccccc\n")
	assertFile(t, path+".2", "bbbbbbbb\n")
	if _, err := os.Stat(path + ".3"); !os.IsNotExist(err) {
		t.Error("only 2 backups should be kept")
	}
}

func TestSetFileSinkFailsOnInvalidPath(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SetFileSink(filepath.Join(blocker, "logs", "xenon.log")); err == nil {
		t.Error("expected an error when the directory cannot be created")
	}
	if _, err := newRotatingFile(filepath.Join(blocker, "x.log"), 1, 1); err == nil {
		t.Error("expected an error when the file cannot be opened")
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != want {
		t.Errorf("%s = %q, want %q", filepath.Base(path), content, want)
	}
}

func TestWithRedactionsMasksValuesInEveryLine(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	buffer := capture(t)
	ctx := WithRedactions(GetContext(context.Background()), "s3cr3t", "", `quo"te<`)
	ctx = WithRedactions(ctx, "second")
	Info(ctx, "Started", map[string]any{"url": "https://api.test/s3cr3t", "header": `quo"te<`, "other": "second"})
	Info(context.Background(), "Unrelated", map[string]any{"value": "s3cr3t"})
	output := buffer.String()
	lines := decodeLines(t, output)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %s", len(lines), output)
	}
	first := strings.Split(strings.TrimSpace(output), "\n")[0]
	if strings.Contains(first, "s3cr3t") || strings.Contains(first, "quo") || strings.Contains(first, "second") {
		t.Fatalf("redacted values leaked: %s", first)
	}
	if strings.Count(first, "•••••") != 3 {
		t.Fatalf("expected three masks: %s", first)
	}
	if !strings.Contains(output, `"value":"s3cr3t"`) {
		t.Fatalf("a context without redactions must log as is: %s", output)
	}
}
