package bgp

import (
	"net"
	"time"
)

// Version is a BGP version implemented by a speaker
type Version uint8

// ASN is an autonomous system number
type ASN uint16

// Identifier is used by a speaker and typically represents an IPv4 address
type Identifier uint32

// Speaker is an individual BGP speaking router
type Speaker struct {
	ASN    ASN
	ID     Identifier
	Peers  []Peer
	LocRIB RIB
}

// NLRI is an IP prefix
type NLRI net.IPNet

// PathAttribute is a TLV of attributes associated with an NLRI
type PathAttribute struct {
	Type   int
	Length uint16
	Value  []byte
}

// Route is an NLRI and its path attributes
type Route struct {
	NLRI           NLRI
	PathAttributes []PathAttribute
}

// A Peer is a session and a finite state machine
type Peer struct {
	Session Session
	FSM     FSM
}

// A Session provides an interface to remote BGP peers
type Session interface {
	// Connect this peer to an established TCP session
	Connect(net.Conn)
	// Shutdown the peer
	Shutdown()
	// Announce a prefix to this peer
	Announce(Route)
	// Withdraw a prefix from this peer
	Withdraw(NLRI)
}

// A Policer is used to apply policy to an announcement
type Policer interface {
	// Apply this policy to the announcement
	Apply(NLRI, []PathAttribute) bool
}

// Event is an administrative event in which the operator interface
// and BGP Policy engine signal the BGP-finite state machine to start or
// stop the BGP state machine.
type Event int

const (
	_ = iota

	// Administrative Events

	// ManualStart - Local system administrator manually starts the peer connection.
	ManualStart
	// ManualStop - Local system administrator manually stops the peer connection.
	ManualStop
	// AutomaticStart - Local system automatically starts the BGP connection.
	AutomaticStart
	// ManualStartWithPassiveTCPEstablishment - Local system administrator manually
	// starts the peer
	// connection, but has PassiveTcpEstablishment enabled.
	ManualStartWithPassiveTCPEstablishment
	// AutomaticStartWithPassiveTCPEstablishment - Local system automatically starts
	// the BGP connection with the PassiveTcpEstablishment enabled.
	AutomaticStartWithPassiveTCPEstablishment
	// AutomaticStartWithDampPeerOscillations -  Local system automatically starts
	// the BGP peer connection with peer oscillation damping enabled.
	AutomaticStartWithDampPeerOscillations
	// AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment - Local
	// system automatically starts the BGP peer connection with peer oscillation
	// damping enabled and PassiveTcpEstablishment enabled.
	AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment
	// AutomaticStop - Local system automatically stops the BGP connection.
	// An example of an automatic stop event is exceeding the number of prefixes for
	// a given peer and the local system automatically disconnecting the peer.
	AutomaticStop

	// Timer Events

	// ConnectRetryTimerExpires - An event generated when the ConnectRetryTimer expires.
	ConnectRetryTimerExpires
	// HoldTimerExpires - An event generated when the HoldTimer expires.
	HoldTimerExpires
	// KeepaliveTimerExpires - An event generated when the KeepaliveTimer expires.
	KeepaliveTimerExpires
	// DelayOpenTimerExpires - An event generated when the DelayOpenTimer expires.
	DelayOpenTimerExpires
	// IdleHoldTimerExpires - An event generated when the IdleHoldTimer expires,
	// indicating that the BGP connection has completed waiting for the back-off
	// period to prevent BGP peer oscillation.
	IdleHoldTimerExpires

	// TCP Connection-Based Events

	// TCPConnectionValid - Event indicating the local system reception of a
	// TCP connection request with a valid source IP address, TCP port, destination
	// IP address, and TCP Port.
	TCPConnectionValid
	// TCPCRInvalid -  Event indicating the local system reception of a TCP
	// connection request with either an invalid source address or port number, or
	// an invalid destination address or port number.
	TCPCRInvalid
	// TCPCRAcked - Event indicating the local system's request to establish a TCP
	// connection to the remote peer. The local system's TCP connection sent a TCP
	// SYN, received a TCP SYN/ACK message, and sent a TCP ACK.
	TCPCRAcked
	// TCPConnectionConfirmed - Event indicating that the local system has received
	// a confirmation that the TCP connection has been established by the remote site.
	// The remote peer's TCP engine sent a TCP SYN. The local peer's TCP engine sent
	// a SYN, ACK message and now has received a final ACK.
	TCPConnectionConfirmed
	// TCPConnectionFails - Event indicating that the local system has received
	// a TCP connection failure notice. The remote BGP peer's TCP machine could have
	// sent a FIN.  The local peer would respond with a FIN-ACK. Another possibility
	// is that the local peer indicated a timeout in the TCP connection and downed
	// the connection.
	TCPConnectionFails

	// BGP Message-Based Events

	// BGPOpen - An event is generated when a valid OPEN message has been received.
	BGPOpen
	// BGPOpenWithDelayOpenTimerRunning - An event is generated when a valid OPEN
	// message has been received for a peer that has a successfully established
	// transport connection and is currently delaying the sending of a BGP open message.
	BGPOpenWithDelayOpenTimerRunning
	// BGPHeaderErr - An event is generated when a received BGP message header is
	// not valid.
	BGPHeaderErr
	// BGPOpenMsgErr - An event is generated when an OPEN message has been received
	// with errors.
	BGPOpenMsgErr
	// OpenCollisionDump - An event generated administratively when a connection
	// collision has been detected while processing an incoming OPEN message and
	// this connection has been selected to be disconnected.
	OpenCollisionDump
	// NotifMsgVerErr - An event is generated when a NOTIFICATION message with
	// "version error" is received.
	NotifMsgVerErr
	// NotifMsg - An event is generated when a NOTIFICATION message is received and
	// the error code is anything but "version error".
	NotifMsg
	// KeepAliveMsg - An event is generated when a KEEPALIVE message is received.
	KeepAliveMsg
	// UpdateMsg - An event is generated when a valid UPDATE message is received.
	UpdateMsg
	// UpdateMsgErr - An event is generated when an invalid UPDATE message is received.
	UpdateMsgErr
)

// State is the state of an FSM
type State int

const (
	// Idle - Initially, the BGP peer FSM is in the Idle state.
	Idle = iota
	// Connect - In this state, BGP FSM is trying to acquire a peer by listening
	// for, and accepting, a TCP connection.
	Connect
	// Active - In this state, BGP FSM is trying to acquire a peer by listening
	// for, and accepting, a TCP connection.
	Active
	// OpenSent - In this state, BGP FSM waits for an OPEN message from its peer.
	OpenSent
	// OpenConfirm - In this state, BGP waits for a KEEPALIVE or NOTIFICATION message.
	OpenConfirm
	// Established State - In the Established state, the BGP FSM can exchange UPDATE,
	// NOTIFICATION, and KEEPALIVE messages with its peer.
	Established
)

// An FSM is a BGP finite state machine
type FSM interface {
	// Send an event into the FSM
	Send(Event)
	// Get the state of the FSM
	State() State
}

// A RIB is a routing information base
type RIB interface {
	Inject(Route)
	Remove(NLRI)
	Lookup(net.IPNet) Route
	SetPolicy(Policer)
	Dump() <-chan Route
}

// Counter is a simple counter
type Counter interface {
	Reset()
	Increment()
	Value() uint64
}

// Timer is a simple timer that can be reset
type Timer interface {
	Reset(time.Duration)
	Stop()
	Running() bool
}

// BGP Timers

// ConnectRetryTime is a mandatory FSM attribute that stores the initial
// value for the ConnectRetryTimer.  The suggested default value for the
// ConnectRetryTime is 120 seconds.
const defaultConnectRetryTime = 120 * time.Second

// HoldTime is a mandatory FSM attribute that stores the initial value
// for the HoldTimer.  The suggested default value for the HoldTime is
// 90 seconds.
const defaultHoldTime = 90 * time.Second

// During some portions of the state machine (see Section 8), the
// HoldTimer is set to a large value.  The suggested default for this
// large value is 4 minutes.
const defaultLargeHoldTimer = 4 * time.Minute

// The KeepaliveTime is a mandatory FSM attribute that stores the
// initial value for the KeepaliveTimer.  The suggested default value
// for the KeepaliveTime is 1/3 of the HoldTime.
const defaultKeepaliveTime = defaultHoldTime / 3

// The suggested default value for the MinASOriginationIntervalTimer is
// 15 seconds.
const minASOriginationIntervalTimer = 15 * time.Second

// The suggested default value for the
// MinRouteAdvertisementIntervalTimer on EBGP connections is 30 seconds.
const minRouteAdvertisementIntervalTimerEBGP = 30 * time.Second

// The suggested default value for the
// MinRouteAdvertisementIntervalTimer on IBGP connections is 5 seconds.
const minRouteAdvertisementIntervalTimerIBGP = 5 * time.Second
