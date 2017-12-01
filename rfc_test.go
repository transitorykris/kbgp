package kbgp

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
	"unsafe"
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

func TestPackPrefix(t *testing.T) {
	b := packPrefix(32, net.ParseIP("1.2.3.4"))
	if len(b) != 4 {
		t.Errorf("Expected 1.2.3.4/32 to be 4 bytes but got %d", b)
	}
	b = packPrefix(25, net.ParseIP("1.2.3.4"))
	if len(b) != 4 {
		t.Errorf("Expected 1.2.3.4/25 to be 4 bytes but got %d", b)
	}
	b = packPrefix(24, net.ParseIP("1.2.3.4"))
	if len(b) != 3 {
		t.Errorf("Expected 1.2.3.4/24 to be 3 bytes but got %d", b)
	}
	b = packPrefix(16, net.ParseIP("1.2.3.4"))
	if len(b) != 2 {
		t.Errorf("Expected 1.2.3.4/16 to be 2 bytes but got %d", b)
	}
	b = packPrefix(8, net.ParseIP("1.2.3.4"))
	if len(b) != 1 {
		t.Errorf("Expected 1.2.3.4/8 to be 2 bytes but got %d", b)
	}
	b = packPrefix(1, net.ParseIP("1.2.3.4"))
	if len(b) != 1 {
		t.Errorf("Expected 1.2.3.4/1 to be 1 bytes but got %d", b)
	}
	b = packPrefix(0, net.ParseIP("1.2.3.4"))
	if len(b) != 1 {
		t.Errorf("Expected 1.2.3.4/0 to be 1 bytes but got %d", b)
	}
}

func TestNewNLRI(t *testing.T) {
	n := newNLRI(23, net.ParseIP("1.2.3.4"))
	if n.length != 23 {
		t.Errorf("Expected length to be 23 but got %d", n.length)
	}
	if bytes.Compare(n.prefix, []byte{1, 2, 3}) != 0 {
		t.Errorf("Expected bytes to be [1,2,3] but got %v", n.prefix)
	}
}

func TestNewKeepalive(t *testing.T) {
	k := newKeepaliveMessage()
	if unsafe.Sizeof(k) != 0 {
		t.Errorf("Keepalive messages must have 0 length")
	}
}

func TestNewNotificationMessage(t *testing.T) {
	data := []byte{0x03, 0x14, 0x25}
	n := newNotificationMessage(messageHeaderError, badMessageType, data)
	if n.code != messageHeaderError {
		t.Errorf("Expected error code %d but got %d", messageHeaderError, n.code)
	}
	if n.subcode != badMessageType {
		t.Errorf("Expected error subcode %d but got %d", badMessageType, n.subcode)
	}
	if bytes.Compare(n.data, data) != 0 {
		t.Errorf("Expected data %v but got %v", data, n.data)
	}
}

func TestNewFSM(t *testing.T) {
	f := newFSM()
	if f.state != idleState {
		t.Errorf("New FSMs start in the idle state, instead it's state %d", f.state)
	}
}
