package kbgp

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func TestMarker(t *testing.T) {
	m := marker()
	if len(m) != markerLength {
		t.Errorf("Expected marker length %d but got %d", markerLength, len(m))
	}
	for i, v := range m {
		if v != 0xFF {
			t.Errorf("Expected all bits to be 1, got %d at position %d", v, i)
		}
	}
}

func TestIsValidHoldTime(t *testing.T) {
	//if isValidHoldTime((maxHoldTime + 1) * time.Second) {
	//	t.Errorf("Expected maxHoldTime+1 to be an invalid hold time")
	//}
	if isValidHoldTime(1 * time.Second) {
		t.Errorf("Expected 1 seconds to be an invalid hold time")
	}
	if isValidHoldTime(2 * time.Second) {
		t.Errorf("Expected 2 seconds to be an invalid hold time")
	}
	if !isValidHoldTime(0 * time.Second) {
		t.Errorf("Expected 0 seconds to be a valid hold time")
	}
	if !isValidHoldTime(3 * time.Second) {
		t.Errorf("Expected 3 seconds to be a valid hold time")
	}
}

func TestFindBGPIdentifier(t *testing.T) {
	_, err := findBGPIdentifier()
	if err != nil {
		t.Errorf("Unexpected error guessing IP: %v", err)
	}
}

func TestIPToBGPIdentifier(t *testing.T) {
	ip := net.ParseIP("1.2.3.4")
	id := ipToUint32(ip)
	ip4 := binary.BigEndian.Uint32(ip.To4())
	if ip4 != id {
		t.Errorf("Incorrect identifier %d != %d", ip4, id)
	}
}
