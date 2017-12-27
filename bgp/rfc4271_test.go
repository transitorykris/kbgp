package kbgp

import (
	"bytes"
	"net"
	"testing"
	"time"
	"unsafe"
)

func TestMarker(t *testing.T) {
	m := marker()
	if len(m) != markerLength {
		t.Errorf("Expected marker length %d but got %d", markerLength, len(m))
	}
	for i, v := range m {
		if v != 0xFF {
			t.Errorf("Expected all bits to be 1, got %d at position %d", v, i)
		}
	}
}

func TestIsValidHoldTime(t *testing.T) {
	//if isValidHoldTime((maxHoldTime + 1) * time.Second) {
	//	t.Errorf("Expected maxHoldTime+1 to be an invalid hold time")
	//}
	if isValidHoldTime(1 * time.Second) {
		t.Errorf("Expected 1 seconds to be an invalid hold time")
	}
	if isValidHoldTime(2 * time.Second) {
		t.Errorf("Expected 2 seconds to be an invalid hold time")
	}
	if !isValidHoldTime(0 * time.Second) {
		t.Errorf("Expected 0 seconds to be a valid hold time")
	}
	if !isValidHoldTime(3 * time.Second) {
		t.Errorf("Expected 3 seconds to be a valid hold time")
	}
}

func TestPackPrefix(t *testing.T) {
	b := packPrefix(32, net.ParseIP("1.2.3.4"))
	if len(b) != 4 {
		t.Errorf("Expected 1.2.3.4/32 to be 4 bytes but got %d", b)
	}
	b = packPrefix(25, net.ParseIP("1.2.3.4"))
	if len(b) != 4 {
		t.Errorf("Expected 1.2.3.4/25 to be 4 bytes but got %d", b)
	}
	b = packPrefix(24, net.ParseIP("1.2.3.4"))
	if len(b) != 3 {
		t.Errorf("Expected 1.2.3.4/24 to be 3 bytes but got %d", b)
	}
	b = packPrefix(16, net.ParseIP("1.2.3.4"))
	if len(b) != 2 {
		t.Errorf("Expected 1.2.3.4/16 to be 2 bytes but got %d", b)
	}
	b = packPrefix(8, net.ParseIP("1.2.3.4"))
	if len(b) != 1 {
		t.Errorf("Expected 1.2.3.4/8 to be 2 bytes but got %d", b)
	}
	b = packPrefix(1, net.ParseIP("1.2.3.4"))
	if len(b) != 1 {
		t.Errorf("Expected 1.2.3.4/1 to be 1 bytes but got %d", b)
	}
	b = packPrefix(0, net.ParseIP("1.2.3.4"))
	if len(b) != 1 {
		t.Errorf("Expected 1.2.3.4/0 to be 1 bytes but got %d", b)
	}
}

func TestNewNLRI(t *testing.T) {
	n := newNLRI(23, net.ParseIP("1.2.3.4"))
	if n.length != 23 {
		t.Errorf("Expected length to be 23 but got %d", n.length)
	}
	if bytes.Compare(n.prefix, []byte{1, 2, 3}) != 0 {
		t.Errorf("Expected bytes to be [1,2,3] but got %v", n.prefix)
	}
}

func TestNewKeepalive(t *testing.T) {
	k := newKeepaliveMessage()
	if unsafe.Sizeof(k) != 0 {
		t.Errorf("Keepalive messages must have 0 length")
	}
}

func TestNewNotificationMessage(t *testing.T) {
	data := []byte{0x03, 0x14, 0x25}
	n := newNotificationMessage(messageHeaderError, badMessageType, data)
	if n.code != messageHeaderError {
		t.Errorf("Expected error code %d but got %d", messageHeaderError, n.code)
	}
	if n.subcode != badMessageType {
		t.Errorf("Expected error subcode %d but got %d", badMessageType, n.subcode)
	}
	if bytes.Compare(n.data, data) != 0 {
		t.Errorf("Expected data %v but got %v", data, n.data)
	}
}

func TestNewFSM(t *testing.T) {
	f := newFSM(1, net.ParseIP("1.2.3.4"))
	if f.state != idle {
		t.Errorf("New FSMs start in the idle state, instead it's state %d", f.state)
	}
}

func TestJitter(t *testing.T) {
	// This needs to be a random value [.75,1) so we'll just run this loop
	// a bunch of times and hope.
	const min float64 = 750
	const max float64 = 1000
	for i := 0; i < 1000; i++ {
		j := jitter()
		if j < time.Duration(min)*time.Millisecond || j > time.Duration(max)*time.Millisecond {
			t.Errorf("Jitter must be [.75,1) seconds, we got %v", j)
		}
	}
}

func TestInitialize(t *testing.T) {
	f := new(fsm)
	f.peer = new(peer)
	f.initialize()
	if f.peer.adjRIBIn == nil {
		t.Errorf("Expected adjRIBIn to be non-nil")
	}
	if f.peer.adjRIBOut == nil {
		t.Errorf("Expected adjRIBOut to be non-nil")
	}
}

func TestRelease(t *testing.T) {
	f := new(fsm)
	f.peer = new(peer)
	f.initialize()
	f.release()
	if f.peer.adjRIBIn != nil {
		t.Errorf("Expected adjRIBIn to be nil but got %v", f.peer.adjRIBIn)
	}
	if f.peer.adjRIBOut != nil {
		t.Errorf("Expected adjRIBOut to be nil but got %v", f.peer.adjRIBOut)
	}
}

func TestReadMessage(t *testing.T) {
	raw := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // Marker
		0x00, 0x00, // Length
		0x01, // Type
	}
	f := new(fsm)
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	f.peer.conn = newConn(raw)
	// Add mock net.Conn to fsm
	// Write raw to it
	header, message := readMessage(f.peer.conn)
	// Check that the header has the correct marker, and expected length and type
	if bytes.Compare(header.marker[:], raw[:16]) != 0 {
		t.Errorf("Header marker should be %v but got %v", raw[:16], header.marker)
	}
	// Check that the message length is equal to the expected length
	if len(message) != 0 {
		t.Errorf("Expected message length to be 0 but got %d", len(message))
	}
	if header.messageType != 0x01 {
		t.Errorf("Expected the message type to be %v but got %d", raw[18], header.messageType)
	}
}

func TestMessageHeaderBytes(t *testing.T) {

}

func TestReadOpen(t *testing.T) {
	raw := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // Marker
		0x00, 0x09, // Length
		0x01,       // Type (Open)
		0x04,       // Version
		0x00, 0x01, // My Autonomous System,
		0x00, 0x00, // 0 second Hold time
		0x01, 0x02, 0x03, 0x04, // BGP Identifier
		0x00, // 0 length = no optional parameters
	}
	f := new(fsm)
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	f.peer.conn = newConn(raw)
	header, message := readMessage(f.peer.conn)
	if len(message) != 9 {
		t.Error("Expected message length to be 0 but got", len(message))
	}
	if header.messageType != 0x01 {
		t.Errorf("Expected the message type to be %v but got %d", raw[18], header.messageType)
	}
	open := readOpen(message)
	if open.version != 4 {
		t.Error("Expected BGP version to be 4 but got", open.version)
	}
	if open.myAS != 1 {
		t.Error("Expected myAS to be 1 but got", open.myAS)
	}
	if open.holdTime != 0 {
		t.Error("Expected hold time to be 0 but got", open.holdTime)
	}
	if open.bgpIdentifier != 16909060 {
		t.Error("Expected bgp identifier to be 16909060 but got", open.bgpIdentifier)
	}
	if open.optParmLen != 0 {
		t.Error("Expected optional parameters length to be 0 but got", open.optParmLen)
	}
	if len(open.optParameters) != 0 {
		t.Error("Expected no optional parameters but got", open.optParameters)
	}
}

func TestOpenBytes(t *testing.T) {
	o := openMessage{
		version:       version,
		myAS:          1,
		holdTime:      3,
		bgpIdentifier: 16909060,
		optParmLen:    0,
		optParameters: []byte{},
	}
	bs := o.bytes()
	raw := []byte{
		0x04,       // Version
		0x00, 0x01, // My Autonomous System,
		0x00, 0x03, // 0 second Hold time
		0x01, 0x02, 0x03, 0x04, // BGP Identifier
		0x00, // 0 length = no optional parameters
	}
	if bytes.Compare(bs, raw) != 0 {
		t.Errorf("Expected open message %v but got %v", raw, bs)
	}
}

func TestReadOptionalParameters(t *testing.T) {
}

func TestReadKeepalive(t *testing.T) {
	raw := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // Marker
		0x00, 0x00, // Length
		0x04, // Type (Open)
	}
	f := new(fsm)
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	f.peer.conn = newConn(raw)
	header, message := readMessage(f.peer.conn)
	if len(message) != 0 {
		t.Error("Expected message length to be 0 but got", len(message))
	}
	if header.messageType != keepalive {
		t.Errorf("Expected the message type to be %d but got %d", keepalive, header.messageType)
	}
	k := readKeepalive(message)
	if k == nil {
		t.Errorf("Did not expect keepalive to be nil")
	}
}

func TestKeepaliveBytes(t *testing.T) {

}

func TestReadNotification(t *testing.T) {
	raw := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // Marker
		0x00, 0x02, // Length
		0x03, // Type (Notification)
		0x01, // OPEN Message Error
		0x02, // Bad Message Length
		// No data
	}
	f := new(fsm)
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	f.peer.conn = newConn(raw)
	header, message := readMessage(f.peer.conn)
	if len(message) != 2 {
		t.Error("Expected message length to be 2 but got", len(message))
	}
	if header.messageType != notification {
		t.Errorf("Expected the message type to be %d but got %d", notification, header.messageType)
	}
	k := readNotification(message)
	if k == nil {
		t.Errorf("Did not expect keepalive to be nil")
	}

	raw = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // Marker
		0x00, 0x06, // Length
		0x03,                   // Type (Notification)
		0x03,                   // UPDATE Message Error
		0x11,                   // Malformed AS_PATH
		0x12, 0x34, 0x56, 0x7a, // Some random data
	}
	f = new(fsm)
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	f.peer.conn = newConn(raw)
	header, message = readMessage(f.peer.conn)
	if len(message) != 6 {
		t.Error("Expected message length to be 6 but got", len(message))
	}
	if header.messageType != notification {
		t.Errorf("Expected the message type to be %d but got %d", notification, header.messageType)
	}
	k = readNotification(message)
	if k == nil {
		t.Errorf("Did not expect keepalive to be nil")
	}
}

func TestNotificationBytes(t *testing.T) {

}

func TestReadUpdate(t *testing.T) {
	raw := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // Marker
		0x00, 0x04, // Length
		0x02, // Type (Update)
		// Update with no data? whhhaaa :D
	}
	f := new(fsm)
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	f.peer.conn = newConn(raw)
	header, message := readMessage(f.peer.conn)
	if len(message) != 4 {
		t.Error("Expected message length to be 4 but got", len(message))
	}
	if header.messageType != update {
		t.Errorf("Expected the message type to be %d but got %d", update, header.messageType)
	}
	k := readUpdate(message)
	if k == nil {
		t.Errorf("Did not expect keepalive to be nil")
	}
	if k.withdrawnRoutesLength != 0 {
		t.Errorf("Expected withdrawn routes length to be 0 but got", k.withdrawnRoutesLength)
	}
	if k.pathAttributesLength != 0 {
		t.Errorf("Expected path attributes length to be 0 but got", k.pathAttributesLength)
	}
}

func TestUpdateBytes(t *testing.T) {

}

func TestReadWithdrawnRoutes(t *testing.T) {
}

func TestWithdrawnRouteBytes(t *testing.T) {

}

func TestReadPathAttributes(t *testing.T) {
}

func TestPathAttributeBytes(t *testing.T) {

}

func TestOptionalAttribute(t *testing.T) {
	a := attributeType{flags: optional}
	if !a.optional() {
		t.Error("Expected attribute to be optional")
	}
	if a.wellKnown() {
		t.Error("Did not expect attribute to be well-known")
	}

	a = attributeType{flags: optional | transitive | partial | extendedLength}
	if !a.optional() {
		t.Error("Expected attribute to be optional")
	}
	if a.wellKnown() {
		t.Error("Did not expect attribute to be well-known")
	}
}

func TestSetOptional(t *testing.T) {
	a := attributeType{flags: wellKnown}
	a.setOptional()
	if a.wellKnown() {
		t.Error("Expected the attribute to be optional")
	}

	a = attributeType{flags: optional}
	a.setOptional()
	if a.wellKnown() {
		t.Error("Expected the attribute to be optional")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestWellKnownAttribute(t *testing.T) {
	a := attributeType{flags: wellKnown}
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}
	if a.optional() {
		t.Error("Did not expect attribute to be optional")
	}

	a = attributeType{flags: wellKnown | transitive | partial | extendedLength}
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}
	if a.optional() {
		t.Error("Did not expect attribute to be optional")
	}
}

func TestSetWellKnown(t *testing.T) {
	a := attributeType{flags: optional}
	a.setWellKnown()
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}

	a = attributeType{flags: wellKnown}
	a.setWellKnown()
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestTransitiveAttribute(t *testing.T) {
	a := attributeType{flags: transitive}
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}
	if a.nonTransitive() {
		t.Error("Did not expect attribute to be non-transitive")
	}

	a = attributeType{flags: optional | transitive | partial | extendedLength}
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}
	if a.nonTransitive() {
		t.Error("Did not expect attribute to be non-transitive")
	}
}

func TestSetTransitive(t *testing.T) {
	a := attributeType{flags: nonTransitive}
	a.setTransitive()
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}

	a = attributeType{flags: transitive}
	a.setTransitive()
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestNonTransitiveAttribute(t *testing.T) {
	a := attributeType{flags: nonTransitive}
	if !a.nonTransitive() {
		t.Error("Expected attribute to be non-transitive")
	}
	if a.transitive() {
		t.Error("Did not expect attribute to be transitive")
	}

	a = attributeType{flags: optional | nonTransitive | partial | extendedLength}
	if !a.nonTransitive() {
		t.Error("Expected attribute to be transitive")
	}
	if a.transitive() {
		t.Error("Did not expect attribute to be transitive")
	}
}

func TestSetNonTransitive(t *testing.T) {
	a := attributeType{flags: transitive}
	a.setNonTransitive()
	if !a.nonTransitive() {
		t.Error("Expected attribute to be non-transitive")
	}

	a = attributeType{flags: nonTransitive}
	a.setNonTransitive()
	if !a.nonTransitive() {
		t.Error("Expected attribute to be non-transitive")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestPartialAttribute(t *testing.T) {
	a := attributeType{flags: partial}
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}
	if a.complete() {
		t.Error("Did not expect attribute to be complete")
	}

	a = attributeType{flags: optional | transitive | partial | extendedLength}
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}
	if a.complete() {
		t.Error("Did not expect attribute to be complete")
	}
}

func TestSetPartial(t *testing.T) {
	a := attributeType{flags: complete}
	a.setPartial()
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}

	a = attributeType{flags: partial}
	a.setPartial()
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestCompleteAttribute(t *testing.T) {
	a := attributeType{flags: complete}
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be partial")
	}

	a = attributeType{flags: optional | transitive | complete | extendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be partial")
	}
}

func TestSetComplete(t *testing.T) {
	a := attributeType{flags: partial}
	a.setComplete()
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}

	a = attributeType{flags: complete}
	a.setComplete()
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestExtendedLengthAttribute(t *testing.T) {
	a := attributeType{flags: extendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be not extended length")
	}

	a = attributeType{flags: optional | transitive | complete | extendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be not extended length")
	}
}

func TestSetExtendedLength(t *testing.T) {
	a := attributeType{flags: notExtendedLength}
	a.setExtendedLength()
	if !a.extendedLength() {
		t.Error("Expected attribute to be extended length")
	}

	a = attributeType{flags: extendedLength}
	a.setExtendedLength()
	if !a.extendedLength() {
		t.Error("Expected attribute to be extended length")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestNotExtendedLengthAttribute(t *testing.T) {
	a := attributeType{flags: notExtendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be not extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be  extended length")
	}

	a = attributeType{flags: optional | transitive | complete | notExtendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be not extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be extended length")
	}
}

func TestSetNotExtendedLength(t *testing.T) {
	a := attributeType{flags: extendedLength}
	a.setNotExtendedLength()
	if !a.nonextendedLength() {
		t.Error("Expected attribute to be extended length")
	}

	a = attributeType{flags: notExtendedLength}
	a.setNotExtendedLength()
	if !a.nonextendedLength() {
		t.Error("Expected attribute to be extended length")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestReadNLRI(t *testing.T) {
}

func TestNLRIBytes(t *testing.T) {

}

// TODO: Mock net.Conn
type conn struct {
	bs []byte
}

func newConn(bs []byte) *conn {
	return &conn{bs: bs}
}

func (c *conn) Read(b []byte) (n int, err error) {
	count := copy(b, c.bs)
	// Remove those bytes from our mock connection's buffer
	c.bs = c.bs[count:]
	return len(b), nil
}

func (c *conn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) LocalAddr() net.Addr {
	return nil
}

func (c *conn) RemoteAddr() net.Addr {
	return nil
}

func (c *conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestNewAdjRIBIn(t *testing.T) {
	a := newAdjRIBIn()
	if a == nil {
		t.Error("Did not expect a new AdjRIBIn to be nil")
	}
	if len(a.routes) != 0 {
		t.Error("Expected 0 routes in the RIB but found", len(a.routes))
	}
}

// compareNLRI is a little bit ugly but will do for now
func compareNLRI(a nlri, b nlri) int {
	if a.length != b.length {
		return 1
	}
	if bytes.Compare(a.prefix, b.prefix) != 0 {
		return 2
	}
	return 0
}

func TestAdjRIBInAdd(t *testing.T) {
	a := newAdjRIBIn()
	n := newNLRI(24, net.ParseIP("10.1.2.0"))
	p := []pathAttribute{}
	a.add(n, p)
	if compareNLRI(a.routes[0].nlri, n) != 0 {
		t.Errorf("Expected the first route's nlri to be %v but found %v", n, a.routes[0].nlri)
	}

	a.add(n, p)
	if len(a.routes) != 1 {
		t.Error("Replaced a route but the size of the slice changed to", len(a.routes))
	}
}

func TestAdjRIBInRemove(t *testing.T) {
	a := newAdjRIBIn()
	n1 := newNLRI(24, net.ParseIP("10.1.2.0"))
	n2 := newNLRI(24, net.ParseIP("10.1.3.0"))
	n3 := newNLRI(24, net.ParseIP("10.1.4.0"))
	p := []pathAttribute{}
	a.add(n1, p)
	a.add(n2, p)
	a.add(n3, p)
	a.remove(withdrawnRoute{length: n1.length, prefix: n1.prefix})
	if len(a.routes) != 2 {
		t.Error("Expected the numnber of routes to be 2 but found", len(a.routes))
	}
	if compareNLRI(a.routes[0].nlri, n2) != 0 {
		t.Errorf("Expected the first prefix to be %v but got %v", n2.prefix, a.routes[0].nlri.prefix)
	}
	if compareNLRI(a.routes[1].nlri, n3) != 0 {
		t.Errorf("Expected the second prefix to be %v but got %v", n2.prefix, a.routes[0].nlri.prefix)
	}

	a = newAdjRIBIn()
	a.add(n1, p)
	a.add(n2, p)
	a.add(n3, p)
	a.remove(withdrawnRoute{length: n2.length, prefix: n2.prefix})
	if len(a.routes) != 2 {
		t.Error("Expected the numnber of routes to be 2 but found", len(a.routes))
	}
	if compareNLRI(a.routes[0].nlri, n1) != 0 {
		t.Errorf("Expected the first prefix to be %v but got %v", n1.prefix, a.routes[0].nlri.prefix)
	}
	if compareNLRI(a.routes[1].nlri, n3) != 0 {
		t.Errorf("Expected the second prefix to be %v but got %v", n3.prefix, a.routes[0].nlri.prefix)
	}

	a = newAdjRIBIn()
	a.add(n1, p)
	a.add(n2, p)
	a.add(n3, p)
	a.remove(withdrawnRoute{length: n3.length, prefix: n3.prefix})
	if len(a.routes) != 2 {
		t.Error("Expected the numnber of routes to be 2 but found", len(a.routes))
	}
	if compareNLRI(a.routes[0].nlri, n1) != 0 {
		t.Errorf("Expected the first prefix to be %v but got %v", n1.prefix, a.routes[0].nlri.prefix)
	}
	if compareNLRI(a.routes[1].nlri, n2) != 0 {
		t.Errorf("Expected the second prefix to be %v but got %v", n2.prefix, a.routes[0].nlri.prefix)
	}
}
