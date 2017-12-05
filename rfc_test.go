package kbgp

import (
	"bytes"
	"fmt"
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
	fmt.Println("Creating a new FSM")
	f := new(fsm)
	fmt.Println("Adding a peer to it")
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	fmt.Println("Creating a mock connection")
	f.peer.conn = newConn(raw)
	// Add mock net.Conn to fsm
	// Write raw to it
	fmt.Println("Reading the message header and the message")
	header, message := f.readMessage()
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
	fmt.Println("Creating a new FSM")
	f := new(fsm)
	fmt.Println("Adding a peer to it")
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	fmt.Println("Creating a mock connection")
	f.peer.conn = newConn(raw)
	fmt.Println("Reading the message header and the message")
	header, message := f.readMessage()
	if len(message) != 9 {
		t.Error("Expected message length to be 0 but got", len(message))
	}
	if header.messageType != 0x01 {
		t.Errorf("Expected the message type to be %v but got %d", raw[18], header.messageType)
	}
	open := f.readOpen(message)
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

func TestReadKeepalive(t *testing.T) {
	raw := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // Marker
		0x00, 0x00, // Length
		0x04, // Type (Open)
	}
	fmt.Println("Creating a new FSM")
	f := new(fsm)
	fmt.Println("Adding a peer to it")
	f.peer = newPeer(1, net.ParseIP("1.2.3.4"))
	fmt.Println("Creating a mock connection")
	f.peer.conn = newConn(raw)
	fmt.Println("Reading the message header and the message")
	header, message := f.readMessage()
	if len(message) != 0 {
		t.Error("Expected message length to be 0 but got", len(message))
	}
	if header.messageType != keepalive {
		t.Errorf("Expected the message type to be %d but got %d", keepalive, header.messageType)
	}
	k := f.readKeepalive(message)
	if k == nil {
		t.Errorf("Did not expect keepalive to be nil")
	}
}

// TODO: Mock net.Conn
type conn struct {
	bs []byte
}

func newConn(bs []byte) *conn {
	fmt.Printf("Creating a new connection with %d bytes of data\n", len(bs))
	fmt.Println("The data is", bs)
	return &conn{bs: bs}
}

func (c *conn) Read(b []byte) (n int, err error) {
	count := copy(b, c.bs)
	fmt.Println("Read copied", count, "bytes")
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
