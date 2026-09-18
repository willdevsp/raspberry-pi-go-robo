package unit

import (
	"bufio"
	"bytes"
	"roverpi/internal/camera"
	"testing"
)

func TestScanJPEG(t *testing.T) {
	// Create a dummy MJPEG stream
	stream := []byte{
		0x00, 0x01, // random bytes
		0xFF, 0xD8, // SOI
		0x10, 0x11, 0x12, // image data
		0xFF, 0xD9, // EOI
		0x02, 0x03, // random bytes
		0xFF, 0xD8, // SOI 2
		0x20, 0x21, // image data 2
		0xFF, 0xD9, // EOI 2
	}

	scanner := bufio.NewScanner(bytes.NewReader(stream))
	scanner.Split(camera.ScanJPEG)

	// First frame
	if !scanner.Scan() {
		t.Fatal("expected first frame")
	}
	frame1 := scanner.Bytes()
	if !bytes.Equal(frame1, []byte{0xFF, 0xD8, 0x10, 0x11, 0x12, 0xFF, 0xD9}) {
		t.Fatalf("unexpected frame1: %x", frame1)
	}

	// Second frame
	if !scanner.Scan() {
		t.Fatal("expected second frame")
	}
	frame2 := scanner.Bytes()
	if !bytes.Equal(frame2, []byte{0xFF, 0xD8, 0x20, 0x21, 0xFF, 0xD9}) {
		t.Fatalf("unexpected frame2: %x", frame2)
	}

	if scanner.Scan() {
		t.Fatal("expected no more frames")
	}
}
