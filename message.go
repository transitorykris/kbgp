package jbgp

import "io"

type msgType uint8

const (
	_ = iota
	open
	update
	notification
	keepalive
)

type msgHeader struct {
	marker  [16]byte
	length  uint16
	msgType msgType
}

func newHeader(length int, msgType msgType) msgHeader {
	return msgHeader{}
}

func (h msgHeader) bytes() []byte {
	return nil
}

func readHeader(w io.Reader) (msgHeader, error) {
	return msgHeader{}, nil
}

type openMsg struct {
	version       uint8
	as            asn
	holdTime      uint16
	bgpIdentifier bgpIdentifier
	optParmLen    uint8
	optParamaters []parameter
}

type parameter struct{}

func readOpen(w io.Reader) (openMsg, error) {
	// if openMsg.version != 4 { fmt.Errorf("Invalid BGP version") }
	return openMsg{version: version, as: 1234}, nil
}

func writeMessage(w io.Writer, msgType msgType, msg []byte) (int, error) {
	msg = append(newHeader(len(msg), msgType).bytes(), msg...)
	n, err := w.Write(msg)
	return n, err
}

func newNotification(err error) []byte {
	return nil
}
