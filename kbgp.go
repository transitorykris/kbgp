package kbgp

import (
	"encoding/binary"
	"fmt"
	"net"
)

type bgpError struct {
	code    int
	subcode int
	message string
}

func newBGPError(code int, subcode int, message string) error {
	return bgpError{code, subcode, message}
}

func (e bgpError) Error() string {
	return fmt.Sprintf("error code: %d subcode: %d message: %s", e.code, e.subcode, e.message)
}

type asn uint16

func (a asn) bytes() []byte {
	return uint16ToBytes(uint16(a))
}

type bgpIdentifier uint32

func newIdentifier(ip net.IP) bgpIdentifier {
	return bgpIdentifier(binary.BigEndian.Uint32(ip.To4()))
}

func (b bgpIdentifier) String() string {
	return fmt.Sprintf("%s", b.ip())
}

func (b bgpIdentifier) ip() net.IP {
	return uint32ToBytes(uint32(b))
}

// A bgpIdentifier is valid if it represents a valid unicast host IP
func (b bgpIdentifier) valid() bool {
	return b.ip().IsGlobalUnicast()
}

func (b bgpIdentifier) bytes() []byte {
	return uint32ToBytes(uint32(b))
}

const version = 4

func uint16ToBytes(v uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return b
}

func uint32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}
