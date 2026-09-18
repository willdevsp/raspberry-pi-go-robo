package unit

import (
	"roverpi/internal/controller"
	"roverpi/tests/mocks"
	"testing"
)

func TestSerialController(t *testing.T) {
	mockPort := &mocks.SerialMock{}
	ctrl := controller.NewSerialControllerWithPort(mockPort)

	err := ctrl.WriteCommand('W')
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockPort.Buffer.String() != "W" {
		t.Fatalf("expected 'W' in mock buffer, got %q", mockPort.Buffer.String())
	}

	err = ctrl.Close()
	if err != nil {
		t.Fatalf("expected no error closing, got %v", err)
	}
	if !mockPort.Closed {
		t.Fatal("expected mock port to be closed")
	}
}
