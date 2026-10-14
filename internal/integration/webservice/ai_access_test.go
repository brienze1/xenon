package webservice

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type fakeListener struct {
	busy      bool
	listening bool
	clients   int
	closed    int
}

func (f *fakeListener) Listen(int) error {
	if f.busy {
		return errors.New("address already in use")
	}
	f.listening = true
	return nil
}

func (f *fakeListener) Close() {
	f.listening = false
	f.closed++
}

func (f *fakeListener) Clients() int { return f.clients }

func TestAIAccessRunsAndCountsClients(t *testing.T) {
	listener := &fakeListener{}
	webservice := NewAIAccessWebservice(listener)
	status := webservice.Apply(context.Background(), entity.AIAccessSettings{Enabled: true, Port: 7777})
	if !status.IsRunning() || status.Port != 7777 || !listener.listening {
		t.Fatalf("Apply() = %+v", status)
	}
	listener.clients = 1
	if got := webservice.Status(context.Background()); got.Clients != 1 {
		t.Fatalf("Status() = %+v", got)
	}
	if stopped := webservice.Stop(context.Background()); stopped.State != "off" || stopped.Port != 7777 || stopped.Clients != 0 || listener.listening {
		t.Fatalf("Stop() = %+v", stopped)
	}
}

func TestAIAccessReportsABusyPortAndTurnsOff(t *testing.T) {
	listener := &fakeListener{busy: true}
	webservice := NewAIAccessWebservice(listener)
	status := webservice.Apply(context.Background(), entity.AIAccessSettings{Enabled: true, Port: 7777})
	if status.State != "error" || status.Code != entity.PortInUseCode || status.Reason != "port "+strconv.Itoa(7777)+" is in use" {
		t.Fatalf("busy port status = %+v", status)
	}
	if off := webservice.Disable(context.Background(), entity.AIAccessSettings{Enabled: false, Port: 7777}); off.State != "off" || off.Port != 7777 {
		t.Fatalf("disabled status = %+v", off)
	}
	if listener.closed != 2 {
		t.Fatalf("closed %d times, want 2", listener.closed)
	}
}
