package tests_test

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bencbradshaw/framework/events"
)

// event represents a parsed SSE event
type sseEvent struct {
	ID    string
	Event string
	Data  string
}

// readSSEEvents reads events from an SSE stream.
// It sends events to the provided channel and closes it when the stream ends or times out.
func readSSEEvents(t *testing.T, resp *http.Response, eventChan chan<- sseEvent, stopSignal <-chan struct{}) {
	reader := bufio.NewReader(resp.Body)
	var currentEvent sseEvent

	defer func() {
		// Ensure channel is closed on any exit from this goroutine
		// This helps prevent deadlocks in the main test goroutine if reading stops unexpectedly
		// For example, if resp.Body is closed by the main test goroutine's defer resp.Body.Close()
		// while this reader is still active.
		if r := recover(); r != nil {
			t.Logf("Recovered in readSSEEvents: %v", r)
		}
		close(eventChan)
		t.Log("readSSEEvents: Event channel closed.")
	}()


	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// If there's an error (like EOF or closed connection), log and exit loop
			if err.Error() != "EOF" && !strings.Contains(err.Error(), "use of closed network connection") && !strings.Contains(err.Error(), "http: Server closed") {
				t.Logf("Error reading SSE stream: %v (%T)", err, err)
			} else {
				t.Logf("SSE stream closed or EOF: %v", err)
			}
			return // This will trigger the defer to close eventChan
		}

		line = strings.TrimSpace(line)
		t.Logf("Read line: '%s'", line)


		select {
		case <-stopSignal:
			t.Log("Stop signal received, closing event reader.")
			return // This will trigger the defer to close eventChan
		default:
		}

		if line == "" { // Empty line signifies end of an event
			if currentEvent.Event != "" || currentEvent.Data != "" || currentEvent.ID != "" { // Dispatch if not empty
				t.Logf("Dispatching event: %+v", currentEvent)
				// Non-blocking send to prevent deadlock if main goroutine is not ready
				select {
				case eventChan <- currentEvent:
				case <-time.After(500 * time.Millisecond): // Timeout for sending to channel
					t.Logf("Timeout sending event to channel: %+v. Channel might be full or receiver blocked.", currentEvent)
				case <-stopSignal: // Check stopSignal again before blocking on send
					t.Log("Stop signal received while trying to dispatch event.")
					return
				}
				currentEvent = sseEvent{} // Reset for next event
			}
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		field := parts[0]
		value := ""
		if len(parts) > 1 {
			value = strings.TrimSpace(parts[1])
		}

		switch field {
		case "event":
			currentEvent.Event = value
		case "data":
			// If data is already present, this is a multi-line data field
			if currentEvent.Data != "" {
				currentEvent.Data += "\n" + value
			} else {
				currentEvent.Data = value
			}
		case "id":
			currentEvent.ID = value
		// Ignoring retry and comments (lines starting with ':')
		default:
			t.Logf("Ignoring SSE line with field: %s", field)
		}
	}
}

func TestEventStream_ConnectionAndInitialEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(events.EventStream))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to connect to event stream: %v", err)
	}
	// Note: resp.Body.Close() will be called by the test function's defer,
	// which might interrupt the readSSEEvents goroutine. The goroutine should handle this.

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %s", resp.Status)
	}
	if contentType := resp.Header.Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", contentType)
	}

	receivedEvents := make(chan sseEvent, 10) // Buffered channel
	stopSignal := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer resp.Body.Close() // Ensure body is closed by the reader goroutine when it's done.
		readSSEEvents(t, resp, receivedEvents, stopSignal)
	}()

	var connectedEvent sseEvent
	select {
	case evt, ok := <-receivedEvents:
		if !ok {
			t.Fatal("Event channel closed unexpectedly before receiving connected event")
		}
		connectedEvent = evt
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for connected event")
	}

	// The events.go sends "event: connected\ndata: connected\n\n"
	// The reader should parse this as Event="connected", Data="connected"
	expectedEventName := "connected"
	expectedData := "connected"

	if connectedEvent.Event != expectedEventName {
		t.Errorf("Expected event name '%s', got '%s'", expectedEventName, connectedEvent.Event)
	}
	if connectedEvent.Data != expectedData {
		t.Errorf("Expected event data '%s', got '%s'", expectedData, connectedEvent.Data)
	}


	close(stopSignal) // Signal reader to stop
	wg.Wait()         // Wait for reader to finish
}


func TestEventStream_ReceivesMessageChanMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(events.EventStream))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to connect to event stream: %v", err)
	}

	receivedEvents := make(chan sseEvent, 10)
	stopSignal := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer resp.Body.Close()
		readSSEEvents(t, resp, receivedEvents, stopSignal)
	}()

	// Wait for "connected" event first
	select {
	case _, ok := <-receivedEvents:
		if !ok {
			t.Fatal("Event channel closed unexpectedly while waiting for connected event")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for connected event")
	}

	// Send a message on MessageChan
	testMessage := "Hello from MessageChan!"
	go func() {
		select {
		case events.MessageChan <- testMessage:
		case <-time.After(500 * time.Millisecond):
			t.Error("Timeout sending message to MessageChan") // Use t.Error to not stop the test
		}
	}()

	var dataEvent sseEvent
	select {
	case evt, ok := <-receivedEvents:
		if !ok {
			t.Fatal("Event channel closed unexpectedly while waiting for data event")
		}
		dataEvent = evt
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for data event from MessageChan")
	}

	// events.go sends: fmt.Fprintf(w, "event: data \ndata: %s\n\n", msg)
	// Note the trailing space in "data " for the event name.
	// current readSSEEvents will parse "data " as is.
	if dataEvent.Event != "data " { // Explicitly check for "data " due to server format
		t.Errorf("Expected event name 'data ', got '%s'", dataEvent.Event)
	}
	if dataEvent.Data != testMessage {
		t.Errorf("Expected data '%s', got '%s'", testMessage, dataEvent.Data)
	}

	close(stopSignal)
	wg.Wait()
}


func TestEventStream_EmitEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(events.EventStream))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to connect to event stream: %v", err)
	}

	receivedEvents := make(chan sseEvent, 10)
	stopSignal := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer resp.Body.Close()
		readSSEEvents(t, resp, receivedEvents, stopSignal)
	}()

	// Wait for "connected" event
	select {
	case <-receivedEvents:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for connected event")
	}

	type CustomData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	eventPayload := CustomData{Name: "TestEmit", Value: 123}
	eventName := "custom_event" // This will have a space appended by server: "custom_event "

	go events.EmitEvent(eventName, eventPayload)


	var emittedEvent sseEvent
	select {
	case evt, ok := <-receivedEvents:
		if !ok {
			t.Fatal("Event channel closed unexpectedly while waiting for emitted event")
		}
		emittedEvent = evt
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for emitted event")
	}

	expectedJSON, _ := json.Marshal(eventPayload)
	// events.go sends: fmt.Fprintf(w, "event: %s \ndata: %s\n\n", event, string(jsonData))
	// Note the trailing space in "event: %s "
	if emittedEvent.Event != eventName+" " {
		t.Errorf("Expected event name '%s ', got '%s'", eventName, emittedEvent.Event)
	}
	if emittedEvent.Data != string(expectedJSON) {
		t.Errorf("Expected event data '%s', got '%s'", string(expectedJSON), emittedEvent.Data)
	}

	// Test EmitEvent with nil data
	nilDataEventName := "nil_data_event" // Will become "nil_data_event "
	go events.EmitEvent(nilDataEventName, nil)


	var nilDataEmittedEvent sseEvent
	select {
	case evt, ok := <-receivedEvents:
		if !ok {
			t.Fatal("Event channel closed unexpectedly while waiting for nil data emitted event")
		}
		nilDataEmittedEvent = evt
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for nil data emitted event")
	}
	// events.go sends: fmt.Fprintf(w, "event: %s \n\n", event) for nil data
	if nilDataEmittedEvent.Event != nilDataEventName+" " {
		t.Errorf("Expected event name '%s ' for nil data, got '%s'", nilDataEventName, nilDataEmittedEvent.Event)
	}
	// When data is nil, server sends no "data:" line. Reader should produce empty Data.
	if nilDataEmittedEvent.Data != "" {
		t.Errorf("Expected event data to be empty for nil data, got '%s'", nilDataEmittedEvent.Data)
	}


	close(stopSignal)
	wg.Wait()
}

func TestEventStream_TimeEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(events.EventStream))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to connect to event stream: %v", err)
	}

	receivedEvents := make(chan sseEvent, 1)
	stopSignal := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer resp.Body.Close()
		readSSEEvents(t, resp, receivedEvents, stopSignal)
	}()

	select {
	case evt, ok := <-receivedEvents:
		if !ok {
			t.Fatal("Event channel closed unexpectedly")
		}
		t.Logf("Received initial event: %+v. Assuming time events would follow.", evt)
		// events.go sends "event: connected\ndata: connected\n\n"
		if evt.Event != "connected" { // This is the event name from the stream
			t.Errorf("Expected first event to be 'connected', got '%s'", evt.Event)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for any event (expecting 'connected')")
	}

	t.Log("TestEventStream_TimeEvent: Verified initial connection. Actual time event not checked due to long interval.")

	close(stopSignal)
	wg.Wait()
}
