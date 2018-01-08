package kbgp

import (
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
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

// bytes implements byter
func (a asn) bytes() []byte {
	return uint16ToBytes(uint16(a))
}

// length implements byter
func (a asn) length() int { return int(reflect.TypeOf(a).Size()) }

type bgpIdentifier uint32

func newIdentifier(ip net.IP) bgpIdentifier {
	return bgpIdentifier(binary.BigEndian.Uint32(ip.To4()))
}

// String implements strings.Stringer
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

// bytes implements byter
func (b bgpIdentifier) bytes() []byte {
	return uint32ToBytes(uint32(b))
}

// length implements byter
func (b bgpIdentifier) length() int { return int(reflect.TypeOf(b).Size()) }

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
