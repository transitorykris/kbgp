package kbgp

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestFindBGPIdentifier(t *testing.T) {
	_, err := FindBGPIdentifier()
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

func TestUint32ToIP(t *testing.T) {
	i := uint32(167772773)
	expectedIP := net.ParseIP("10.0.2.101")
	ip := Uint32ToIP(i)
	if !ip.Equal(expectedIP) {
		t.Errorf("Expected IP %v but got %v", expectedIP, ip)
	}
}
