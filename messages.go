package kbgp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/transitorykris/kbgp/stream"
)

// 4.  Message Formats

//    This section describes message formats used by BGP.

//    BGP messages are sent over TCP connections.  A message is processed
//    only after it is entirely received.  The maximum message size is 4096
//    octets.  All implementations are required to support this maximum
//    message size.  The smallest message that may be sent consists of a
//    BGP header without a data portion (19 octets).
const minMessageLength = 19
const maxMessageLength = 4096

//    All multi-octet fields are in network byte order.

// 4.1.  Message Header Format
//    Each message has a fixed-size header.  There may or may not be a data
//    portion following the header, depending on the message type.  The
//    layout of these fields is shown below:
//       0                   1                   2                   3
//       0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//       |                                                               |
//       +                                                               +
//       |                                                               |
//       +                                                               +
//       |                           Marker                              |
//       +                                                               +
//       |                                                               |
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//       |          Length               |      Type     |
//       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type messageHeader struct {
	//       Marker:
	//          This 16-octet field is included for compatibility; it MUST be
	//          set to all ones.
	marker [markerLength]byte

	//       Length:
	//          This 2-octet unsigned integer indicates the total length of the
	//          message, including the header in octets.  Thus, it allows one
	//          to locate the (Marker field of the) next message in the TCP
	//          stream.  The value of the Length field MUST always be at least
	//          19 and no greater than 4096, and MAY be further constrained,
	//          depending on the message type.  "padding" of extra data after
	//          the message is not allowed.  Therefore, the Length field MUST
	//          have the smallest value required, given the rest of the
	//          message.
	length uint16

	//       Type:
	//          This 1-octet unsigned integer indicates the type code of the
	//          message.  This document defines the following type codes:

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

//       Type:
//          This 1-octet unsigned integer indicates the type code of the
//          message.  This document defines the following type codes:
const (
	_            = iota
	open         // 1 - OPEN
	update       // 2 - UPDATE
	notification // 3 - NOTIFICATION
	keepalive    // 4 - KEEPALIVE
	//routeRefresh // [RFC2918] defines one more type code.
)

var messageName = map[int]string{
	1: "OPEN",
	2: "UPDATE",
	3: "NOTIFICATION",
	4: "KEEPALIVE",
	//5: "ROUTE-REFRESH" // [RFC2918] defines one more type code.
}

// 4.2.  OPEN Message Format
//    After a TCP connection is established, the first message sent by each
//    side is an OPEN message.  If the OPEN message is acceptable, a
//    KEEPALIVE message confirming the OPEN is sent back.
//    In addition to the fixed-size BGP header, the OPEN message contains
//    the following fields:
//        0                   1                   2                   3
//        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//        +-+-+-+-+-+-+-+-+
//        |    Version    |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        |     My Autonomous System      |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        |           Hold Time           |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        |                         BGP Identifier                        |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        | Opt Parm Len  |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//        |                                                               |
//        |             Optional Parameters (variable)                    |
//        |                                                               |
//        +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type openMessage struct {
	//       Version:
	//          This 1-octet unsigned integer indicates the protocol version
	//          number of the message.  The current BGP version number is 4.
	version byte

	//       My Autonomous System:
	//          This 2-octet unsigned integer indicates the Autonomous System
	//          number of the sender.
	myAS uint16

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
	holdTime uint16

	//       BGP Identifier:
	//          This 4-octet unsigned integer indicates the BGP Identifier of
	//          the sender.  A given BGP speaker sets the value of its BGP
	//          Identifier to an IP address that is assigned to that BGP
	//          speaker.  The value of the BGP Identifier is determined upon
	//          startup and is the same for every local interface and BGP peer.
	bgpIdentifier uint32

	//       Optional Parameters Length:
	//          This 1-octet unsigned integer indicates the total length of the
	//          Optional Parameters field in octets.  If the value of this
	//          field is zero, no Optional Parameters are present.
	optParmLen byte

	//       Optional Parameters:
	//          This field contains a list of optional parameters, in which
	//          each parameter is encoded as a <Parameter Type, Parameter
	//          Length, Parameter Value> triplet.
	//          0                   1
	//          0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
	//          +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-...
	//          |  Parm. Type   | Parm. Length  |  Parameter Value (variable)
	//          +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-...
	optParameters []byte
}

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

//       Optional Parameters:
//          This field contains a list of optional parameters, in which
//          each parameter is encoded as a <Parameter Type, Parameter
//          Length, Parameter Value> triplet.
//          0                   1
//          0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
//          +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-...
//          |  Parm. Type   | Parm. Length  |  Parameter Value (variable)
//          +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-...
type parameter struct {
	//          Parameter Type is a one octet field that unambiguously
	//          identifies individual parameters.  Parameter Length is a one
	//          octet field that contains the length of the Parameter Value
	//          field in octets.  Parameter Value is a variable length field
	//          that is interpreted according to the value of the Parameter
	//          Type field.
	parmType   byte
	parmLength byte
	parmValue  []byte
}

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

//          [RFC3392] defines the Capabilities Optional Parameter.

//    The minimum length of the OPEN message is 29 octets (including the
//    message header).
const minOpenMessageLength = 29

// 4.3.  UPDATE Message Format
//    UPDATE messages are used to transfer routing information between BGP
//    peers.  The information in the UPDATE message can be used to
//    construct a graph that describes the relationships of the various
//    Autonomous Systems.  By applying rules to be discussed, routing
//    information loops and some other anomalies may be detected and
//    removed from inter-AS routing.
//    An UPDATE message is used to advertise feasible routes that share
//    common path attributes to a peer, or to withdraw multiple unfeasible
//    routes from service (see 3.1).  An UPDATE message MAY simultaneously
//    advertise a feasible route and withdraw multiple unfeasible routes
//    from service.  The UPDATE message always includes the fixed-size BGP
//    header, and also includes the other fields, as shown below (note,
//    some of the shown fields may not be present in every UPDATE message):
//       +-----------------------------------------------------+
//       |   Withdrawn Routes Length (2 octets)                |
//       +-----------------------------------------------------+
//       |   Withdrawn Routes (variable)                       |
//       +-----------------------------------------------------+
//       |   Total Path Attribute Length (2 octets)            |
//       +-----------------------------------------------------+
//       |   Path Attributes (variable)                        |
//       +-----------------------------------------------------+
//       |   Network Layer Reachability Information (variable) |
//       +-----------------------------------------------------+
type updateMessage struct {
	withdrawnRoutesLength uint16
	withdrawnRoutes       []withdrawnRoute
	pathAttributesLength  uint16
	pathAttributes        []pathAttribute
	nlris                 []nlri
}

func readUpdate(message []byte) *updateMessage {
	buf := bytes.NewBuffer(message)
	update := new(updateMessage)
	update.withdrawnRoutesLength = stream.ReadUint16(buf)
	update.withdrawnRoutes = readWithdrawnRoutes(
		int(update.withdrawnRoutesLength),
		stream.ReadBytes(int(update.withdrawnRoutesLength), buf),
	)
	update.pathAttributesLength = stream.ReadUint16(buf)
	update.pathAttributes = readPathAttributes(
		int(update.pathAttributesLength),
		stream.ReadBytes(int(update.pathAttributesLength), buf),
	)
	// TODO: Add reading path attributes
	// TODO: Add reading NLRIs
	return update
}

func (u updateMessage) valid() (*notificationMessage, bool) {
	// updateMessageError
	// malformedAttributeList         // 1 - Malformed Attribute List.
	// unrecognizedWellKnownAttribute // 2 - Unrecognized Well-known Attribute.
	// missingWellKnownAttribute      // 3 - Missing Well-known Attribute.
	// attributeFlagsError            // 4 - Attribute Flags Error.
	// attributeLengthError           // 5 - Attribute Length Error.
	// invalidOriginAttribute         // 6 - Invalid ORIGIN Attribute.
	// invalidNextHopAttribute        // 8 - Invalid NEXT_HOP Attribute.
	// optionalAttributeError         // 9 - Optional Attribute Error.
	// invalidNetworkField            // 10 - Invalid Network Field.
	// malformedASPath                // 11 - Malformed AS_PATH.
	return nil, true
}

func (u updateMessage) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

//       Withdrawn Routes Length:
//          This 2-octets unsigned integer indicates the total length of
//          the Withdrawn Routes field in octets.  Its value allows the
//          length of the Network Layer Reachability Information field to
//          be determined, as specified below.
//          A value of 0 indicates that no routes are being withdrawn from
//          service, and that the WITHDRAWN ROUTES field is not present in
//          this UPDATE message.
const noWithdrawnRoutes = 0

//       Withdrawn Routes:
//          This is a variable-length field that contains a list of IP
//          address prefixes for the routes that are being withdrawn from
//          service.  Each IP address prefix is encoded as a 2-tuple of the
//          form <length, prefix>, whose fields are described below:
//                   +---------------------------+
//                   |   Length (1 octet)        |
//                   +---------------------------+
//                   |   Prefix (variable)       |
//                   +---------------------------+
//          The use and the meaning of these fields are as follows:
type withdrawnRoute struct {
	//          a) Length:
	//             The Length field indicates the length in bits of the IP
	//             address prefix.  A length of zero indicates a prefix that
	//             matches all IP addresses (with prefix, itself, of zero
	//             octets).
	length byte

	//          b) Prefix:
	//             The Prefix field contains an IP address prefix, followed by
	//             the minimum number of trailing bits needed to make the end
	//             of the field fall on an octet boundary.  Note that the value
	//             of trailing bits is irrelevant.
	prefix []byte
}

func readWithdrawnRoutes(length int, bs []byte) []withdrawnRoute {
	count := 0
	routes := []withdrawnRoute{}
	for count != length {
		wr := readWithdrawnRoute(bs)
		routes = append(routes, *wr)
		// Remove what we just read into a withdrawn route
		bs = bs[wr.length:]
		count += int(wr.length)
	}
	return routes
}

func readWithdrawnRoute(bs []byte) *withdrawnRoute {
	buf := bytes.NewBuffer(bs)
	route := new(withdrawnRoute)
	route.length = stream.ReadByte(buf)
	route.prefix = stream.ReadBytes(int(route.length), buf)
	return route
}

func (w withdrawnRoute) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

//       Total Path Attribute Length:
//          This 2-octet unsigned integer indicates the total length of the
//          Path Attributes field in octets.  Its value allows the length
//          of the Network Layer Reachability field to be determined as
//          specified below.

//          A value of 0 indicates that neither the Network Layer
//          Reachability Information field nor the Path Attribute field is
//          present in this UPDATE message.

//       Path Attributes:
//          A variable-length sequence of path attributes is present in
//          every UPDATE message, except for an UPDATE message that carries
//          only the withdrawn routes.  Each path attribute is a triple
//          <attribute type, attribute length, attribute value> of variable
//          length.
type pathAttribute struct {
	attributeType   attributeType
	attributeLength uint16
	attributeValue  []byte
}

func readPathAttributes(length int, bs []byte) []pathAttribute {
	count := 0
	attributes := []pathAttribute{}
	for count != length {
		pa := readPathAttribute(bs)
		attributes = append(attributes, *pa)
		// Remove what we just read
		bs = bs[pa.attributeLength:]
		count += int(pa.attributeLength)
	}
	return attributes
}

func readPathAttribute(bs []byte) *pathAttribute {
	attribute := new(pathAttribute)
	attribute.attributeType = readAttributeType(bs)
	// Remove the 2 byte type from our bytes
	bs = bs[2:]
	buf := bytes.NewBuffer(bs)
	attribute.attributeLength = stream.ReadUint16(buf)
	attribute.attributeValue = stream.ReadBytes(int(attribute.attributeLength), buf)
	return attribute
}

func (p pathAttribute) valid() (*notificationMessage, bool) {
	// 	attributeLengthError           // 5 - Attribute Length Error.
	//invalidNextHopAttribute        // 8 - Invalid NEXT_HOP Attribute.
	return nil, true
}

func (p pathAttribute) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

//          Attribute Type is a two-octet field that consists of the
//          Attribute Flags octet, followed by the Attribute Type Code
//          octet.
//                0                   1
//                0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
//                +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//                |  Attr. Flags  |Attr. Type Code|
//                +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type attributeType struct {
	flags byte
	code  byte
}

func readAttributeType(bs []byte) attributeType {
	attribute := attributeType{
		flags: bs[0],
		code:  bs[1],
	}
	return attribute
}

func (a attributeType) valid() (*notificationMessage, bool) {
	// 	attributeFlagsError            // 4 - Attribute Flags Error.
	// 	invalidOriginAttribute         // 6 - Invalid ORIGIN Attribute.
	return nil, true
}

func (a attributeType) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

//          The high-order bit (bit 0) of the Attribute Flags octet is the
//          Optional bit.  It defines whether the attribute is optional (if
//          set to 1) or well-known (if set to 0).
const optional = 1 << 7
const wellKnown = 0
const optionalMask = 1 << 7

func (a *attributeType) optional() bool {
	return a.flags&optionalMask == optional
}

func (a *attributeType) setOptional() {
	a.flags = a.flags | optional
}

func (a *attributeType) wellKnown() bool {
	return a.flags&optionalMask == wellKnown
}

func (a *attributeType) setWellKnown() {
	a.flags = a.flags &^ optionalMask
}

//          The second high-order bit (bit 1) of the Attribute Flags octet
//          is the Transitive bit.  It defines whether an optional
//          attribute is transitive (if set to 1) or non-transitive (if set
//          to 0).
const transitive = 1 << 6
const nonTransitive = 0
const transitiveMask = 1 << 6

func (a *attributeType) transitive() bool {
	return a.flags&transitiveMask == transitive
}

func (a *attributeType) setTransitive() {
	a.flags = a.flags | transitive
}

func (a *attributeType) nonTransitive() bool {
	return a.flags&transitiveMask == nonTransitive
}

func (a *attributeType) setNonTransitive() {
	a.flags = a.flags &^ transitiveMask
}

//          For well-known attributes, the Transitive bit MUST be set to 1.
//          (See Section 5 for a discussion of transitive attributes.)

//          The third high-order bit (bit 2) of the Attribute Flags octet
//          is the Partial bit.  It defines whether the information
//          contained in the optional transitive attribute is partial (if
//          set to 1) or complete (if set to 0).  For well-known attributes
//          and for optional non-transitive attributes, the Partial bit
//          MUST be set to 0.
const partial = 1 << 5
const complete = 0
const partialMask = 1 << 5

func (a *attributeType) partial() bool {
	return a.flags&partialMask == partial
}

func (a *attributeType) setPartial() {
	a.flags = a.flags | partial
}

func (a *attributeType) complete() bool {
	return a.flags&partialMask == complete
}

func (a *attributeType) setComplete() {
	a.flags = a.flags &^ partial
}

//          The fourth high-order bit (bit 3) of the Attribute Flags octet
//          is the Extended Length bit.  It defines whether the Attribute
//          Length is one octet (if set to 0) or two octets (if set to 1).
const extendedLength = 1 << 4
const notExtendedLength = 0
const extendedLengthMask = 1 << 4

func (a *attributeType) extendedLength() bool {
	return a.flags&extendedLengthMask == extendedLength
}

func (a *attributeType) setExtendedLength() {
	a.flags = a.flags | extendedLength
}

// TODO: Fix this name, it should be notExtendedLength
func (a *attributeType) nonextendedLength() bool {
	return a.flags&extendedLengthMask == notExtendedLength
}

func (a *attributeType) setNotExtendedLength() {
	a.flags = a.flags &^ extendedLength
}

//          The lower-order four bits of the Attribute Flags octet are
//          unused.  They MUST be zero when sent and MUST be ignored when
//          received.

//          The Attribute Type Code octet contains the Attribute Type Code.
//          Currently defined Attribute Type Codes are discussed in Section
//          5.

//          If the Extended Length bit of the Attribute Flags octet is set
//          to 0, the third octet of the Path Attribute contains the length
//          of the attribute data in octets.

//          If the Extended Length bit of the Attribute Flags octet is set
//          to 1, the third and fourth octets of the path attribute contain
//          the length of the attribute data in octets.

//          The remaining octets of the Path Attribute represent the
//          attribute value and are interpreted according to the Attribute
//          Flags and the Attribute Type Code.  The supported Attribute
//          Type Codes, and their attribute values and uses are as follows:

//          a) ORIGIN (Type Code 1):
//             ORIGIN is a well-known mandatory attribute that defines the
//             origin of the path information.  The data octet can assume
//             the following values:
//                Value      Meaning
const (
	//                0         IGP - Network Layer Reachability Information
	//                             is interior to the originating AS
	igp = iota
	//                1         EGP - Network Layer Reachability Information
	//                             learned via the EGP protocol [RFC904]
	egp
	//                2         INCOMPLETE - Network Layer Reachability
	//                             Information learned by some other means
	incomplete
)

var originName = map[int]string{
	0: "IGP",
	1: "EGP",
	2: "INCOMPLETE",
}

//             Usage of this attribute is defined in 5.1.1.

//          b) AS_PATH (Type Code 2):

//             AS_PATH is a well-known mandatory attribute that is composed
//             of a sequence of AS path segments.  Each AS path segment is
//             represented by a triple <path segment type, path segment
//             length, path segment value>.

//             The path segment type is a 1-octet length field with the
//             following values defined:
//                Value      Segment Type

const (
	_ = iota
	//                1         AS_SET: unordered set of ASes a route in the
	//                             UPDATE message has traversed
	asSet

	//                2         AS_SEQUENCE: ordered set of ASes a route in
	//                             the UPDATE message has traversed
	asSequence
)

var asPathName = map[int]string{
	1: "AS_SET",
	2: "AS_SEQUENCE",
}

//             The path segment length is a 1-octet length field,
//             containing the number of ASes (not the number of octets) in
//             the path segment value field.

//             The path segment value field contains one or more AS
//             numbers, each encoded as a 2-octet length field.

//             Usage of this attribute is defined in 5.1.2.

const (
	_ = iota
	origin
	asPath
	//          c) NEXT_HOP (Type Code 3):
	//             This is a well-known mandatory attribute that defines the
	//             (unicast) IP address of the router that SHOULD be used as
	//             the next hop to the destinations listed in the Network Layer
	//             Reachability Information field of the UPDATE message.
	//             Usage of this attribute is defined in 5.1.3.
	nextHop
	//          d) MULTI_EXIT_DISC (Type Code 4):
	//             This is an optional non-transitive attribute that is a
	//             four-octet unsigned integer.  The value of this attribute
	//             MAY be used by a BGP speaker's Decision Process to
	//             discriminate among multiple entry points to a neighboring
	//             autonomous system.
	//             Usage of this attribute is defined in 5.1.4.
	multiExitDisc
	//          e) LOCAL_PREF (Type Code 5):
	//             LOCAL_PREF is a well-known attribute that is a four-octet
	//             unsigned integer.  A BGP speaker uses it to inform its other
	//             internal peers of the advertising speaker's degree of
	//             preference for an advertised route.
	//             Usage of this attribute is defined in 5.1.5.
	localPref
	//          f) ATOMIC_AGGREGATE (Type Code 6)
	//             ATOMIC_AGGREGATE is a well-known discretionary attribute of
	//             length 0.
	//             Usage of this attribute is defined in 5.1.6.
	atomicAggregate
	//          g) AGGREGATOR (Type Code 7)
	//             AGGREGATOR is an optional transitive attribute of length 6.
	//             The attribute contains the last AS number that formed the
	//             aggregate route (encoded as 2 octets), followed by the IP
	//             address of the BGP speaker that formed the aggregate route
	//             (encoded as 4 octets).  This SHOULD be the same address as
	//             the one used for the BGP Identifier of the speaker.
	aggregator
)

var pathAttributeName = map[int]string{
	1: "ORIGIN",
	2: "AS_PATH",
	3: "NEXT_HOP",
	4: "MULTI_EXIT_DISC",
	5: "LOCAL_PREF",
	6: "ATOMIC_AGGREGATE",
	7: "AGGREGATOR",
}

//             Usage of this attribute is defined in 5.1.7.

//       Network Layer Reachability Information:
//          This variable length field contains a list of IP address
//          prefixes.  The length, in octets, of the Network Layer
//          Reachability Information is not encoded explicitly, but can be
//          calculated as:
//                UPDATE message Length - 23 - Total Path Attributes Length
//                - Withdrawn Routes Length
//          where UPDATE message Length is the value encoded in the fixed-
//          size BGP header, Total Path Attribute Length, and Withdrawn
//          Routes Length are the values encoded in the variable part of
//          the UPDATE message, and 23 is a combined length of the fixed-
//          size BGP header, the Total Path Attribute Length field, and the
//          Withdrawn Routes Length field.
//          Reachability information is encoded as one or more 2-tuples of
//          the form <length, prefix>, whose fields are described below:
//                   +---------------------------+
//                   |   Length (1 octet)        |
//                   +---------------------------+
//                   |   Prefix (variable)       |
//                   +---------------------------+
//          The use and the meaning of these fields are as follows:
type nlri struct {
	//          a) Length:
	//             The Length field indicates the length in bits of the IP
	//             address prefix.  A length of zero indicates a prefix that
	//             matches all IP addresses (with prefix, itself, of zero
	//             octets).
	length byte

	//          b) Prefix:
	//             The Prefix field contains an IP address prefix, followed by
	//             enough trailing bits to make the end of the field fall on an
	//             octet boundary.  Note that the value of the trailing bits is
	//             irrelevant.
	prefix []byte
}

func newNLRI(length int, prefix net.IP) nlri {
	n := nlri{
		length: byte(length),
		prefix: packPrefix(length, prefix),
	}
	return n
}

func packPrefix(length int, ip net.IP) []byte {
	ip4 := ip.To4()
	bs := []byte{ip4[0], ip4[1], ip4[2], ip4[3]}
	return bs[:int(math.Ceil(math.Max(float64(length)/8.0, 1)))]
}

func readNLRI(bs []byte) *nlri {
	buf := bytes.NewBuffer(bs)
	nlri := new(nlri)
	nlri.length = stream.ReadByte(buf)
	nlri.prefix = stream.ReadBytes(int(nlri.length), buf)
	return nlri
}

func (n nlri) valid() (*notificationMessage, bool) {
	return nil, true
}

func (n nlri) bytes() []byte {
	// TODO: Implement me
	return []byte{}
}

//    The minimum length of the UPDATE message is 23 octets -- 19 octets
//    for the fixed header + 2 octets for the Withdrawn Routes Length + 2
//    octets for the Total Path Attribute Length (the value of Withdrawn
//    Routes Length is 0 and the value of Total Path Attribute Length is
//    0).
const minUpdateMessageLength = 23

//    An UPDATE message can advertise, at most, one set of path attributes,
//    but multiple destinations, provided that the destinations share these
//    attributes.  All path attributes contained in a given UPDATE message
//    apply to all destinations carried in the NLRI field of the UPDATE
//    message.

//    An UPDATE message can list multiple routes that are to be withdrawn
//    from service.  Each such route is identified by its destination
//    (expressed as an IP prefix), which unambiguously identifies the route
//    in the context of the BGP speaker - BGP speaker connection to which
//    it has been previously advertised.

//    An UPDATE message might advertise only routes that are to be
//    withdrawn from service, in which case the message will not include
//    path attributes or Network Layer Reachability Information.
//    Conversely, it may advertise only a feasible route, in which case the
//    WITHDRAWN ROUTES field need not be present.

//    An UPDATE message SHOULD NOT include the same address prefix in the
//    WITHDRAWN ROUTES and Network Layer Reachability Information fields.
//    However, a BGP speaker MUST be able to process UPDATE messages in
//    this form.  A BGP speaker SHOULD treat an UPDATE message of this form
//    as though the WITHDRAWN ROUTES do not contain the address prefix.

// 4.4.  KEEPALIVE Message Format

//    BGP does not use any TCP-based, keep-alive mechanism to determine if
//    peers are reachable.  Instead, KEEPALIVE messages are exchanged
//    between peers often enough not to cause the Hold Timer to expire.  A
//    reasonable maximum time between KEEPALIVE messages would be one third
//    of the Hold Time interval.  KEEPALIVE messages MUST NOT be sent more
//    frequently than one per second.  An implementation MAY adjust the
//    rate at which it sends KEEPALIVE messages as a function of the Hold
//    Time interval.
const minKeepaliveInterval = 1 * time.Second

//    If the negotiated Hold Time interval is zero, then periodic KEEPALIVE
//    messages MUST NOT be sent.

//    A KEEPALIVE message consists of only the message header and has a
//    length of 19 octets.
type keepaliveMessage struct{}

func newKeepaliveMessage() keepaliveMessage {
	return keepaliveMessage{}
}

func readKeepalive(message []byte) *keepaliveMessage {
	// Related events
	//if len(message) != 0 {
	// Send a notification
	//	return nil, newNotificationMessage(messageHeaderError, badMessageLength, nil)
	//}
	// Note: this should occur elsewhere
	//f.sendEvent(keepAliveMsg)
	return &keepaliveMessage{}
}

func (k keepaliveMessage) valid() (*notificationMessage, bool) {
	// TODO: Implement me
	return nil, true
}

func (k keepaliveMessage) bytes() []byte {
	return []byte{}
}

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
