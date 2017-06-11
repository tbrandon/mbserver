package mbserver

import (
	"testing"
	"time"

	"github.com/goburrow/modbus"
)

func TestAduRegisterAndNumber(t *testing.T) {
	var frame TCPFrame
	SetDataWithRegisterAndNumber(&frame, 0, 64)

	expect := []byte{0, 0, 0, 64}
	got := frame.Data
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

func TestAduSetDataWithRegisterAndNumberAndValues(t *testing.T) {
	var frame TCPFrame
	SetDataWithRegisterAndNumberAndValues(&frame, 7, 2, []uint16{3, 4})

	expect := []byte{0, 7, 0, 2, 4, 0, 3, 0, 4}
	got := frame.Data
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}

func TestUnsupportedFunction(t *testing.T) {
	s := NewServer()
	var frame TCPFrame
	frame.Function = 255

	var req Request
	req.frame = &frame
	response := s.handle(&req)
	exception := GetException(response)
	if exception != IllegalFunction {
		t.Errorf("expected IllegalFunction (%d), got (%v)", IllegalFunction, exception)
	}
}

func TestModbus(t *testing.T) {
	// Server
	s := NewServer()
	err := s.ListenTCP("127.0.0.1:3333")
	if err != nil {
		t.Fatalf("failed to listen, got %v\n", err)
	}
	defer s.Close()

	// Allow the server to start and to avoid a connection refused on the client
	time.Sleep(1 * time.Millisecond)

	// Client
	handler := modbus.NewTCPClientHandler("127.0.0.1:3333")
	// Connect manually so that multiple requests are handled in one connection session
	err = handler.Connect()
	if err != nil {
		t.Errorf("failed to connect, got %v\n", err)
		t.FailNow()
	}
	defer handler.Close()
	client := modbus.NewClient(handler)

	// Coils
	results, err := client.WriteMultipleCoils(100, 9, []byte{255, 1})
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}

	results, err = client.ReadCoils(100, 16)
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}
	expect := []byte{255, 1}
	got := results
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}

	// Discrete inputs
	results, err = client.ReadDiscreteInputs(0, 64)
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}
	// test: 2017/05/14 21:09:53 modbus: sending 00 01 00 00 00 06 ff 02 00 00 00 40
	// test: 2017/05/14 21:09:53 modbus: received 00 01 00 00 00 0b ff 02 08 00 00 00 00 00 00 00 00
	expect = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}

	// Holding registers
	results, err = client.WriteMultipleRegisters(1, 2, []byte{0, 3, 0, 4})
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}
	// received: 00 01 00 00 00 06 ff 10 00 01 00 02
	expect = []byte{0, 2}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}

	results, err = client.ReadHoldingRegisters(1, 2)
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}
	expect = []byte{0, 3, 0, 4}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}

	// Input registers
	s.InputRegisters[65530] = 1
	s.InputRegisters[65535] = 65535
	results, err = client.ReadInputRegisters(65530, 6)
	if err != nil {
		t.Errorf("expected nil, got %v\n", err)
		t.FailNow()
	}
	expect = []byte{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255}
	got = results
	if !isEqual(expect, got) {
		t.Errorf("expected %v, got %v", expect, got)
	}
}
