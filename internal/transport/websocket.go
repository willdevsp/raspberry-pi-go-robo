package transport

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local testing
	},
}

// Controller defines the interface needed to send commands to the hardware
type Controller interface {
	WriteCommand(cmd byte) error
}

// WebSocketHandler manages the rover's WebSocket connection
type WebSocketHandler struct {
	controller    Controller
	activeSession *websocket.Conn
	mu            sync.Mutex
}

func NewWebSocketHandler(c Controller) *WebSocketHandler {
	return &WebSocketHandler{
		controller: c,
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 4.1: Check PIN
	pin := r.URL.Query().Get("pin")
	if pin != "4321" {
		http.Error(w, "Unauthorized: Invalid PIN", http.StatusUnauthorized)
		return
	}

	h.mu.Lock()
	if h.activeSession != nil {
		h.mu.Unlock()
		http.Error(w, "Conflict: Another pilot is active", http.StatusConflict)
		return
	}
	h.mu.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	h.mu.Lock()
	h.activeSession = conn
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		h.activeSession = nil
		h.mu.Unlock()
		conn.Close()
		// Safety: stop rover on disconnect
		_ = h.controller.WriteCommand('X')
	}()

	// 4.4 Watchdog
	watchdog := time.NewTimer(400 * time.Millisecond)
	go func() {
		<-watchdog.C
		// Watchdog expired, stop rover
		_ = h.controller.WriteCommand('X')
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		// Reset watchdog
		if !watchdog.Stop() {
			// Drain the channel if stopped concurrently
			select {
			case <-watchdog.C:
			default:
			}
		}
		watchdog.Reset(400 * time.Millisecond)

		if len(msg) > 0 {
			cmd := msg[0]
			// Relaying valid commands
			switch cmd {
			case 'W', 'A', 'S', 'D', 'Q', 'E', 'X':
				_ = h.controller.WriteCommand(cmd)
			}
		}
	}
}
