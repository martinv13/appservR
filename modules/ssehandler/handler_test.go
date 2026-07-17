package ssehandler

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

const shortWait = 200 * time.Millisecond

func init() {
	gin.SetMode(gin.TestMode)
}

func TestControllerStreamsBroadcastMessages(t *testing.T) {
	broker := NewMessageBroker()

	router := gin.New()
	router.GET("/events", broker.Controller())
	server := httptest.NewServer(router)
	// Not deferring server.Close(): Controller()'s Stream loop blocks reading
	// clientChan and only re-checks for disconnect after the next broadcast,
	// by which time this client is already removed from TotalClients — so it
	// can never unblock on its own. A graceful server.Close() would hang
	// forever waiting for that goroutine; the listener is reclaimed when the
	// test binary exits.

	resp, err := http.Get(server.URL + "/events")
	if err != nil {
		t.Fatalf("failed to connect to SSE endpoint: %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %q", ct)
	}

	reader := bufio.NewReader(resp.Body)

	// Give the handler a moment to register as a client before broadcasting.
	time.Sleep(20 * time.Millisecond)
	broker.Message <- "streamed message"

	found := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.Contains(line, "streamed message") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to read broadcast message from SSE stream")
	}
}

func TestNewClientReceivesCurrentLastMessage(t *testing.T) {
	broker := NewMessageBroker()

	client := make(chan string)
	broker.NewClients <- client

	select {
	case msg := <-client:
		if msg != "" {
			t.Errorf("expected empty initial last message, got %q", msg)
		}
	case <-time.After(shortWait):
		t.Fatal("timed out waiting for initial message")
	}
}

func TestBroadcastsNewMessageToClient(t *testing.T) {
	broker := NewMessageBroker()

	client := make(chan string)
	broker.NewClients <- client
	<-client // drain initial (empty) last message

	broker.Message <- "hello"

	select {
	case msg := <-client:
		if msg != "hello" {
			t.Errorf("expected %q, got %q", "hello", msg)
		}
	case <-time.After(shortWait):
		t.Fatal("timed out waiting for broadcast message")
	}
}

func TestNewClientReceivesPreviouslyBroadcastMessage(t *testing.T) {
	broker := NewMessageBroker()

	// A message is broadcast before any client subscribes.
	firstClient := make(chan string)
	broker.NewClients <- firstClient
	<-firstClient

	broker.Message <- "prior message"
	if got := <-firstClient; got != "prior message" {
		t.Fatalf("setup: expected first client to receive %q, got %q", "prior message", got)
	}

	// A late-joining client should immediately get the last broadcast message.
	lateClient := make(chan string)
	broker.NewClients <- lateClient

	select {
	case msg := <-lateClient:
		if msg != "prior message" {
			t.Errorf("expected late client to receive last message %q, got %q", "prior message", msg)
		}
	case <-time.After(shortWait):
		t.Fatal("timed out waiting for late client's initial message")
	}
}

func TestDuplicateMessageIsNotRebroadcast(t *testing.T) {
	broker := NewMessageBroker()

	client := make(chan string)
	broker.NewClients <- client
	<-client // drain initial empty message

	broker.Message <- "same"
	if got := <-client; got != "same" {
		t.Fatalf("setup: expected %q, got %q", "same", got)
	}

	// Sending the identical message again should not produce a second delivery.
	broker.Message <- "same"

	select {
	case msg := <-client:
		t.Fatalf("expected no rebroadcast of identical message, got %q", msg)
	case <-time.After(shortWait):
		// expected: no message delivered
	}
}

func TestClosedClientStopsReceivingMessages(t *testing.T) {
	broker := NewMessageBroker()

	client := make(chan string)
	broker.NewClients <- client
	<-client // drain initial empty message

	broker.ClosedClients <- client

	// Give the broker's listen loop a moment to process the removal before
	// broadcasting, since channel send/receive only guarantees the value
	// transfer happened, not that the case body has finished executing.
	time.Sleep(20 * time.Millisecond)

	broker.Message <- "after close"

	select {
	case msg := <-client:
		t.Fatalf("expected closed client to not receive further messages, got %q", msg)
	case <-time.After(shortWait):
		// expected: no message delivered to the closed client
	}
}

func TestMultipleClientsAllReceiveBroadcast(t *testing.T) {
	broker := NewMessageBroker()

	clientA := make(chan string)
	clientB := make(chan string)
	broker.NewClients <- clientA
	<-clientA
	broker.NewClients <- clientB
	<-clientB

	broker.Message <- "fan-out"

	for name, c := range map[string]chan string{"A": clientA, "B": clientB} {
		select {
		case msg := <-c:
			if msg != "fan-out" {
				t.Errorf("client %s: expected %q, got %q", name, "fan-out", msg)
			}
		case <-time.After(shortWait):
			t.Fatalf("client %s: timed out waiting for broadcast", name)
		}
	}
}
