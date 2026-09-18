package controller

import (
	"io"
	"sync"
	"time"

	"go.bug.st/serial"
)

// SerialController manages the connection and commands to the Arduino via Serial.
type SerialController struct {
	port io.ReadWriteCloser
	mu   sync.Mutex
}

// NewSerialController initializes a real serial port.
func NewSerialController(portName string, baudRate int) (*SerialController, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	// Give Arduino time to reset if necessary
	time.Sleep(2 * time.Second)

	return &SerialController{
		port: port,
	}, nil
}

// NewSerialControllerWithPort allows injecting a mock port.
func NewSerialControllerWithPort(port io.ReadWriteCloser) *SerialController {
	return &SerialController{
		port: port,
	}
}

// WriteCommand sends a single byte command to the serial port.
func (c *SerialController) WriteCommand(cmd byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.port.Write([]byte{cmd})
	return err
}

// Close closes the serial port connection.
func (c *SerialController) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.port.Close()
}
