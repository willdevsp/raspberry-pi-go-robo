package mocks

import (
	"errors"
	"roverpi/internal/camera"
)

// CameraMock implements the camera.Streamer interface for testing.
type CameraMock struct {
	IsRunning   bool
	Frames      [][]byte
	frameIndex  int
	ReturnError error
}

// Ensure CameraMock implements camera.Streamer
var _ camera.Streamer = &CameraMock{}

func (m *CameraMock) Start() error {
	if m.ReturnError != nil {
		return m.ReturnError
	}
	m.IsRunning = true
	m.frameIndex = 0
	return nil
}

func (m *CameraMock) Stop() error {
	m.IsRunning = false
	return nil
}

func (m *CameraMock) ReadFrame() ([]byte, error) {
	if !m.IsRunning {
		return nil, errors.New("camera not running")
	}
	if len(m.Frames) == 0 {
		// return a dummy frame if none configured
		return []byte{0xFF, 0xD8, 0x00, 0x00, 0xFF, 0xD9}, nil
	}
	frame := m.Frames[m.frameIndex%len(m.Frames)]
	m.frameIndex++
	return frame, nil
}
