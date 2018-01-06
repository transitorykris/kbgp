package jbgp

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/transitorykris/jbgp/stream"
)

type msgType uint8

const (
	_ = iota
	open
	update
	notification
	keepalive
)

var msgTypeLookup = map[msgType]string{
	open:         "OPEN",
	update:       "UPDATE",
	notification: "NOTIFICATION",
	keepalive:    "KEEPALIVE",
}

func (m msgType) String() string {
	t, ok := msgTypeLookup[m]
	if !ok {
		return "UNKNOWN"
	}
	return t
}

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

// String implements strings.Stringer
func (h msgHeader) String() string {
	return fmt.Sprintf("Message length: %d type: %s", h.length, h.msgType)
}

const markerLength = 16
const lengthLength = 2
const typeLength = 1
const messageHeaderLength = markerLength + lengthLength + typeLength

func readHeader(r io.Reader) (msgHeader, []byte, error) {
	log.Println("Reading message header")
	rawHeader := stream.Read(r, messageHeaderLength)
	buf := bytes.NewBuffer(rawHeader)

	// Read in the message header
	header := msgHeader{}
	copy(header.marker[:], buf.Next(markerLength))
	// TODO: Check that the marker is all 1s
	header.length = stream.ReadUint16(buf)
	header.msgType = msgType(stream.ReadByte(buf))

	log.Println("Got header", header)

	// Read in the message's body
	body := stream.Read(r, int(header.length))

	return header, body, nil
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
