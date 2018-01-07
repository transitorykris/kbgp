package message

import (
	"bytes"
	"net"

	"github.com/transitorykris/kbgp/stream"
)

// BGP messages are sent over TCP connections.  A message is processed
// only after it is entirely received.  The maximum message size is 4096
// octets.  All implementations are required to support this maximum
// message size.  The smallest message that may be sent consists of a
// BGP header without a data portion (19 octets).
const minMessageLength = 19
const maxMessageLength = 4096

// Each message has a fixed-size header.  There may or may not be a data
// portion following the header, depending on the message type.
type messageHeader struct {
	// This 16-octet field is included for compatibility; it MUST be
	// set to all ones.
	marker [markerLength]byte
	// This 2-octet unsigned integer indicates the total length of the
	// message, including the header in octets.  Thus, it allows one
	// to locate the (Marker field of the) next message in the TCP
	// stream.  The value of the Length field MUST always be at least
	// 19 and no greater than 4096, and MAY be further constrained,
	// depending on the message type.  "padding" of extra data after
	// the message is not allowed.  Therefore, the Length field MUST
	// have the smallest value required, given the rest of the
	// message.
	length uint16
	// This 1-octet unsigned integer indicates the type code of the
	// message.
	messageType byte
}

const markerLength = 16
const lengthLength = 2
const typeLength = 1
const messageHeaderLength = markerLength + lengthLength + typeLength

func marker() [markerLength]byte {
	b := bytes.Repeat([]byte{0xFF}, markerLength)
	var m [markerLength]byte
	copy(m[:], b)
	return m
}

// This 1-octet unsigned integer indicates the type code of the
// message.  This document defines the following type codes:
const (
	_            = iota
	open         // 1 - OPEN
	update       // 2 - UPDATE
	notification // 3 - NOTIFICATION
	keepalive    // 4 - KEEPALIVE
	//routeRefresh // [RFC2918] defines one more type code.
)

// Reverse value lookup for message type names
var messageName = map[int]string{
	1: "OPEN",
	2: "UPDATE",
	3: "NOTIFICATION",
	4: "KEEPALIVE",
	//5: "ROUTE-REFRESH" // [RFC2918] defines one more type code.
}

func readMessage(conn net.Conn) (messageHeader, []byte) {
	rawHeader := stream.Read(conn, messageHeaderLength)
	buf := bytes.NewBuffer(rawHeader)

	var marker [markerLength]byte
	copy(marker[:], buf.Next(markerLength)) // Note: We expect this to be all 1's
	length := stream.ReadUint16(buf)
	messageType := stream.ReadByte(buf)

	header := messageHeader{
		marker:      marker,
		length:      length,
		messageType: messageType,
	}
	message := stream.Read(conn, int(header.length))
	return header, message
}

func (m messageHeader) valid() (*notificationMessage, bool) {
	if bytes.Compare(m.marker[:], bytes.Repeat([]byte{0xFF}, 16)) != 0 {
		return newNotificationMessage(messageHeaderError, noErrorSubcode, nil), false
	}
	// TODO: Check if we can process this message type?
	return nil, true
}

func (m messageHeader) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

// 6.1.  Message Header Error Handling

//    All errors detected while processing the Message Header MUST be
//    indicated by sending the NOTIFICATION message with the Error Code
//    Message Header Error.  The Error Subcode elaborates on the specific
//    nature of the error.

//    The expected value of the Marker field of the message header is all
//    ones.  If the Marker field of the message header is not as expected,
//    then a synchronization error has occurred and the Error Subcode MUST
//    be set to Connection Not Synchronized.

//    If at least one of the following is true:

//       - if the Length field of the message header is less than 19 or
//         greater than 4096, or

//       - if the Length field of an OPEN message is less than the minimum
//         length of the OPEN message, or

//       - if the Length field of an UPDATE message is less than the
//         minimum length of the UPDATE message, or

//       - if the Length field of a KEEPALIVE message is not equal to 19,
//         or

//       - if the Length field of a NOTIFICATION message is less than the
//         minimum length of the NOTIFICATION message,

//    then the Error Subcode MUST be set to Bad Message Length.  The Data
//    field MUST contain the erroneous Length field.

//    If the Type field of the message header is not recognized, then the
//    Error Subcode MUST be set to Bad Message Type.  The Data field MUST
//    contain the erroneous Type field.
