package jbgp

import (
	"fmt"
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

const version = 4
