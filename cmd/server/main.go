package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"roverpi/internal/camera"
	"roverpi/internal/controller"
	"roverpi/internal/transport"
	"roverpi/web"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	serialPort := flag.String("serial", "/dev/ttyUSB0", "Serial port for Arduino")
	flag.Parse()

	log.Println("Starting RoverPi server...")

	// 1. Initialize Serial Controller
	ctrl, err := controller.NewSerialController(*serialPort, 115200)
	if err != nil {
		log.Printf("Warning: Failed to open serial port %s: %v", *serialPort, err)
		log.Println("Continuing without serial hardware (dry run)")
	} else {
		defer ctrl.Close()
	}

	var activeCtrl transport.Controller
	if ctrl != nil {
		activeCtrl = ctrl
	} else {
		// Mock controller if serial fails so we can test UI
		activeCtrl = &dummyController{}
	}

	// 2. Initialize Camera Streamer
	streamer := camera.NewPiStreamer()
	if err := streamer.Start(); err != nil {
		log.Printf("Warning: Failed to start camera: %v", err)
	} else {
		defer streamer.Stop()
	}

	// 3. Setup WebSocket
	wsHandler := transport.NewWebSocketHandler(activeCtrl)

	// 4. Setup Routes
	mux := http.NewServeMux()

	// Static files (go:embed)
	mux.Handle("/", http.FileServer(http.FS(web.Assets)))

	// WebSocket endpoint
	mux.Handle("/ws", wsHandler)

	// Stream endpoint
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		for {
			frame, err := streamer.ReadFrame()
			if err != nil {
				return
			}
			fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame))
			w.Write(frame)
			w.Write([]byte("\r\n"))
		}
	})

	// 5. Start Server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server listening on http://0.0.0.0%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

type dummyController struct{}

func (d *dummyController) WriteCommand(cmd byte) error {
	log.Printf("Dummy serial write: %c", cmd)
	return nil
}
