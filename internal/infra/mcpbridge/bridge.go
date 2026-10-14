package mcpbridge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
)

const (
	xenonServerName   = "xenon"
	initializeMethod  = "initialize"
	drainTimeout      = 5 * time.Second
	internalErrorCode = -32603
)

var (
	errAccessOff   = errors.New("Xenon AI access is off — turn it on in Xenon → Settings → AI access (MCP).")
	errNotRunning  = errors.New("Xenon is not running — start Xenon with AI access on, then retry.")
	errXenonClosed = errors.New("Xenon closed the connection.")
)

func foreignServerError(port int) error {
	return fmt.Errorf("Port %d is in use by another program — change the port in Xenon → Settings → AI access (MCP).", port)
}

// Run bridges newline-delimited JSON-RPC on stdin/stdout to the running Xenon MCP server and returns the exit code.
func Run(stdin io.Reader, stdout, stderr io.Writer) int {
	if err := run(stdin, stdout); err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())
		return 1
	}
	return 0
}

func run(stdin io.Reader, stdout io.Writer) error {
	access := findAIAccess(context.Background())
	if !access.ShouldListen() {
		return errAccessOff
	}
	port := access.EffectivePort(properties.Properties().Application.MCPPortOverride)
	connection, err := net.DialTimeout("tcp", entity.AIAccessListenAddress(port), connectTimeout)
	if err != nil {
		return errNotRunning
	}
	_ = connection.Close()
	return newBridge(port, entity.AIAccessURL(port), stdout).pump(stdin)
}

type bridge struct {
	client   *xenonClient
	out      *output
	failure  chan error
	inflight sync.WaitGroup
	port     int
}

func newBridge(port int, endpoint string, stdout io.Writer) *bridge {
	return &bridge{
		port:    port,
		client:  newXenonClient(endpoint),
		out:     &output{writer: stdout},
		failure: make(chan error, 1),
	}
}

func (b *bridge) pump(stdin io.Reader) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer b.out.close()
	lines := readLines(ctx, stdin)
	for {
		select {
		case err := <-b.failure:
			return err
		case line, open := <-lines:
			if !open {
				return b.drain()
			}
			if err := b.forward(ctx, line); err != nil {
				return err
			}
		}
	}
}

func (b *bridge) forward(ctx context.Context, line []byte) error {
	message, err := jsonrpc.DecodeMessage(line)
	request, isRequest := message.(*jsonrpc.Request)
	if err == nil && isRequest && request.IsCall() && request.Method == initializeMethod {
		return b.initialize(ctx, line, request)
	}
	b.inflight.Add(1)
	go func() {
		defer b.inflight.Done()
		if err := b.relay(ctx, line, request); err != nil {
			b.fail(err)
		}
	}()
	return nil
}

func (b *bridge) initialize(ctx context.Context, line []byte, request *jsonrpc.Request) error {
	var received []jsonrpc.Message
	err := b.client.send(ctx, line, func(message jsonrpc.Message) {
		received = append(received, message)
	})
	if err != nil {
		return foreignServerError(b.port)
	}
	result, answered := initializeResultFor(received, request.ID)
	if !answered || result.ServerInfo.Name != xenonServerName {
		return foreignServerError(b.port)
	}
	b.client.setProtocolVersion(result.ProtocolVersion)
	for _, message := range received {
		if err := b.out.write(message); err != nil {
			return err
		}
	}
	return nil
}

func (b *bridge) relay(ctx context.Context, line []byte, request *jsonrpc.Request) error {
	isCall := request != nil && request.IsCall()
	answered := false
	var writeErr error
	err := b.client.send(ctx, line, func(message jsonrpc.Message) {
		if response, isResponse := message.(*jsonrpc.Response); isCall && isResponse && response.ID == request.ID {
			answered = true
		}
		if err := b.out.write(message); err != nil && writeErr == nil {
			writeErr = err
		}
	})
	var status *statusError
	switch {
	case writeErr != nil:
		return writeErr
	case ctx.Err() != nil:
		return nil
	case errors.As(err, &status):
		if isCall {
			return b.out.write(&jsonrpc.Response{ID: request.ID, Error: &jsonrpc.Error{Code: internalErrorCode, Message: status.Error()}})
		}
		return nil
	case err != nil:
		return err
	case isCall && !answered:
		return errXenonClosed
	}
	return nil
}

func (b *bridge) drain() error {
	finished := make(chan struct{})
	go func() {
		b.inflight.Wait()
		close(finished)
	}()
	timer := time.NewTimer(drainTimeout)
	defer timer.Stop()
	select {
	case err := <-b.failure:
		return err
	case <-timer.C:
	case <-finished:
	}
	select {
	case err := <-b.failure:
		return err
	default:
	}
	b.client.closeSession()
	return nil
}

func (b *bridge) fail(err error) {
	select {
	case b.failure <- err:
	default:
	}
}

type initializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	ServerInfo      struct {
		Name string `json:"name"`
	} `json:"serverInfo"`
}

func initializeResultFor(messages []jsonrpc.Message, id jsonrpc.ID) (initializeResult, bool) {
	for _, message := range messages {
		response, isResponse := message.(*jsonrpc.Response)
		if !isResponse || response.ID != id || response.Error != nil {
			continue
		}
		var result initializeResult
		if err := json.Unmarshal(response.Result, &result); err != nil {
			return initializeResult{}, false
		}
		return result, true
	}
	return initializeResult{}, false
}

func readLines(ctx context.Context, stdin io.Reader) <-chan []byte {
	lines := make(chan []byte)
	go func() {
		defer close(lines)
		reader := bufio.NewReader(stdin)
		for {
			line, err := reader.ReadBytes('\n')
			if trimmed := bytes.TrimSpace(line); len(trimmed) > 0 {
				select {
				case lines <- trimmed:
				case <-ctx.Done():
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	return lines
}

type output struct {
	writer io.Writer
	mu     sync.Mutex
	closed bool
}

func (o *output) write(message jsonrpc.Message) error {
	data, err := jsonrpc.EncodeMessage(message)
	if err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return nil
	}
	_, err = o.writer.Write(append(data, '\n'))
	return err
}

func (o *output) close() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.closed = true
}
