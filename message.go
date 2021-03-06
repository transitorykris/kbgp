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
	marker    marker
	msgLength uint16 // Includes the length of the header
	msgType   msgType
}

func newHeader(length int, msgType msgType) msgHeader {
	m := msgHeader{
		marker:    newMarker(),
		msgLength: uint16(length),
		msgType:   msgType,
	}
	return m
}

type marker [16]byte

func newMarker() marker {
	var m marker
	copy(m[:], bytes.Repeat([]byte{0xFF}, len(m)))
	return m
}

// bytes implements byter
func (m marker) bytes() []byte {
	return m[:]
}

// length implements byter
func (m marker) length() int { return len(m.bytes()) }

// bytes implements byter
func (h msgHeader) bytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Write(h.marker.bytes())
	buf.Write(uint16ToBytes(h.msgLength + 19)) // Ugly ugly, we have to include the length of the header
	buf.Write(h.msgType.bytes())
	return buf.Bytes()
}

// length implements byter
func (h msgHeader) length() int { return len(h.bytes()) }

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
	log.Println("Got raw header")
	buf := bytes.NewBuffer(rawHeader)

	// Read in the message header
	header := msgHeader{}
	copy(header.marker[:], buf.Next(markerLength))
	// TODO: Check that the marker is all 1s
	header.msgLength = stream.ReadUint16(buf)
	header.msgType = msgType(stream.ReadByte(buf))
	log.Println("Got header", header)

	// Read in the message's body
	body := stream.Read(r, int(header.msgLength)-messageHeaderLength)

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
		holdTime:      uint16(defaultHoldTime.Seconds()),     //TODO: make configurable
		bgpIdentifier: newIdentifier(net.ParseIP("1.2.3.4")), //TODO: make configurable
		optParmLen:    0,
		optParamaters: []parameter{},
	}
	log.Println("Open message:", o)
	return o
}

// bytes implements byter
func (o openMsg) bytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(o.version)
	buf.Write(o.as.bytes())
	buf.Write(uint16ToBytes(o.holdTime))
	buf.Write(o.bgpIdentifier.bytes())
	buf.WriteByte(0) // TODO: implement parameters
	// write optParameters []parameter
	return buf.Bytes()
}

// length implements byter
func (o openMsg) length() int {
	log.Println("LENGTH: got open")
	log.Println("LENGTH: got bytes", o.bytes())
	return len(o.bytes())
}

type byter interface {
	bytes() []byte
	length() int
}

func writeMessage(w io.Writer, msgType msgType, msg byter) (int, error) {
	var m []byte
	log.Println("Message length", msg.length())
	m = append(newHeader(msg.length(), msgType).bytes(), msg.bytes()...)
	n, err := w.Write(m)
	log.Println("Wrote message", m)
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

var errorCodeLookup = map[uint8]string{
	messageHeaderError:    "Message Header Error",
	openMessageError:      "OPEN Message Error",
	updateMessageError:    "UPDATE Message Error",
	holdTimerExpiredError: "Hold Timer Expired",
	fsmError:              "Finite State Machine Error",
	cease:                 "Cease",
}

const (
	_ = iota
	connectionNotSynchronized
	badMessageLength
	badMessageType
)

var messageHeaderErrorLookup = map[uint8]string{
	connectionNotSynchronized: "Connection Not Synchronized",
	badMessageLength:          "Bad Message Length",
	badMessageType:            "Bad Message Type",
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

var openMessageErrorLookup = map[uint8]string{
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

var updateMessageErrorLookup = map[uint8]string{
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

type notificationMsg struct {
	code    uint8
	subcode uint8
	data    []byte
}

func newNotification(err error) notificationMsg {
	return notificationMsg{uint8(err.(bgpError).code), uint8(err.(bgpError).subcode), []byte(err.(bgpError).message)}
}

// bytes implements byter
func (n notificationMsg) bytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(n.code)
	buf.WriteByte(n.subcode)
	buf.Write(n.data)
	return buf.Bytes()
}

// length implements byter
func (n notificationMsg) length() int { return len(n.bytes()) }

type keepaliveMsg struct{}

func newKeepalive() keepaliveMsg {
	return keepaliveMsg{}
}

func readKeepalive(msg []byte) error {
	if len(msg) != 0 {
		newBGPError(messageHeaderError, badMessageLength, "a keepalive should not come with data")
	}
	return nil
}

// bytes implements byter
func (k keepaliveMsg) bytes() []byte {
	return []byte{}
}

// length implements byter
func (k keepaliveMsg) length() int { return len(k.bytes()) }

func readNotification(msg []byte) (notificationMsg, error) {
	log.Println("Reading NOTIFICATION message")
	buf := bytes.NewBuffer(msg)
	nm := notificationMsg{
		code:    stream.ReadByte(buf),
		subcode: stream.ReadByte(buf),
		data:    stream.ReadBytes(len(msg), buf),
	}
	log.Println("Got NOTIFICATION message:", nm)
	return nm, nil
}

// String implements strings.Stringer
func (n notificationMsg) String() string {
	var subcode string
	switch n.code {
	case messageHeaderError:
		subcode = messageHeaderErrorLookup[n.subcode]
	case openMessageError:
		subcode = openMessageErrorLookup[n.subcode]
	case updateMessageError:
		subcode = updateMessageErrorLookup[n.subcode]
	default:
		subcode = "unknown"
	}
	return fmt.Sprintf("%s (%d) %s (%d) %s",
		errorCodeLookup[n.code], n.code, subcode, n.subcode, string(n.data))
}

type updateMsg struct{}

func newUpdate() updateMsg {
	//TODO: Implement me
	return updateMsg{}
}

func readUpdate(msg []byte) (updateMsg, error) {
	return updateMsg{}, nil
}

// bytes implements byter
func (u updateMsg) bytes() []byte {
	//TODO: Implement me
	return nil
}

// String implements strings.Stringer
func (u updateMsg) String() string {
	//TODO: Implement me
	return ""
}
