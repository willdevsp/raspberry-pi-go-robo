package mocks

import (
	"bytes"
	"io"
)

// SerialMock implements io.ReadWriteCloser for testing.
type SerialMock struct {
	Buffer      bytes.Buffer
	Closed      bool
	ReturnError error
}

// Ensure SerialMock implements io.ReadWriteCloser
var _ io.ReadWriteCloser = &SerialMock{}

func (m *SerialMock) Read(p []byte) (n int, err error) {
	if m.ReturnError != nil {
		return 0, m.ReturnError
	}
	if m.Closed {
		return 0, io.EOF
	}
	return m.Buffer.Read(p)
}

func (m *SerialMock) Write(p []byte) (n int, err error) {
	if m.ReturnError != nil {
		return 0, m.ReturnError
	}
	if m.Closed {
		return 0, io.ErrClosedPipe
	}
	return m.Buffer.Write(p)
}

func (m *SerialMock) Close() error {
	m.Closed = true
	return nil
}
