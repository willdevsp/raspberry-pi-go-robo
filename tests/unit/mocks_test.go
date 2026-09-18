package unit

import (
	"io"
	"roverpi/internal/camera"
	"roverpi/tests/mocks"
	"testing"
)

func TestCameraMock(t *testing.T) {
	var streamer camera.Streamer = &mocks.CameraMock{}

	err := streamer.Start()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	frame, err := streamer.ReadFrame()
	if err != nil {
		t.Fatalf("expected no error reading frame, got %v", err)
	}
	if len(frame) == 0 {
		t.Fatal("expected dummy frame, got empty byte slice")
	}

	err = streamer.Stop()
	if err != nil {
		t.Fatalf("expected no error stopping, got %v", err)
	}
}

func TestSerialMock(t *testing.T) {
	var serial io.ReadWriteCloser = &mocks.SerialMock{}

	_, err := serial.Write([]byte("W"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	buf := make([]byte, 1)
	_, err = serial.Read(buf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if buf[0] != 'W' {
		t.Fatalf("expected 'W', got %c", buf[0])
	}

	err = serial.Close()
	if err != nil {
		t.Fatalf("expected no error closing, got %v", err)
	}
}
