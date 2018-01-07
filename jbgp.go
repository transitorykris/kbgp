package jbgp

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
type bgpIdentifier uint32

func (b bgpIdentifier) String() string {
	// TODO: convert to net.IP.String()
	return fmt.Sprintf("%s", b.ip())
}

func (b bgpIdentifier) ip() net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(b))
	return ip
}

const version = 4
