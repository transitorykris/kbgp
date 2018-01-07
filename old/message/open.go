package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/transitorykris/kbgp/bgp"
	"github.com/transitorykris/kbgp/stream"
)

// After a TCP connection is established, the first message sent by each
// side is an OPEN message.  If the OPEN message is acceptable, a
// KEEPALIVE message confirming the OPEN is sent back.
type OpenMessage struct {
	// This 1-octet unsigned integer indicates the protocol version
	// number of the message.  The current BGP version number is 4.
	version bgp.Version
	// This 2-octet unsigned integer indicates the Autonomous System
	// number of the sender.
	myAS bgp.ASN
	// This 2-octet unsigned integer indicates the number of seconds
	// the sender proposes for the value of the Hold Timer.  Upon
	// receipt of an OPEN message, a BGP speaker MUST calculate the
	// value of the Hold Timer by using the smaller of its configured
	// Hold Time and the Hold Time received in the OPEN message.  The
	// Hold Time MUST be either zero or at least three seconds.  An
	// implementation MAY reject connections on the basis of the Hold
	// Time.  The calculated value indicates the maximum number of
	// seconds that may elapse between the receipt of successive
	// KEEPALIVE and/or UPDATE messages from the sender.
	holdTime uint16
	// This 4-octet unsigned integer indicates the BGP Identifier of
	// the sender.  A given BGP speaker sets the value of its BGP
	// Identifier to an IP address that is assigned to that BGP
	// speaker.  The value of the BGP Identifier is determined upon
	// startup and is the same for every local interface and BGP peer.
	bgpIdentifier bgp.Identifier
	// This 1-octet unsigned integer indicates the total length of the
	// Optional Parameters field in octets.  If the value of this
	// field is zero, no Optional Parameters are present.
	optParmLen byte
	// This field contains a list of optional parameters, in which
	// each parameter is encoded as a <Parameter Type, Parameter
	// Length, Parameter Value> triplet.
	optParameters []byte
}

// This field contains a list of optional parameters, in which
// each parameter is encoded as a <Parameter Type, Parameter
// Length, Parameter Value> triplet.
type Parameter struct {
	// Parameter Type is a one octet field that unambiguously
	// identifies individual parameters.  Parameter Length is a one
	// octet field that contains the length of the Parameter Value
	// field in octets.  Parameter Value is a variable length field
	// that is interpreted according to the value of the Parameter
	// Type field.
	parmType   byte
	parmLength byte
	parmValue  []byte
}

//       Version:
//          This 1-octet unsigned integer indicates the protocol version
//          number of the message.  The current BGP version number is 4.
const version = 4

//       Hold Time:
//          This 2-octet unsigned integer indicates the number of seconds
//          the sender proposes for the value of the Hold Timer.  Upon
//          receipt of an OPEN message, a BGP speaker MUST calculate the
//          value of the Hold Timer by using the smaller of its configured
//          Hold Time and the Hold Time received in the OPEN message.  The
//          Hold Time MUST be either zero or at least three seconds.  An
//          implementation MAY reject connections on the basis of the Hold
//          Time.  The calculated value indicates the maximum number of
//          seconds that may elapse between the receipt of successive
//          KEEPALIVE and/or UPDATE messages from the sender.
var maxHoldTime = time.Duration(int(math.Pow(2, 16))) * time.Second

const largeHoldTimer = 4 * time.Minute // See 8.2.2

//       BGP Identifier:
//          This 4-octet unsigned integer indicates the BGP Identifier of
//          the sender.  A given BGP speaker sets the value of its BGP
//          Identifier to an IP address that is assigned to that BGP
//          speaker.  The value of the BGP Identifier is determined upon
//          startup and is the same for every local interface and BGP peer.

//       Optional Parameters Length:
//          This 1-octet unsigned integer indicates the total length of the
//          Optional Parameters field in octets.  If the value of this
//          field is zero, no Optional Parameters are present.
const minOptParametersLength = 0
const maxOptParametersLength = 255

//    The minimum length of the OPEN message is 29 octets (including the
//    message header).
const minOpenMessageLength = 29

func readOpen(message []byte) *openMessage {
	buf := bytes.NewBuffer(message)
	om := &openMessage{
		version:       stream.ReadByte(buf),
		myAS:          stream.ReadUint16(buf),
		holdTime:      stream.ReadUint16(buf),
		bgpIdentifier: stream.ReadUint32(buf),
		optParmLen:    stream.ReadByte(buf),
		// Note: We should be reading this into a parameters struct
	}
	om.optParameters = stream.ReadBytes(int(om.optParmLen), buf)
	return om
}

func (o openMessage) valid(remoteAS uint16, holdTime uint16) (*notificationMessage, bool) {
	if o.version != version {
		return newNotificationMessage(openMessageError, unsupportedVersionNumber, nil), false
	}
	if o.myAS != remoteAS {
		return newNotificationMessage(openMessageError, badPeerAS, nil), false
	}
	if o.holdTime > 0 && holdTime < 3 {
		return newNotificationMessage(openMessageError, unacceptableHoldTime, nil), false
	}
	// TODO: What is an unacceptable bgp identifier?
	// badBGPIdentifier             // 3 - Bad BGP Identifier.
	// unsupportedOptionalParameter // 4 - Unsupported Optional Parameter.
	return nil, true
}

func (o openMessage) bytes() []byte {
	buf := bytes.NewBuffer([]byte{})

	buf.WriteByte(o.version)

	myAS := make([]byte, 2)
	binary.BigEndian.PutUint16(myAS, o.myAS)
	buf.Write(myAS)

	holdTime := make([]byte, 2)
	binary.BigEndian.PutUint16(holdTime, o.holdTime)
	buf.Write(holdTime)

	id := make([]byte, 4)
	binary.BigEndian.PutUint32(id, o.bgpIdentifier)
	buf.Write(id)

	buf.WriteByte(byte(len(o.optParameters)))
	buf.Write(o.optParameters)

	return buf.Bytes()
}

func isValidHoldTime(hold time.Duration) bool {
	if hold > maxHoldTime {
		return false
	}
	if hold > 0 && hold < 3*time.Second {
		return false
	}
	return true
}

func durationToUint16(t time.Duration) uint16 {
	return uint16(t.Seconds())
}

func parametersLength(parms []parameter) (byte, error) {
	var length uint16
	for _, p := range parms {
		length += uint16(p.parmLength)
	}
	if length > maxOptParametersLength {
		return 0x00, fmt.Errorf("Parameters length exceeds %d", maxOptParametersLength)
	}
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, length)
	return bs[0], nil
}

const maxParameterLength = 255

func newParameter(t byte, v []byte) (parameter, error) {
	if len(v) > maxParameterLength {
		return parameter{}, fmt.Errorf("Parameter exceeds maximum length of %d", maxParameterLength)
	}
	length := byte(len(v))
	return parameter{t, length, v}, nil
}

func readOptionalParameters(params []byte) []*parameter {
	// TODO: Implement me
	//optionalAttributeError         // 9 - Optional Attribute Error.
	return nil
}

func (p parameter) valid() (*notificationMessage, bool) {
	return nil, true
}

func (p parameter) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

// 6.2.  OPEN Message Error Handling

//    All errors detected while processing the OPEN message MUST be
//    indicated by sending the NOTIFICATION message with the Error Code
//    OPEN Message Error.  The Error Subcode elaborates on the specific
//    nature of the error.

//    If the version number in the Version field of the received OPEN
//    message is not supported, then the Error Subcode MUST be set to
//    Unsupported Version Number.  The Data field is a 2-octet unsigned
//    integer, which indicates the largest, locally-supported version
//    number less than the version the remote BGP peer bid (as indicated in

//    the received OPEN message), or if the smallest, locally-supported
//    version number is greater than the version the remote BGP peer bid,
//    then the smallest, locally-supported version number.

//    If the Autonomous System field of the OPEN message is unacceptable,
//    then the Error Subcode MUST be set to Bad Peer AS.  The determination
//    of acceptable Autonomous System numbers is outside the scope of this
//    protocol.

//    If the Hold Time field of the OPEN message is unacceptable, then the
//    Error Subcode MUST be set to Unacceptable Hold Time.  An
//    implementation MUST reject Hold Time values of one or two seconds.
//    An implementation MAY reject any proposed Hold Time.  An
//    implementation that accepts a Hold Time MUST use the negotiated value
//    for the Hold Time.

//    If the BGP Identifier field of the OPEN message is syntactically
//    incorrect, then the Error Subcode MUST be set to Bad BGP Identifier.
//    Syntactic correctness means that the BGP Identifier field represents
//    a valid unicast IP host address.

//    If one of the Optional Parameters in the OPEN message is not
//    recognized, then the Error Subcode MUST be set to Unsupported
//    Optional Parameters.

//    If one of the Optional Parameters in the OPEN message is recognized,
//    but is malformed, then the Error Subcode MUST be set to 0
//    (Unspecific).
