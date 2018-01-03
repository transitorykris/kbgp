package message

import (
	"bytes"

	"github.com/transitorykris/kbgp/stream"
)

// 4.5.  NOTIFICATION Message Format
//    A NOTIFICATION message is sent when an error condition is detected.
//    The BGP connection is closed immediately after it is sent.
//    In addition to the fixed-size BGP header, the NOTIFICATION message
//    contains the following fields:
//       0                   1                   2                   3
//       0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//       | Error code    | Error subcode |   Data (variable)             |
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type notificationMessage struct {
	code    byte
	subcode byte
	data    []byte
}

//       Error Code:
//          This 1-octet unsigned integer indicates the type of
//          NOTIFICATION.  The following Error Codes have been defined:
//                           Error Code       Symbolic Name               Reference
const (
	_                       = iota
	messageHeaderError      // 1         Message Header Error             Section 6.1
	openMessageError        // 2         OPEN Message Error               Section 6.2
	updateMessageError      // 3         UPDATE Message Error             Section 6.3
	holdTimerExpired        // 4         Hold Timer Expired               Section 6.5
	finiteStateMachineError // 5         Finite State Machine Error       Section 6.6
	cease                   // 6         Cease                            Section 6.7
)

// Reverse value lookup for error code names
var errorCodeName = map[int]string{
	1: "Message Header Error",
	2: "OPEN Message Error",
	3: "UPDATE Message Error",
	4: "Hold Timer Expired",
	5: "Finite State Machine Error",
	6: "Cease",
}

//       Error subcode:
//          This 1-octet unsigned integer provides more specific
//          information about the nature of the reported error.  Each Error
//          Code may have one or more Error Subcodes associated with it.
//          If no appropriate Error Subcode is defined, then a zero
//          (Unspecific) value is used for the Error Subcode field.
//       Message Header Error subcodes:
const (
	_                         = iota
	connectionNotSynchronized // 1 - Connection Not Synchronized.
	badMessageLength          // 2 - Bad Message Length.
	badMessageType            // 3 - Bad Message Type.
)

// Reverse value lookup for message header error subcodes
var messageHeaderErrorSubcodeName = map[int]string{
	1: "Connection Not Synchronized",
	2: "Bad Message Length",
	3: "Bad Message Type",
}

const noErrorSubcode = 0

//       OPEN Message Error subcodes:
const (
	_                            = iota
	unsupportedVersionNumber     // 1 - Unsupported Version Number.
	badPeerAS                    // 2 - Bad Peer AS.
	badBGPIdentifier             // 3 - Bad BGP Identifier.
	unsupportedOptionalParameter // 4 - Unsupported Optional Parameter.
	_                            // 5 - [Deprecated - see Appendix A].
	unacceptableHoldTime         // 6 - Unacceptable Hold Time.
)

// Reverse value lookup for open message error subcodes
var openMessageErrorSubcodeName = map[int]string{
	1: "Unsupported Version Number",
	2: "Bad Peer AS",
	3: "Bad BGP Identifier",
	4: "Unsupported Optional Parameter",
	// 5 is deprecated
	6: "Unacceptable Hold Time",
}

//       UPDATE Message Error subcodes:
const (
	_                              = iota
	malformedAttributeList         // 1 - Malformed Attribute List.
	unrecognizedWellKnownAttribute // 2 - Unrecognized Well-known Attribute.
	missingWellKnownAttribute      // 3 - Missing Well-known Attribute.
	attributeFlagsError            // 4 - Attribute Flags Error.
	attributeLengthError           // 5 - Attribute Length Error.
	invalidOriginAttribute         // 6 - Invalid ORIGIN Attribute.
	_                              // 7 - [Deprecated - see Appendix A].
	invalidNextHopAttribute        // 8 - Invalid NEXT_HOP Attribute.
	optionalAttributeError         // 9 - Optional Attribute Error.
	invalidNetworkField            // 10 - Invalid Network Field.
	malformedASPath                // 11 - Malformed AS_PATH.
)

// Reverse value lookup for update message error subcodes
var updateMessageErrorSubcodeName = map[int]string{
	1: "Malformed Attribute List",
	2: "Unrecognized Well-known Attribute",
	3: "Missing Well-known Attribute",
	4: "Attribute Flags Error",
	5: "Attribute Length Error",
	6: "Invalid ORIGIN Attribute",
	// 7 is deprecated
	8:  "Invalid NEXT_HOP Attribute",
	9:  "Optional Attribute Error",
	10: "Invalid Network Field",
	11: "Malformed AS_PATH",
}

//       Data:

//          This variable-length field is used to diagnose the reason for
//          the NOTIFICATION.  The contents of the Data field depend upon
//          the Error Code and Error Subcode.  See Section 6 for more
//          details.

//          Note that the length of the Data field can be determined from
//          the message Length field by the formula:

//                   Message Length = 21 + Data Length

//    The minimum length of the NOTIFICATION message is 21 octets
//    (including message header).
const minNotificationMessageLength = 21

func newNotificationMessage(code int, subcode int, data []byte) *notificationMessage {
	n := notificationMessage{
		code:    byte(code),
		subcode: byte(subcode),
		data:    data,
	}
	return &n
}

func readNotification(message []byte) *notificationMessage {
	buf := bytes.NewBuffer(message)
	code := stream.ReadByte(buf)
	subcode := stream.ReadByte(buf)
	var data []byte
	buf.Read(data)

	n := &notificationMessage{
		code:    code,
		subcode: subcode,
		data:    data,
	}
	//if n.code == openMessageError && n.code == unsupportedVersionNumber {
	//	f.sendEvent(notifMsgVerErr)
	//	return nil
	//}
	//f.sendEvent(notifMsg)

	// TODO: How and where do we make the notification data available?
	return n
}

func (n notificationMessage) valid() (*notificationMessage, bool) {
	// Note: here for completeness
	return nil, true
}

func (n notificationMessage) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

// 6.4.  NOTIFICATION Message Error Handling

//    If a peer sends a NOTIFICATION message, and the receiver of the
//    message detects an error in that message, the receiver cannot use a
//    NOTIFICATION message to report this error back to the peer.  Any such
//    error (e.g., an unrecognized Error Code or Error Subcode) SHOULD be
//    noticed, logged locally, and brought to the attention of the
//    administration of the peer.  The means to do this, however, lies
//    outside the scope of this document.
