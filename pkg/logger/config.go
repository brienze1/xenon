package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	logFileMaxBytes = 10 * 1024 * 1024
	logFileBackups  = 3
)

type config struct {
	output  io.Writer
	service string
	mu      sync.RWMutex
}

var settings = &config{output: os.Stdout, service: "xenon"}

func SetOutput(output io.Writer) {
	settings.mu.Lock()
	defer settings.mu.Unlock()
	settings.output = output
}

func SetServiceName(service string) {
	settings.mu.Lock()
	defer settings.mu.Unlock()
	settings.service = service
}

// SetFileSink sends logs to path, rotated at 10 MB keeping 3 backups; the caller closes the returned sink.
func SetFileSink(path string) (io.Closer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	sink, err := newRotatingFile(path, logFileMaxBytes, logFileBackups)
	if err != nil {
		return nil, err
	}
	SetOutput(sink)
	return sink, nil
}

func write(line []byte) {
	settings.mu.RLock()
	defer settings.mu.RUnlock()
	_, _ = settings.output.Write(line)
}

func serviceName() string {
	settings.mu.RLock()
	defer settings.mu.RUnlock()
	return settings.service
}
