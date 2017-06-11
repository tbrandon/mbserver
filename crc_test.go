package mbserver

import "testing"

func TestCRC(t *testing.T) {
	got := crcModbus([]byte{0x01, 0x04, 0x02, 0xFF, 0xFF})
	expect := 0x80B8
	if !isEqual(expect, got) {
		t.Errorf("expected %x, got %x", expect, got)
	}
}
