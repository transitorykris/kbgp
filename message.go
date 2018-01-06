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

// https://tools.ietf.org/html/rfc4271#section-4.1
type msgHeader struct {
	marker  [16]byte
	length  uint16 // Includes the length of the header
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
	body := stream.Read(r, int(header.length)-messageHeaderLength)

	log.Println("read body")
	return header, body, nil
}

// https://tools.ietf.org/html/rfc4271#section-4.2
type openMsg struct {
	version       uint8
	as            asn
	holdTime      uint16
	bgpIdentifier bgpIdentifier
	optParmLen    uint8
	optParamaters []parameter
}

// String implements strings.Stringer
func (o openMsg) String() string {
	return fmt.Sprintf("Version: %d AS: %d HoldTime: %d bgpIdentifier: %s",
		o.version, o.as, o.holdTime, o.bgpIdentifier)
}

const minOpenMessageLength = 29

type parameter struct{}

func readOpen(msg []byte) (openMsg, error) {
	log.Println("Reading OPEN message")
	buf := bytes.NewBuffer(msg)
	om := openMsg{
		version:       stream.ReadByte(buf),
		as:            asn(stream.ReadUint16(buf)),
		holdTime:      stream.ReadUint16(buf),
		bgpIdentifier: bgpIdentifier(stream.ReadUint32(buf)),
		optParmLen:    stream.ReadByte(buf),
	}
	log.Println("Got OPEN message:", om)
	// TODO: Implement optional parameter parsing
	_ = stream.ReadBytes(int(om.optParmLen), buf)
	return om, nil
}

func writeMessage(w io.Writer, msgType msgType, msg []byte) (int, error) {
	msg = append(newHeader(len(msg), msgType).bytes(), msg...)
	n, err := w.Write(msg)
	return n, err
}

func newNotification(err error) []byte {
	return nil
}
