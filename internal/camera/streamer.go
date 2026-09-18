package camera

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os/exec"
)

type PiStreamer struct {
	cmd     *exec.Cmd
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

// Ensure PiStreamer implements Streamer
var _ Streamer = &PiStreamer{}

func NewPiStreamer() *PiStreamer {
	return &PiStreamer{}
}

func (s *PiStreamer) Start() error {
	s.cmd = exec.Command("rpicam-vid", "-t", "0", "--inline", "--width", "640", "--height", "480", "--framerate", "30", "--codec", "mjpeg", "-o", "-")
	
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	s.stdout = stdout

	if err := s.cmd.Start(); err != nil {
		return err
	}

	s.scanner = bufio.NewScanner(s.stdout)
	// 5MB max frame size
	buf := make([]byte, 0, 1024*1024)
	s.scanner.Buffer(buf, 5*1024*1024)
	
	s.scanner.Split(ScanJPEG)

	return nil
}

func (s *PiStreamer) Stop() error {
	if s.cmd != nil && s.cmd.Process != nil {
		err := s.cmd.Process.Kill()
		_ = s.cmd.Wait() // cleanup
		return err
	}
	return nil
}

// ScanJPEG is a split function for bufio.Scanner to extract JPEG frames.
func ScanJPEG(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	
	// Find Start of Image (SOI) marker FF D8
	soi := bytes.Index(data, []byte{0xFF, 0xD8})
	if soi == -1 {
		// Not found, discard all data if at EOF, otherwise keep last byte just in case it's 0xFF
		if atEOF {
			return len(data), nil, nil
		}
		return len(data) - 1, nil, nil
	}

	// Find End of Image (EOI) marker FF D9
	// We start looking from the SOI marker + 2
	eoi := bytes.Index(data[soi+2:], []byte{0xFF, 0xD9})
	if eoi == -1 {
		// Not found, need more data
		if atEOF {
			return len(data), nil, io.ErrUnexpectedEOF
		}
		return soi, nil, nil // advance to SOI
	}

	// We found a complete frame
	frameLen := eoi + 2
	return soi + 2 + frameLen, data[soi : soi+2+frameLen], nil
}

// ReadFrame parses MJPEG from stdout using bytes.Index
func (s *PiStreamer) ReadFrame() ([]byte, error) {
	if s.scanner == nil {
		return nil, errors.New("streamer not started")
	}

	if s.scanner.Scan() {
		return s.scanner.Bytes(), nil
	}
	
	if err := s.scanner.Err(); err != nil {
		return nil, err
	}
	
	return nil, io.EOF
}
