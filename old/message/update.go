package message

import (
	"bytes"
	"math"
	"net"

	"github.com/transitorykris/kbgp/stream"
)

// UPDATE messages are used to transfer routing information between BGP
// peers.  The information in the UPDATE message can be used to
// construct a graph that describes the relationships of the various
// Autonomous Systems.  By applying rules to be discussed, routing
// information loops and some other anomalies may be detected and
// removed from inter-AS routing.
// An UPDATE message is used to advertise feasible routes that share
// common path attributes to a peer, or to withdraw multiple unfeasible
// routes from service (see 3.1).  An UPDATE message MAY simultaneously
// advertise a feasible route and withdraw multiple unfeasible routes
// from service.  The UPDATE message always includes the fixed-size BGP
// header, and also includes the other fields, as shown below (note,
// some of the shown fields may not be present in every UPDATE message):
type UpdateMessage struct {
	withdrawnRoutesLength uint16
	withdrawnRoutes       []WithdrawnRoute
	pathAttributesLength  uint16
	pathAttributes        []PathAttribute
	nlris                 []NLRI
}

// This is a variable-length field that contains a list of IP
// address prefixes for the routes that are being withdrawn from
// service.  Each IP address prefix is encoded as a 2-tuple of the
// form <length, prefix>, whose fields are described below:
type WithdrawnRoute struct {
	// The Length field indicates the length in bits of the IP
	// address prefix.  A length of zero indicates a prefix that
	// matches all IP addresses (with prefix, itself, of zero
	// octets).
	length byte
	// The Prefix field contains an IP address prefix, followed by
	// the minimum number of trailing bits needed to make the end
	// of the field fall on an octet boundary.  Note that the value
	// of trailing bits is irrelevant.
	prefix []byte
}

// A variable-length sequence of path attributes is present in
// every UPDATE message, except for an UPDATE message that carries
// only the withdrawn routes.  Each path attribute is a triple
// <attribute type, attribute length, attribute value> of variable
// length.
type PathAttribute struct {
	attributeType   AttributeType
	attributeLength uint16
	attributeValue  []byte
}

// Attribute Type is a two-octet field that consists of the
// Attribute Flags octet, followed by the Attribute Type Code
// octet.
type AttributeType struct {
	flags byte
	code  byte
}

// The high-order bit (bit 0) of the Attribute Flags octet is the
// Optional bit.  It defines whether the attribute is optional (if
// set to 1) or well-known (if set to 0).
const optional = 1 << 7
const wellKnown = 0


func (a *attributeType) Optional() bool {
	return a.flags&optional == optional
}

func (a *attributeType) SetOptional() {
	a.flags = a.flags | optional
}

func (a *attributeType) WellKnown() bool {
	return a.flags&optional == wellKnown
}

func (a *attributeType) SetWellKnown() {
	a.flags = a.flags &^ optional
}


}


// The second high-order bit (bit 1) of the Attribute Flags octet
// is the Transitive bit.  It defines whether an optional
// attribute is transitive (if set to 1) or non-transitive (if set
// to 0).
// For well-known attributes, the Transitive bit MUST be set to 1.
// (See Section 5 for a discussion of transitive attributes.)
const transitive = 1 << 6
const nonTransitive = 0

func (a *attributeType) Transitive() bool {
	return a.flags&transitive == transitive
}

func (a *attributeType) SetTransitive() {
	a.flags = a.flags | transitive
}

func (a *attributeType) NonTransitive() bool {
	return a.flags&transitive == nonTransitive
}

func (a *attributeType) SetNonTransitive() {
	a.flags = a.flags &^ transitive
}


// The third high-order bit (bit 2) of the Attribute Flags octet
// is the Partial bit.  It defines whether the information
// contained in the optional transitive attribute is partial (if
// set to 1) or complete (if set to 0).  For well-known attributes
// and for optional non-transitive attributes, the Partial bit
// MUST be set to 0.
const partial = 1 << 5
const complete = 0

func (a *attributeType) Partial() bool {
	return a.flags&partial == partial
}

func (a *attributeType) SetPartial() {
	a.flags = a.flags | partial
}

func (a *attributeType) Complete() bool {
	return a.flags&partial == complete
}

func (a *attributeType) SetComplete() {
	a.flags = a.flags &^ partial
}


// The fourth high-order bit (bit 3) of the Attribute Flags octet
// is the Extended Length bit.  It defines whether the Attribute
// Length is one octet (if set to 0) or two octets (if set to 1).
const extendedLength = 1 << 4
const notExtendedLength = 0

func (a *attributeType) ExtendedLength() bool {
	return a.flags&extendedLength == extendedLength
}

func (a *attributeType) SetExtendedLength() {
	a.flags = a.flags | extendedLength
}

func (a *attributeType) NotExtendedLength() bool {
	return a.flags&extendedLength == notExtendedLength
}

func (a *attributeType) SetNotExtendedLength() {
	a.flags = a.flags &^ extendedLength
}

// ORIGIN is a well-known mandatory attribute that defines the
// origin of the path information.  The data octet can assume
// the following values:
const (
	// IGP - Network Layer Reachability Information
	// is interior to the originating AS
	igp = iota
	// EGP - Network Layer Reachability Information
	//learned via the EGP protocol [RFC904]
	egp
	// INCOMPLETE - Network Layer Reachability
	// Information learned by some other means
	incomplete
)

// Reverse value lookup for origin names
var originName = map[int]string{
	0: "IGP",
	1: "EGP",
	2: "INCOMPLETE",
}

// AS_PATH is a well-known mandatory attribute that is composed
// of a sequence of AS path segments.  Each AS path segment is
// represented by a triple <path segment type, path segment
// length, path segment value>.
// The path segment type is a 1-octet length field with the
// following values defined:
const (
	_ = iota
	// AS_SET: unordered set of ASes a route in the
	// UPDATE message has traversed
	asSet
	// AS_SEQUENCE: ordered set of ASes a route in
	// the UPDATE message has traversed
	asSequence
)

// Reverse value lookup for AS path value names
var asPathName = map[int]string{
	1: "AS_SET",
	2: "AS_SEQUENCE",
}

const (
	_ = iota
	origin
	asPath
	// This is a well-known mandatory attribute that defines the
	// (unicast) IP address of the router that SHOULD be used as
	// the next hop to the destinations listed in the Network Layer
	// Reachability Information field of the UPDATE message.
	// Usage of this attribute is defined in 5.1.3.
	nextHop
	// This is an optional non-transitive attribute that is a
	// four-octet unsigned integer.  The value of this attribute
	// MAY be used by a BGP speaker's Decision Process to
	// discriminate among multiple entry points to a neighboring
	// autonomous system.
	// Usage of this attribute is defined in 5.1.4.
	multiExitDisc
	// LOCAL_PREF is a well-known attribute that is a four-octet
	// unsigned integer.  A BGP speaker uses it to inform its other
	// internal peers of the advertising speaker's degree of
	// preference for an advertised route.
	// Usage of this attribute is defined in 5.1.5.
	localPref
	// ATOMIC_AGGREGATE is a well-known discretionary attribute of
	// length 0.
	atomicAggregate
	// AGGREGATOR is an optional transitive attribute of length 6.
	// The attribute contains the last AS number that formed the
	// aggregate route (encoded as 2 octets), followed by the IP
	// address of the BGP speaker that formed the aggregate route
	// (encoded as 4 octets).  This SHOULD be the same address as
	// the one used for the BGP Identifier of the speaker.
	aggregator
)

// Reverse value lookup for path attribute names
var pathAttributeName = map[int]string{
	1: "ORIGIN",
	2: "AS_PATH",
	3: "NEXT_HOP",
	4: "MULTI_EXIT_DISC",
	5: "LOCAL_PREF",
	6: "ATOMIC_AGGREGATE",
	7: "AGGREGATOR",
}

// This variable length field contains a list of IP address
// prefixes.  The length, in octets, of the Network Layer
// Reachability Information is not encoded explicitly, but can be
// calculated as:
//     UPDATE message Length - 23 - Total Path Attributes Length
//         - Withdrawn Routes Length
// where UPDATE message Length is the value encoded in the fixed-
// size BGP header, Total Path Attribute Length, and Withdrawn
// Routes Length are the values encoded in the variable part of
// the UPDATE message, and 23 is a combined length of the fixed-
// size BGP header, the Total Path Attribute Length field, and the
// Withdrawn Routes Length field.
// Reachability information is encoded as one or more 2-tuples of
// the form <length, prefix>, whose fields are described below:
type NLRI struct {
	// The Length field indicates the length in bits of the IP
	// address prefix.  A length of zero indicates a prefix that
	// matches all IP addresses (with prefix, itself, of zero
	// octets).
	length byte
	// The Prefix field contains an IP address prefix, followed by
	// enough trailing bits to make the end of the field fall on an
	// octet boundary.  Note that the value of the trailing bits is
	// irrelevant.
	prefix []byte
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

//       Total Path Attribute Length:
//          This 2-octet unsigned integer indicates the total length of the
//          Path Attributes field in octets.  Its value allows the length
//          of the Network Layer Reachability field to be determined as
//          specified below.

//          A value of 0 indicates that neither the Network Layer
//          Reachability Information field nor the Path Attribute field is
//          present in this UPDATE message.

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

//             The path segment length is a 1-octet length field,
//             containing the number of ASes (not the number of octets) in
//             the path segment value field.

//             The path segment value field contains one or more AS
//             numbers, each encoded as a 2-octet length field.

//             Usage of this attribute is defined in 5.1.2.

//             Usage of this attribute is defined in 5.1.7.

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

// 6.3.  UPDATE Message Error Handling

//    All errors detected while processing the UPDATE message MUST be
//    indicated by sending the NOTIFICATION message with the Error Code
//    UPDATE Message Error.  The error subcode elaborates on the specific
//    nature of the error.

//    Error checking of an UPDATE message begins by examining the path
//    attributes.  If the Withdrawn Routes Length or Total Attribute Length
//    is too large (i.e., if Withdrawn Routes Length + Total Attribute
//    Length + 23 exceeds the message Length), then the Error Subcode MUST
//    be set to Malformed Attribute List.

//    If any recognized attribute has Attribute Flags that conflict with
//    the Attribute Type Code, then the Error Subcode MUST be set to
//    Attribute Flags Error.  The Data field MUST contain the erroneous
//    attribute (type, length, and value).

//    If any recognized attribute has an Attribute Length that conflicts
//    with the expected length (based on the attribute type code), then the
//    Error Subcode MUST be set to Attribute Length Error.  The Data field
//    MUST contain the erroneous attribute (type, length, and value).

//    If any of the well-known mandatory attributes are not present, then
//    the Error Subcode MUST be set to Missing Well-known Attribute.  The
//    Data field MUST contain the Attribute Type Code of the missing,
//    well-known attribute.

//    If any of the well-known mandatory attributes are not recognized,
//    then the Error Subcode MUST be set to Unrecognized Well-known
//    Attribute.  The Data field MUST contain the unrecognized attribute
//    (type, length, and value).

//    If the ORIGIN attribute has an undefined value, then the Error Sub-
//    code MUST be set to Invalid Origin Attribute.  The Data field MUST
//    contain the unrecognized attribute (type, length, and value).

//    If the NEXT_HOP attribute field is syntactically incorrect, then the
//    Error Subcode MUST be set to Invalid NEXT_HOP Attribute.  The Data
//    field MUST contain the incorrect attribute (type, length, and value).
//    Syntactic correctness means that the NEXT_HOP attribute represents a
//    valid IP host address.

//    The IP address in the NEXT_HOP MUST meet the following criteria to be
//    considered semantically correct:

//       a) It MUST NOT be the IP address of the receiving speaker.

//       b) In the case of an EBGP, where the sender and receiver are one
//          IP hop away from each other, either the IP address in the
//          NEXT_HOP MUST be the sender's IP address that is used to
//          establish the BGP connection, or the interface associated with
//          the NEXT_HOP IP address MUST share a common subnet with the
//          receiving BGP speaker.

//    If the NEXT_HOP attribute is semantically incorrect, the error SHOULD
//    be logged, and the route SHOULD be ignored.  In this case, a
//    NOTIFICATION message SHOULD NOT be sent, and the connection SHOULD
//    NOT be closed.

//    The AS_PATH attribute is checked for syntactic correctness.  If the
//    path is syntactically incorrect, then the Error Subcode MUST be set
//    to Malformed AS_PATH.

//    If the UPDATE message is received from an external peer, the local
//    system MAY check whether the leftmost (with respect to the position
//    of octets in the protocol message) AS in the AS_PATH attribute is
//    equal to the autonomous system number of the peer that sent the
//    message.  If the check determines this is not the case, the Error
//    Subcode MUST be set to Malformed AS_PATH.

//    If an optional attribute is recognized, then the value of this
//    attribute MUST be checked.  If an error is detected, the attribute
//    MUST be discarded, and the Error Subcode MUST be set to Optional
//    Attribute Error.  The Data field MUST contain the attribute (type,
//    length, and value).

//    If any attribute appears more than once in the UPDATE message, then
//    the Error Subcode MUST be set to Malformed Attribute List.

//    The NLRI field in the UPDATE message is checked for syntactic
//    validity.  If the field is syntactically incorrect, then the Error
//    Subcode MUST be set to Invalid Network Field.

//    If a prefix in the NLRI field is semantically incorrect (e.g., an
//    unexpected multicast IP address), an error SHOULD be logged locally,
//    and the prefix SHOULD be ignored.

//    An UPDATE message that contains correct path attributes, but no NLRI,
//    SHALL be treated as a valid UPDATE message.
