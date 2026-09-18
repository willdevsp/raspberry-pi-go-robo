package camera

// Streamer defines the interface for capturing and providing MJPEG frames.
type Streamer interface {
	// Start begins capturing the video stream.
	Start() error
	// Stop ends the video capture.
	Stop() error
	// ReadFrame reads a single MJPEG frame.
	ReadFrame() ([]byte, error)
}
