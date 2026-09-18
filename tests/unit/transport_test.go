package unit

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"roverpi/internal/transport"

	"github.com/gorilla/websocket"
)

// MockController implements transport.Controller
type MockController struct {
	LastCmd byte
}

func (m *MockController) WriteCommand(cmd byte) error {
	m.LastCmd = cmd
	return nil
}

func TestWebSocketHandler_Unauthorized(t *testing.T) {
	mockCtrl := &MockController{}
	handler := transport.NewWebSocketHandler(mockCtrl)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?pin=wrong"
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected error dialing with wrong pin")
	}
}

func TestWebSocketHandler_SingleSession(t *testing.T) {
	mockCtrl := &MockController{}
	handler := transport.NewWebSocketHandler(mockCtrl)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?pin=4321"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect first client: %v", err)
	}
	defer conn1.Close()

	// Try to connect a second client
	_, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected error dialing second client (should be blocked)")
	}
}

func TestWebSocketHandler_CommandsAndWatchdog(t *testing.T) {
	mockCtrl := &MockController{}
	handler := transport.NewWebSocketHandler(mockCtrl)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?pin=4321"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Send a command
	err = conn.WriteMessage(websocket.TextMessage, []byte("W"))
	if err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Allow some time for processing
	time.Sleep(50 * time.Millisecond)

	if mockCtrl.LastCmd != 'W' {
		t.Fatalf("expected command 'W', got %c", mockCtrl.LastCmd)
	}

	// Wait for watchdog to trigger (>400ms)
	time.Sleep(500 * time.Millisecond)

	if mockCtrl.LastCmd != 'X' {
		t.Fatalf("expected watchdog to send 'X', got %c", mockCtrl.LastCmd)
	}

	conn.Close()
}
