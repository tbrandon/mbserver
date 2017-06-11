// +build linux

package mbserver

import (
	"log"
	"os/exec"
	"testing"
	"time"

	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
)

// The serial read and close has a known race condition.
// https://github.com/golang/go/issues/10001
func TestModbusRTU(t *testing.T) {
	// Create a pair of virutal serial devices.
	cmd := exec.Command("socat",
		"pty,raw,echo=0,link=ttyFOO",
		"pty,raw,echo=0,link=ttyBAR")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer cmd.Wait()
	defer cmd.Process.Kill()

	// Allow the virutal serial devices to be created.
	time.Sleep(10 * time.Millisecond)

	// Server
	s := NewServer()
	err = s.ListenRTU(&serial.Config{
		Address:  "ttyFOO",
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  10 * time.Second})
	if err != nil {
		t.Fatalf("failed to listen, got %v\n", err)
	}
	defer s.Close()

	// Allow the server to start and to avoid a connection refused on the client
	time.Sleep(1 * time.Millisecond)

	// Client
	handler := modbus.NewRTUClientHandler("ttyBAR")
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second
	// Connect manually so that multiple requests are handled in one connection session
	err = handler.Connect()
	if err != nil {
		t.Errorf("failed to connect, got %v\n", err)
		t.FailNow()
	}
	defer handler.Close()
	client := modbus.NewClient(handler)

	// Coils
	_, err = client.WriteMultipleCoils(100, 9, []byte{255, 1})
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}

	results, err := client.ReadCoils(100, 16)
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}
	expect := []byte{255, 1}
	got := results
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}
