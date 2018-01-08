package kbgp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/transitorykris/kbgp/stream"
)

type msgType uint8

func (m msgType) bytes() []byte {
	return []byte{byte(m)}
}

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
	marker  marker
	length  uint16 // Includes the length of the header
	msgType msgType
}

func newHeader(length int, msgType msgType) msgHeader {
	m := msgHeader{
		marker:  newMarker(),
		length:  uint16(length),
		msgType: msgType,
	}
	return m
}

type marker [16]byte

func newMarker() marker {
	var m marker
	copy(m[:], bytes.Repeat([]byte{0x1}, len(m)))
	return m
}

func (m marker) bytes() []byte {
	return bytes.Repeat([]byte{0x1}, len(m))
}

func (h msgHeader) bytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Write(h.marker.bytes())
	buf.Write(uint16ToBytes(h.length))
	buf.Write(h.msgType.bytes())
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
	return fmt.Sprintf("Version:%d AS:%d HoldTime:%d bgpIdentifier:%s",
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

func newOpen(p *Peer) openMsg {
	o := openMsg{
		version:       version,
		as:            p.myAS,
		holdTime:      uint16(defaultHoldTime.Seconds()),       //TODO: make configurable
		bgpIdentifier: newIdentifier(net.ParseIP("127.0.0.1")), //TODO: make configurable
		optParmLen:    0,
		optParamaters: []parameter{},
	}
	return o
}

func (o openMsg) bytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(o.version)
	buf.Write(o.as.bytes())
	buf.Write(o.bgpIdentifier.bytes())
	buf.Write(uint16ToBytes(o.holdTime))
	buf.Write(o.bgpIdentifier.bytes())
	buf.WriteByte(0) // TODO: implement parameters
	// write optParameters []parameter
	return buf.Bytes()
}

type byter interface {
	bytes() []byte
}

func writeMessage(w io.Writer, msgType msgType, msg byter) (int, error) {
	msg = append(newHeader(len(msg), msgType).bytes(), msg.bytes())
	n, err := w.Write(msg)
	return n, err
}

const (
	_ = iota
	messageHeaderError
	openMessageError
	updateMessageError
	holdTimerExpiredError
	fsmError
	cease
)

var errorCodeLookup = map[int]string{
	messageHeaderError:    "Message Header Error",
	openMessageError:      "OPEN Message Error",
	updateMessageError:    "UPDATE Message Error",
	holdTimerExpiredError: "Hold Timer Expired",
	fsmError:              "Finite State Machine Error",
	cease:                 "Cease",
}

const (
	_ = iota
	unsupportedVersionNumber
	badPeerAS
	badBGPIdentifier
	unsupportedOptionalParameter
	_ // 5 is deprecated
	unacceptableHoldTime
)

var openMessageErrorLookup = map[int]string{
	unsupportedVersionNumber:     "Unsupported Version Number",
	badPeerAS:                    "Bad Peer AS",
	badBGPIdentifier:             "Bad BGP Identifier",
	unsupportedOptionalParameter: "Unsupported Optional Parameter",
	unacceptableHoldTime:         "Unacceptable Hold Time",
}

const (
	_ = iota
	malformedAttributeList
	unrecognizedWellKnownAttribute
	missingWellKnownAttribute
	attributeFlagsError
	attributeLengthError
	invalidOriginAttribute
	_ // 7 is deprecated
	invalidNextHopAttribute
	optionalAttributeError
	invalidNetworkField
	malformedASPath
)

var updateMessageErrorLookup = map[int]string{
	malformedAttributeList:         "Malformed Attribute List",
	unrecognizedWellKnownAttribute: "Unrecognized Well-known Attribute",
	missingWellKnownAttribute:      "Missing Well-known Attribute",
	attributeFlagsError:            "Attribute Flags Error",
	attributeLengthError:           "Attribute Length Error",
	invalidOriginAttribute:         "Invalid ORIGIN Attribute",
	invalidNextHopAttribute:        "Invalid NEXT_HOP Attribute",
	optionalAttributeError:         "Optional Attribute Erro",
	invalidNetworkField:            "Invalid Network Field",
	malformedASPath:                "Malformed AS_PATH",
}

func newNotification(err error) []byte {
	return nil
}
