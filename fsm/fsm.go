package fsm

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/transitorykris/kbgp/timer"
)

// 8.  BGP Finite State Machine (FSM)
//    The data structures and FSM described in this document are conceptual
//    and do not have to be implemented precisely as described here, as
//    long as the implementations support the described functionality and
//    they exhibit the same externally visible behavior.
type fsm struct {

	//    This section specifies the BGP operation in terms of a Finite State
	//    Machine (FSM).  The section falls into two parts:

	//       1) Description of Events for the State machine (Section 8.1)
	//       2) Description of the FSM (Section 8.2)

	//    Session attributes required (mandatory) for each connection are:

	state               int           // 1) State
	connectRetryCounter int           // 2) ConnectRetryCounter
	connectRetryTimer   *timer.Timer  // 3) ConnectRetryTimer
	connectRetryTime    time.Duration // 4) ConnectRetryTime
	holdTimer           *timer.Timer  // 5) HoldTimer
	// 6) HoldTime
	initialHoldTime time.Duration // initialHoldTime is the configured hold time
	holdTime        time.Duration // holdTime is the negotiated hold time
	keepaliveTimer  *timer.Timer  // 7) KeepaliveTimer
	keepaliveTime   time.Duration // 8) KeepaliveTime

	//    The state session attribute indicates the current state of the BGP
	//    FSM.  The ConnectRetryCounter indicates the number of times a BGP
	//    peer has tried to establish a peer session.

	//    The mandatory attributes related to timers are described in Section
	//    10.  Each timer has a "timer" and a "time" (the initial value).

	//    The optional Session attributes are listed below.  These optional
	//    attributes may be supported, either per connection or per local
	//    system:

	acceptConnectionsUnconfiguredPeers bool          // 1) AcceptConnectionsUnconfiguredPeers
	allowAutomaticStart                bool          // 2) AllowAutomaticStart
	allowAutomaticStop                 bool          // 3) AllowAutomaticStop
	collisionDetectEstablishedState    bool          // 4) CollisionDetectEstablishedState
	dampPeerOscillations               bool          // 5) DampPeerOscillations
	delayOpen                          bool          // 6) DelayOpen
	delayOpenTime                      time.Duration // 7) DelayOpenTime
	delayOpenTimer                     *timer.Timer  // 8) DelayOpenTimer
	idleHoldTime                       time.Duration // 9) IdleHoldTime
	idleHoldTimer                      *timer.Timer  // 10) IdleHoldTimer
	passiveTCPEstablishment            bool          // 11) PassiveTcpEstablishment
	sendNotificationwithoutOpen        bool          // 12) SendNOTIFICATIONwithoutOPEN
	trackTCPState                      bool          // 13) TrackTcpState

	//    The optional session attributes support different features of the BGP
	//    functionality that have implications for the BGP FSM state
	//    transitions.  Two groups of the attributes which relate to timers
	//    are:

	//       group 1: DelayOpen, DelayOpenTime, DelayOpenTimer
	//       group 2: DampPeerOscillations, IdleHoldTime, IdleHoldTimer

	//    The first parameter (DelayOpen, DampPeerOscillations) is an optional
	//    attribute that indicates that the Timer function is active.  The
	//    "Time" value specifies the initial value for the "Timer"
	//    (DelayOpenTime, IdleHoldTime).  The "Timer" specifies the actual
	//    timer.

	//    Please refer to Section 8.1.1 for an explanation of the interaction
	//    between these optional attributes and the events signaled to the
	//    state machine.  Section 8.2.1.3 also provides a short overview of the
	//    different types of optional attributes (flags or timers).

	peer *peer
}

type peer struct {
	// conn is the winning connection with this peer
	conn net.Conn

	// incomingConn and outgoingConn is a waiting room while managing
	// connection collisions
	incomingConn net.Conn
	outgoingConn net.Conn

	remoteAS uint16
	remoteIP net.IP

	adjRIBIn  *adjRIBIn
	adjRIBOut *adjRIBOut
}

func newPeer(remoteAS uint16, remoteIP net.IP) *peer {
	p := &peer{
		remoteAS: remoteAS,
		remoteIP: remoteIP,
	}
	return p
}

func (f *fsm) initialize() {
	f.peer.adjRIBIn = newAdjRIBIn()
	f.peer.adjRIBOut = newAdjRIBOut()
}

func (f *fsm) release() {
	f.peer.adjRIBIn = nil
	f.peer.adjRIBOut = nil
}

// dial attempts to form a TCP connection with the peer
func (f *fsm) dial() {
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", f.peer.remoteIP.String(), port))
	if err != nil {
		log.Fatal(err)
		// TODO: Figure out how to handle errors here
	}

	// Event 16: Tcp_CR_Acked

	// 		 Definition: Event indicating the local system's request to
	// 					 establish a TCP connection to the remote peer.

	// 					 The local system's TCP connection sent a TCP SYN,
	// 					 received a TCP SYN/ACK message, and sent a TCP ACK.

	// 		 Status:     Mandatory
	f.peer.outgoingConn = conn
	// Note: This event is probably incorrect. I believe it's for incoming
	// connection requests.
	f.sendEvent(tcpCRAcked)
}

// drops the TCP connection
func (f *fsm) drop() error {
	err := f.peer.conn.Close()
	return err
}

func (f *fsm) write(v interface{}) {
	// TODO: Serialize the interface's values as raw bytes
	// If the write fails, the TCP connection is likely closed
	// f.sendEvent(tcpConnectionFails)
}

// read perpetually processes messages and routes them to where they need to go
func (f *fsm) reader() {
	header, message := readMessage(f.peer.conn)
	switch header.messageType {
	case open:
		open := readOpen(message)
		if notif, ok := open.valid(f.peer.remoteAS, durationToUint16(f.holdTime)); !ok {
			// TODO: Send the notification
			log.Println("Sending notification", notif)
		}
	case update:
		// From section 9
		//    An UPDATE message may be received only in the Established state.
		//    Receiving an UPDATE message in any other state is an error.
		if f.state != established {
			// notif := newNotificationMessage()
			// TODO: Send the notification
			// log.Println("Sending notification", notif)
			break
		}
		update := readUpdate(message)
		if notif, ok := update.valid(); !ok {
			// TODO: Send the notification
			log.Println("Sending notification", notif)
		}
	case notification:
		notification := readKeepalive(message)
		if notif, ok := notification.valid(); !ok {
			// TODO: Send the notification
			log.Println("Sending notification", notif)
		}
	case keepalive:
		readKeepalive(message)
		keepalive := readKeepalive(message)
		if notif, ok := keepalive.valid(); !ok {
			// TODO: Send the notification
			log.Println("Sending notification", notif)
		}
	default:
		// NOTIFICATION - messageHeaderError, badMessageType
	}
}

func (f *fsm) open() {
	// TODO: Implement me
	//o := newOpenMessage()
	//f.write(o)
}

func (f *fsm) update() {
	// TODO: Implement me
	//u := newUpdateMessage()
	//f.write(u)
	// Note: Each time the local system sends a KEEPALIVE or UPDATE message, it
	//       restarts its KeepaliveTimer, unless the negotiated HoldTime value
	//       is zero.
	if f.holdTime != 0 {
		f.keepaliveTimer.Reset(defaultKeepaliveTime)
	}
}

func (f *fsm) notification(code int, subcode int, data []byte) {
	// TODO: Implement me
	n := newNotificationMessage(code, subcode, data)
	f.write(n)
}

func (f *fsm) keepalive() {
	// TODO: Implement me
	k := newKeepaliveMessage()
	f.write(k)
	// Note: Each time the local system sends a KEEPALIVE or UPDATE message, it
	//       restarts its KeepaliveTimer, unless the negotiated HoldTime value
	//       is zero.
	if f.holdTime != 0 {
		f.keepaliveTimer.Reset(defaultKeepaliveTime)
	}
}

func (f *fsm) sendEvent(event int) func() {
	return func() {
		log.Printf("Sending (%d) %s to state (%d) %s", event, eventName[event], f.state, stateName[f.state])
		switch f.state {
		case idle:
			f.idle(event)
		case connect:
			f.connect(event)
		case active:
			f.active(event)
		case openConfirm:
			f.openConfirm(event)
		case established:
			f.established(event)
		}
	}
}

// Note: This is a guess, RFC4271 does not appear to specify a value
const defaultDelayOpenTime = 1 * time.Second

// 8.1.  Events for the BGP FSM

// 8.1.1.  Optional Events Linked to Optional Session Attributes

//    The Inputs to the BGP FSM are events.  Events can either be mandatory
//    or optional.  Some optional events are linked to optional session
//    attributes.  Optional session attributes enable several groups of FSM
//    functionality.

//    The linkage between FSM functionality, events, and the optional
//    session attributes are described below.

//       Group 1: Automatic Administrative Events (Start/Stop)

//          Optional Session Attributes: AllowAutomaticStart,
//                                       AllowAutomaticStop,
//                                       DampPeerOscillations,
//                                       IdleHoldTime, IdleHoldTimer

//          Option 1:    AllowAutomaticStart

//          Description: A BGP peer connection can be started and stopped
//                       by administrative control.  This administrative
//                       control can either be manual, based on operator
//                       intervention, or under the control of logic that
//                       is specific to a BGP implementation.  The term
//                       "automatic" refers to a start being issued to the
//                       BGP peer connection FSM when such logic determines
//                       that the BGP peer connection should be restarted.

//                       The AllowAutomaticStart attribute specifies that
//                       this BGP connection supports automatic starting of
//                       the BGP connection.

//                       If the BGP implementation supports
//                       AllowAutomaticStart, the peer may be repeatedly
//                       restarted.  Three other options control the rate
//                       at which the automatic restart occurs:
//                       DampPeerOscillations, IdleHoldTime, and the
//                       IdleHoldTimer.

//                       The DampPeerOscillations option specifies that the
//                       implementation engages additional logic to damp
//                       the oscillations of BGP peers in the face of
//                       sequences of automatic start and automatic stop.
//                       IdleHoldTime specifies the length of time the BGP
//                       peer is held in the Idle state prior to allowing
//                       the next automatic restart.  The IdleHoldTimer is
//                       the timer that holds the peer in Idle state.

//                       An example of DampPeerOscillations logic is an
//                       increase of the IdleHoldTime value if a BGP peer
//                       oscillates connectivity (connected/disconnected)
//                       repeatedly within a time period.  To engage this
//                       logic, a peer could connect and disconnect 10
//                       times within 5 minutes.  The IdleHoldTime value
//                       would be reset from 0 to 120 seconds.

//          Values:      TRUE or FALSE

//          Option 2:    AllowAutomaticStop

//          Description: This BGP peer session optional attribute indicates
//                       that the BGP connection allows "automatic"
//                       stopping of the BGP connection.  An "automatic"
//                       stop is defined as a stop under the control of
//                       implementation-specific logic.  The
//                       implementation-specific logic is outside the scope
//                       of this specification.

//          Values:      TRUE or FALSE

//          Option 3:    DampPeerOscillations

//          Description: The DampPeerOscillations optional session
//                       attribute indicates that the BGP connection is
//                       using logic that damps BGP peer oscillations in
//                       the Idle State.

//          Value:       TRUE or FALSE

//          Option 4:    IdleHoldTime

//          Description: The IdleHoldTime is the value that is set in the
//                       IdleHoldTimer.

//          Values:      Time in seconds

//          Option 5:    IdleHoldTimer

//          Description: The IdleHoldTimer aids in controlling BGP peer
//                       oscillation.  The IdleHoldTimer is used to keep
//                       the BGP peer in Idle for a particular duration.
//                       The IdleHoldTimer_Expires event is described in
//                       Section 8.1.3.

//          Values:      Time in seconds

//       Group 2: Unconfigured Peers

//          Optional Session Attributes: AcceptConnectionsUnconfiguredPeers

//          Option 1:    AcceptConnectionsUnconfiguredPeers

//          Description: The BGP FSM optionally allows the acceptance of
//                       BGP peer connections from neighbors that are not
//                       pre-configured.  The
//                       "AcceptConnectionsUnconfiguredPeers" optional
//                       session attribute allows the FSM to support the
//                       state transitions that allow the implementation to
//                       accept or reject these unconfigured peers.

//                       The AcceptConnectionsUnconfiguredPeers has
//                       security implications.  Please refer to the BGP
//                       Vulnerabilities document [RFC4272] for details.

//          Value:       True or False

//       Group 3: TCP processing

//          Optional Session Attributes: PassiveTcpEstablishment,
//                                       TrackTcpState

//          Option 1:    PassiveTcpEstablishment

//          Description: This option indicates that the BGP FSM will
//                       passively wait for the remote BGP peer to
//                       establish the BGP TCP connection.

//          value:       TRUE or FALSE

//          Option 2:    TrackTcpState

//          Description: The BGP FSM normally tracks the end result of a
//                       TCP connection attempt rather than individual TCP
//                       messages.  Optionally, the BGP FSM can support
//                       additional interaction with the TCP connection
//                       negotiation.  The interaction with the TCP events
//                       may increase the amount of logging the BGP peer
//                       connection requires and the number of BGP FSM
//                       changes.

//          Value:       TRUE or FALSE

//       Group 4:  BGP Message Processing

//          Optional Session Attributes: DelayOpen, DelayOpenTime,
//                                       DelayOpenTimer,
//                                       SendNOTIFICATIONwithoutOPEN,
//                                       CollisionDetectEstablishedState

//          Option 1:     DelayOpen

//          Description: The DelayOpen optional session attribute allows
//                       implementations to be configured to delay sending
//                       an OPEN message for a specific time period
//                       (DelayOpenTime).  The delay allows the remote BGP
//                       Peer time to send the first OPEN message.

//          Value:       TRUE or FALSE

//          Option 2:    DelayOpenTime

//          Description: The DelayOpenTime is the initial value set in the
//                       DelayOpenTimer.

//          Value:       Time in seconds

//          Option 3:    DelayOpenTimer

//          Description: The DelayOpenTimer optional session attribute is
//                       used to delay the sending of an OPEN message on a

//                       connection.  The DelayOpenTimer_Expires event
//                       (Event 12) is described in Section 8.1.3.

//          Value:       Time in seconds

//          Option 4:    SendNOTIFICATIONwithoutOPEN

//          Description: The SendNOTIFICATIONwithoutOPEN allows a peer to
//                       send a NOTIFICATION without first sending an OPEN
//                       message.  Without this optional session attribute,
//                       the BGP connection assumes that an OPEN message
//                       must be sent by a peer prior to the peer sending a
//                       NOTIFICATION message.

//          Value:       True or False

//          Option 5:    CollisionDetectEstablishedState

//          Description: Normally, a Detect Collision (see Section 6.8)
//                       will be ignored in the Established state.  This
//                       optional session attribute indicates that this BGP
//                       connection processes collisions in the Established
//                       state.

//          Value:       True or False

//       Note: The optional session attributes clarify the BGP FSM
//             description for existing features of BGP implementations.
//             The optional session attributes may be pre-defined for an
//             implementation and not readable via management interfaces
//             for existing correct implementations.  As newer BGP MIBs
//             (version 2 and beyond) are supported, these fields will be
//             accessible via a management interface.

const (
	_ = iota
	// 8.1.2.  Administrative Events

	//    An administrative event is an event in which the operator interface
	//    and BGP Policy engine signal the BGP-finite state machine to start or
	//    stop the BGP state machine.  The basic start and stop indications are
	//    augmented by optional connection attributes that signal a certain
	//    type of start or stop mechanism to the BGP FSM.  An example of this
	//    combination is Event 5, AutomaticStart_with_PassiveTcpEstablishment.
	//    With this event, the BGP implementation signals to the BGP FSM that
	//    the implementation is using an Automatic Start with the option to use
	//    a Passive TCP Establishment.  The Passive TCP establishment signals
	//    that this BGP FSM will wait for the remote side to start the TCP
	//    establishment.

	//    Note that only Event 1 (ManualStart) and Event 2 (ManualStop) are
	//    mandatory administrative events.  All other administrative events are
	//    optional (Events 3-8).  Each event below has a name, definition,
	//    status (mandatory or optional), and the optional session attributes
	//    that SHOULD be set at each stage.  When generating Event 1 through
	//    Event 8 for the BGP FSM, the conditions specified in the "Optional
	//    Attribute Status" section are verified.  If any of these conditions
	//    are not satisfied, then the local system should log an FSM error.

	//    The settings of optional session attributes may be implicit in some
	//    implementations, and therefore may not be set explicitly by an
	//    external operator action.  Section 8.2.1.5 describes these implicit
	//    settings of the optional session attributes.  The administrative
	//    states described below may also be implicit in some implementations
	//    and not directly configurable by an external operator.

	//       Event 1: ManualStart
	//          Definition: Local system administrator manually starts the peer
	//                      connection.
	//          Status:     Mandatory
	//          Optional
	//          Attribute
	//          Status:     The PassiveTcpEstablishment attribute SHOULD be set
	//                      to FALSE.
	manualStart

	//       Event 2: ManualStop
	//          Definition: Local system administrator manually stops the peer
	//                      connection.
	//          Status:     Mandatory
	//          Optional
	//          Attribute
	//          Status:     No interaction with any optional attributes.
	manualStop

	//       Event 3: AutomaticStart
	//          Definition: Local system automatically starts the BGP
	//                      connection.
	//          Status:     Optional, depending on local system
	//          Optional
	//          Attribute
	//          Status:     1) The AllowAutomaticStart attribute SHOULD be set
	//                         to TRUE if this event occurs.
	//                      2) If the PassiveTcpEstablishment optional session
	//                         attribute is supported, it SHOULD be set to
	//                         FALSE.
	//                      3) If the DampPeerOscillations is supported, it
	//                         SHOULD be set to FALSE when this event occurs.
	automaticStart

	//       Event 4: ManualStart_with_PassiveTcpEstablishment
	//          Definition: Local system administrator manually starts the peer
	//                      connection, but has PassiveTcpEstablishment
	//                      enabled.  The PassiveTcpEstablishment optional
	//                      attribute indicates that the peer will listen prior
	//                      to establishing the connection.
	//          Status:     Optional, depending on local system
	//          Optional
	//          Attribute
	//          Status:     1) The PassiveTcpEstablishment attribute SHOULD be
	//                         set to TRUE if this event occurs.
	//                      2) The DampPeerOscillations attribute SHOULD be set
	//                         to FALSE when this event occurs.
	manualStartWithPassiveTCPEstablishment

	//       Event 5: AutomaticStart_with_PassiveTcpEstablishment
	//          Definition: Local system automatically starts the BGP
	//                      connection with the PassiveTcpEstablishment
	//                      enabled.  The PassiveTcpEstablishment optional
	//                      attribute indicates that the peer will listen prior
	//                      to establishing a connection.
	//          Status:     Optional, depending on local system
	//          Optional
	//          Attribute
	//          Status:     1) The AllowAutomaticStart attribute SHOULD be set
	//                         to TRUE.
	//                      2) The PassiveTcpEstablishment attribute SHOULD be
	//                         set to TRUE.
	//                      3) If the DampPeerOscillations attribute is
	//                         supported, the DampPeerOscillations SHOULD be
	//                         set to FALSE.
	automaticStartWithPassiveTCPEstablishment

	//       Event 6: AutomaticStart_with_DampPeerOscillations
	//          Definition: Local system automatically starts the BGP peer
	//                      connection with peer oscillation damping enabled.
	//                      The exact method of damping persistent peer
	//                      oscillations is determined by the implementation
	//                      and is outside the scope of this document.
	//          Status:     Optional, depending on local system.
	//          Optional
	//          Attribute
	//          Status:     1) The AllowAutomaticStart attribute SHOULD be set
	//                         to TRUE.
	//                      2) The DampPeerOscillations attribute SHOULD be set
	//                         to TRUE.
	//                      3) The PassiveTcpEstablishment attribute SHOULD be
	//                         set to FALSE.
	automaticStartWithDampPeerOscillations

	//       Event 7: AutomaticStart_with_DampPeerOscillations_and_
	//       PassiveTcpEstablishment
	//          Definition: Local system automatically starts the BGP peer
	//                      connection with peer oscillation damping enabled
	//                      and PassiveTcpEstablishment enabled.  The exact
	//                      method of damping persistent peer oscillations is
	//                      determined by the implementation and is outside the
	//                      scope of this document.
	//          Status:     Optional, depending on local system
	//          Optional
	//          Attributes
	//          Status:     1) The AllowAutomaticStart attribute SHOULD be set
	//                         to TRUE.
	//                      2) The DampPeerOscillations attribute SHOULD be set
	//                         to TRUE.
	//                      3) The PassiveTcpEstablishment attribute SHOULD be
	//                         set to TRUE.
	automaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment

	//       Event 8: AutomaticStop
	//          Definition: Local system automatically stops the BGP
	//                      connection.
	//                      An example of an automatic stop event is exceeding
	//                      the number of prefixes for a given peer and the
	//                      local system automatically disconnecting the peer.
	//          Status:     Optional, depending on local system
	//          Optional
	//          Attribute
	//          Status:     1) The AllowAutomaticStop attribute SHOULD be TRUE.
	automaticStop

	// 8.1.3.  Timer Events

	//       Event 9: ConnectRetryTimer_Expires
	//          Definition: An event generated when the ConnectRetryTimer
	//                      expires.
	//          Status:     Mandatory
	connectRetryTimerExpires

	//       Event 10: HoldTimer_Expires
	//          Definition: An event generated when the HoldTimer expires.
	//          Status:     Mandatory
	holdTimerExpires

	//       Event 11: KeepaliveTimer_Expires
	//          Definition: An event generated when the KeepaliveTimer expires.
	//          Status:     Mandatory
	keepaliveTimerExpires

	//       Event 12: DelayOpenTimer_Expires
	//          Definition: An event generated when the DelayOpenTimer expires.
	//          Status:     Optional
	//          Optional
	//          Attribute
	//          Status:     If this event occurs,
	//                      1) DelayOpen attribute SHOULD be set to TRUE,
	//                      2) DelayOpenTime attribute SHOULD be supported,
	//                      3) DelayOpenTimer SHOULD be supported.
	delayOpenTimerExpires

	//       Event 13: IdleHoldTimer_Expires
	//          Definition: An event generated when the IdleHoldTimer expires,
	//                      indicating that the BGP connection has completed
	//                      waiting for the back-off period to prevent BGP peer
	//                      oscillation.
	//                      The IdleHoldTimer is only used when the persistent
	//                      peer oscillation damping function is enabled by
	//                      setting the DampPeerOscillations optional attribute
	//                      to TRUE.
	//                      Implementations not implementing the persistent
	//                      peer oscillation damping function may not have the
	//                      IdleHoldTimer.
	//          Status:     Optional
	//          Optional
	//          Attribute
	//          Status:     If this event occurs:
	//                      1) DampPeerOscillations attribute SHOULD be set to
	//                         TRUE.
	//                      2) IdleHoldTimer SHOULD have just expired.
	idleHoldTimerExpires

	// 8.1.4.  TCP Connection-Based Events

	//       Event 14: TcpConnection_Valid
	//          Definition: Event indicating the local system reception of a
	//                      TCP connection request with a valid source IP
	//                      address, TCP port, destination IP address, and TCP
	//                      Port.  The definition of invalid source and invalid
	//                      destination IP address is determined by the
	//                      implementation.
	//                      BGP's destination port SHOULD be port 179, as
	//                      defined by IANA.
	//                      TCP connection request is denoted by the local
	//                      system receiving a TCP SYN.
	//          Status:     Optional
	//          Optional
	//          Attribute
	//          Status:     1) The TrackTcpState attribute SHOULD be set to
	//                         TRUE if this event occurs.
	tcpConnectionValid

	//       Event 15: Tcp_CR_Invalid
	//          Definition: Event indicating the local system reception of a
	//                      TCP connection request with either an invalid
	//                      source address or port number, or an invalid
	//                      destination address or port number.
	//                      BGP destination port number SHOULD be 179, as
	//                      defined by IANA.
	//                      A TCP connection request occurs when the local
	//                      system receives a TCP SYN.
	//          Status:     Optional
	//          Optional
	//          Attribute
	//          Status:     1) The TrackTcpState attribute should be set to
	//                         TRUE if this event occurs.
	tcpCRInvalid

	//       Event 16: Tcp_CR_Acked
	//          Definition: Event indicating the local system's request to
	//                      establish a TCP connection to the remote peer.
	//                      The local system's TCP connection sent a TCP SYN,
	//                      received a TCP SYN/ACK message, and sent a TCP ACK.
	//          Status:     Mandatory
	tcpCRAcked

	//       Event 17: TcpConnectionConfirmed
	//          Definition: Event indicating that the local system has received
	//                      a confirmation that the TCP connection has been
	//                      established by the remote site.
	//                      The remote peer's TCP engine sent a TCP SYN.  The
	//                      local peer's TCP engine sent a SYN, ACK message and
	//                      now has received a final ACK.
	//          Status:     Mandatory
	tcpConnectionConfirmed

	//       Event 18: TcpConnectionFails
	//          Definition: Event indicating that the local system has received
	//                      a TCP connection failure notice.
	//                      The remote BGP peer's TCP machine could have sent a
	//                      FIN.  The local peer would respond with a FIN-ACK.
	//                      Another possibility is that the local peer
	//                      indicated a timeout in the TCP connection and
	//                      downed the connection.
	//          Status:     Mandatory
	tcpConnectionFails

	// 8.1.5.  BGP Message-Based Events

	//       Event 19: BGPOpen
	//          Definition: An event is generated when a valid OPEN message has
	//                      been received.
	//          Status:     Mandatory
	//          Optional
	//          Attribute
	//          Status:     1) The DelayOpen optional attribute SHOULD be set
	//                         to FALSE.
	//                      2) The DelayOpenTimer SHOULD not be running.
	bgpOpen

	//       Event 20: BGPOpen with DelayOpenTimer running
	//          Definition: An event is generated when a valid OPEN message has
	//                      been received for a peer that has a successfully
	//                      established transport connection and is currently
	//                      delaying the sending of a BGP open message.
	//          Status:     Optional
	//          Optional
	//          Attribute
	//          Status:     1) The DelayOpen attribute SHOULD be set to TRUE.
	//                      2) The DelayOpenTimer SHOULD be running.
	bgpOpenWithDelayOpenTimerRunning

	//       Event 21: BGPHeaderErr
	//          Definition: An event is generated when a received BGP message
	//                      header is not valid.
	//          Status:     Mandatory
	bgpHeaderErr

	//       Event 22: BGPOpenMsgErr
	//          Definition: An event is generated when an OPEN message has been
	//                      received with errors.
	//          Status:     Mandatory
	bgpOpenMsgErr

	//       Event 23: OpenCollisionDump
	//          Definition: An event generated administratively when a
	//                      connection collision has been detected while
	//                      processing an incoming OPEN message and this
	//                      connection has been selected to be disconnected.
	//                      See Section 6.8 for more information on collision
	//                      detection.
	//                      Event 23 is an administrative action generated by
	//                      implementation logic that determines whether this
	//                      connection needs to be dropped per the rules in
	//                      Section 6.8.  This event may occur if the FSM is
	//                      implemented as two linked state machines.
	//          Status:     Optional
	//          Optional
	//          Attribute
	//          Status:     If the state machine is to process this event in
	//                      the Established state,
	//                      1) CollisionDetectEstablishedState optional
	//                         attribute SHOULD be set to TRUE.
	//                      Please note: The OpenCollisionDump event can occur
	//                      in Idle, Connect, Active, OpenSent, and OpenConfirm
	//                      without any optional attributes being set.
	openCollisionDump

	//       Event 24: NotifMsgVerErr
	//          Definition: An event is generated when a NOTIFICATION message
	//                      with "version error" is received.
	//          Status:     Mandatory
	notifMsgVerErr

	//       Event 25: NotifMsg
	//          Definition: An event is generated when a NOTIFICATION message
	//                      is received and the error code is anything but
	//                      "version error".
	//          Status:     Mandatory
	notifMsg

	//       Event 26: KeepAliveMsg
	//          Definition: An event is generated when a KEEPALIVE message is
	//                      received.
	//          Status:     Mandatory
	keepAliveMsg

	//       Event 27: UpdateMsg
	//          Definition: An event is generated when a valid UPDATE message
	//                      is received.
	//          Status:     Mandatory
	updateMsg

	//       Event 28: UpdateMsgErr
	//          Definition: An event is generated when an invalid UPDATE
	//                      message is received.
	//          Status:     Mandatory
	updateMsgErr
)

// Reverse value lookup for event names
var eventName = map[int]string{
	1:  "ManualStart",
	2:  "ManualStop",
	3:  "AutomaticStart",
	4:  "ManualStart_with_PassiveTcpEstablishment",
	5:  "AutomaticStart_with_PassiveTcpEstablishment",
	6:  "AutomaticStart_with_DampPeerOscillations",
	7:  "AutomaticStart_with_DampPeerOscillations_and_PassiveTcpEstablishment",
	8:  "AutomaticStop",
	9:  "ConnectRetryTimer_Expires",
	10: "HoldTimer_Expires",
	11: "KeepaliveTimer_Expires",
	12: "KeepaliveTimer_Expires",
	13: "DelayOpenTimer_Expires",
	14: "IdleHoldTimer_Expires",
	15: "TcpConnection_Valid",
	16: "Tcp_CR_Acked",
	17: "TcpConnectionConfirmed",
	18: "TcpConnectionFails",
	19: "BGPOpen",
	20: "BGPOpen with DelayOpenTimer running",
	21: "BGPHeaderErr",
	22: "BGPOpenMsgErr",
	23: "OpenCollisionDump",
	24: "NotifMsgVerErr",
	25: "NotifMsg",
	26: "KeepAliveMsg",
	27: "UpdateMsg",
	28: "UpdateMsgErr",
}

const (
	idle = iota
	connect
	active
	openSent
	openConfirm
	established
)

// Reverse value lookup for state names
var stateName = map[int]string{
	0: "Idle",
	1: "Connect",
	2: "Active",
	3: "OpenConfirm",
	4: "Established",
}

//    Idle state:
//       Initially, the BGP peer FSM is in the Idle state.  Hereafter, the
//       BGP peer FSM will be shortened to BGP FSM.
func (f *fsm) idle(event int) {
	switch event {
	case manualStart:
		//       In this state, BGP FSM refuses all incoming BGP connections for
		//       this peer.  No resources are allocated to the peer.  In response
		//       to a ManualStart event (Event 1) or an AutomaticStart event (Event
		//       3), the local system:
		//         - initializes all BGP resources for the peer connection,
		f.initialize()
		//         - sets ConnectRetryCounter to zero,
		f.connectRetryCounter = 0
		//         - starts the ConnectRetryTimer with the initial value,
		f.connectRetryTimer = timer.New(defaultConnectRetryTime, f.sendEvent(connectRetryTimerExpires))
		//         - initiates a TCP connection to the other BGP peer,
		go f.dial()
		//         - listens for a connection that may be initiated by the remote
		//           BGP peer, and
		// NOTE: Listening is already happening, but may need a better mechanism
		// e.g. We add a this peer to the listener, remove it from the listener when
		// we shut down. Right now the listener scans through all FSMs.
		//         - changes its state to Connect.
		f.transition(connect)
	case automaticStart:
		//       In this state, BGP FSM refuses all incoming BGP connections for
		//       this peer.  No resources are allocated to the peer.  In response
		//       to a ManualStart event (Event 1) or an AutomaticStart event (Event
		//       3), the local system:
		//         - initializes all BGP resources for the peer connection,
		f.initialize()
		//         - sets ConnectRetryCounter to zero,
		f.connectRetryCounter = 0
		//         - starts the ConnectRetryTimer with the initial value,
		f.connectRetryTimer = timer.New(defaultConnectRetryTime, f.sendEvent(connectRetryTimerExpires))
		//         - initiates a TCP connection to the other BGP peer,
		go f.dial()
		//         - listens for a connection that may be initiated by the remote
		//           BGP peer, and
		// NOTE: Listening is already happening, but may need a better mechanism
		// e.g. We add a this peer to the listener, remove it from the listener when
		// we shut down. Right now the listener scans through all FSMs.
		//         - changes its state to Connect.
		f.transition(connect)
	case manualStop:
		//       The ManualStop event (Event 2) and AutomaticStop (Event 8) event
		//       are ignored in the Idle state.
	case automaticStop:
		//       The ManualStop event (Event 2) and AutomaticStop (Event 8) event
		//       are ignored in the Idle state.
	case manualStartWithPassiveTCPEstablishment:
		//       In response to a ManualStart_with_PassiveTcpEstablishment event
		//       (Event 4) or AutomaticStart_with_PassiveTcpEstablishment event
		//       (Event 5), the local system:
		//         - initializes all BGP resources,
		f.initialize()
		//         - sets the ConnectRetryCounter to zero,
		f.connectRetryCounter = 0
		//         - starts the ConnectRetryTimer with the initial value,
		f.connectRetryTimer = timer.New(defaultConnectRetryTime, f.sendEvent(connectRetryTimerExpires))
		//         - listens for a connection that may be initiated by the remote
		//           peer, and
		// NOTE: Listening is already happening, but may need a better mechanism
		// e.g. We add a this peer to the listener, remove it from the listener when
		// we shut down. Right now the listener scans through all FSMs.
		//         - changes its state to Active.
		f.transition(active)
		//       The exact value of the ConnectRetryTimer is a local matter, but it
		//       SHOULD be sufficiently large to allow TCP initialization.
	case automaticStartWithPassiveTCPEstablishment:
		//       In response to a ManualStart_with_PassiveTcpEstablishment event
		//       (Event 4) or AutomaticStart_with_PassiveTcpEstablishment event
		//       (Event 5), the local system:
		//         - initializes all BGP resources,
		f.initialize()
		//         - sets the ConnectRetryCounter to zero,
		f.connectRetryCounter = 0
		//         - starts the ConnectRetryTimer with the initial value,
		f.connectRetryTimer = timer.New(defaultConnectRetryTime, f.sendEvent(connectRetryTimerExpires))
		//         - listens for a connection that may be initiated by the remote
		//           peer, and
		// NOTE: Listening is already happening, but may need a better mechanism
		// e.g. We add a this peer to the listener, remove it from the listener when
		// we shut down. Right now the listener scans through all FSMs.
		//         - changes its state to Active.
		f.transition(active)
		//       The exact value of the ConnectRetryTimer is a local matter, but it
		//       SHOULD be sufficiently large to allow TCP initialization.
	case automaticStartWithDampPeerOscillations:
		//       If the DampPeerOscillations attribute is set to TRUE, the
		//       following three additional events may occur within the Idle state:
		//         - AutomaticStart_with_DampPeerOscillations (Event 6),
		//         - AutomaticStart_with_DampPeerOscillations_and_
		//           PassiveTcpEstablishment (Event 7),
		//         - IdleHoldTimer_Expires (Event 13).

		//       Upon receiving these 3 events, the local system will use these
		//       events to prevent peer oscillations.  The method of preventing
		//       persistent peer oscillation is outside the scope of this document.
	case automaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
		//       If the DampPeerOscillations attribute is set to TRUE, the
		//       following three additional events may occur within the Idle state:
		//         - AutomaticStart_with_DampPeerOscillations (Event 6),
		//         - AutomaticStart_with_DampPeerOscillations_and_
		//           PassiveTcpEstablishment (Event 7),
		//         - IdleHoldTimer_Expires (Event 13).

		//       Upon receiving these 3 events, the local system will use these
		//       events to prevent peer oscillations.  The method of preventing
		//       persistent peer oscillation is outside the scope of this document.
	case idleHoldTimerExpires:
		//       If the DampPeerOscillations attribute is set to TRUE, the
		//       following three additional events may occur within the Idle state:
		//         - AutomaticStart_with_DampPeerOscillations (Event 6),
		//         - AutomaticStart_with_DampPeerOscillations_and_
		//           PassiveTcpEstablishment (Event 7),
		//         - IdleHoldTimer_Expires (Event 13).

		//       Upon receiving these 3 events, the local system will use these
		//       events to prevent peer oscillations.  The method of preventing
		//       persistent peer oscillation is outside the scope of this document.
	default:
		//       Any other event (Events 9-12, 15-28) received in the Idle state
		//       does not cause change in the state of the local system.
	}
}

//    Connect State:
//       In this state, BGP FSM is waiting for the TCP connection to be
//       completed.
func (f *fsm) connect(event int) {
	switch event {
	//       The start events (Events 1, 3-7) are ignored in the Connect state.
	case manualStart:
	case automaticStart:
	case manualStartWithPassiveTCPEstablishment:
	case automaticStartWithPassiveTCPEstablishment:
	case automaticStartWithDampPeerOscillations:
	case automaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case manualStop:
		//       In response to a ManualStop event (Event 2), the local system:
		//         - drops the TCP connection,
		f.drop()
		//         - releases all BGP resources,
		f.release()
		//         - sets ConnectRetryCounter to zero,
		f.connectRetryCounter = 0
		//         - stops the ConnectRetryTimer and sets ConnectRetryTimer to
		//           zero, and
		f.connectRetryTimer.Stop()
		//         - changes its state to Idle.
		f.transition(idle)
	case connectRetryTimerExpires:
		//       In response to the ConnectRetryTimer_Expires event (Event 9), the
		//       local system:
		//         - drops the TCP connection,
		f.drop()
		//         - restarts the ConnectRetryTimer,
		f.connectRetryTimer.Reset(f.connectRetryTime)
		//         - stops the DelayOpenTimer and resets the timer to zero,
		f.delayOpenTimer.Stop()
		//         - initiates a TCP connection to the other BGP peer,
		f.dial()
		//         - continues to listen for a connection that may be initiated by
		//           the remote BGP peer, and
		//         - stays in the Connect state.
	case delayOpenTimerExpires:
		//       If the DelayOpenTimer_Expires event (Event 12) occurs in the
		//       Connect state, the local system:
		//         - sends an OPEN message to its peer,
		f.open()
		//         - sets the HoldTimer to a large value, and
		f.holdTimer.Reset(largeHoldTimer)
		//         - changes its state to OpenSent.
		f.transition(openSent)
	case tcpConnectionValid:
		//       If the BGP FSM receives a TcpConnection_Valid event (Event 14),
		//       the TCP connection is processed, and the connection remains in the
		//       Connect state.
	case tcpCRInvalid:
		//       If the BGP FSM receives a Tcp_CR_Invalid event (Event 15), the
		//       local system rejects the TCP connection, and the connection
		//       remains in the Connect state.
	case tcpCRAcked:
		//       If the TCP connection succeeds (Event 16 or Event 17), the local
		//       system checks the DelayOpen attribute prior to processing.  If the
		//       DelayOpen attribute is set to TRUE, the local system:
		if f.delayOpen {
			//         - stops the ConnectRetryTimer (if running) and sets the
			//           ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//         - sets the DelayOpenTimer to the initial value, and
			f.delayOpenTimer.Reset(f.initialHoldTime)
			//         - stays in the Connect state.
		} else {
			//       If the DelayOpen attribute is set to FALSE, the local system:
			//         - stops the ConnectRetryTimer (if running) and sets the
			//           ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//         - completes BGP initialization
			//         - sends an OPEN message to its peer,
			f.open()
			//         - sets the HoldTimer to a large value, and
			f.holdTimer.Reset(largeHoldTimer)
			//         - changes its state to OpenSent.
			f.transition(openSent)
			return
		}

		//       A HoldTimer value of 4 minutes is suggested.

		//       If the TCP connection fails (Event 18), the local system checks
		//       the DelayOpenTimer.  If the DelayOpenTimer is running, the local
		//       system:
		if f.delayOpenTimer.Running() {
			//         - restarts the ConnectRetryTimer with the initial value,
			f.connectRetryTimer.Reset(f.connectRetryTime)
			//         - stops the DelayOpenTimer and resets its value to zero,
			f.delayOpenTimer.Stop()
			//         - continues to listen for a connection that may be initiated by
			//           the remote BGP peer, and
			//         - changes its state to Active.
			f.transition(active)
			return
		}
		//       If the DelayOpenTimer is not running, the local system:
		//         - stops the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - drops the TCP connection,
		f.drop()
		//         - releases all BGP resources, and
		f.release()
		//         - changes its state to Idle.
		f.transition(idle)
	case tcpConnectionConfirmed:
		//       If the TCP connection succeeds (Event 16 or Event 17), the local
		//       system checks the DelayOpen attribute prior to processing.  If the
		//       DelayOpen attribute is set to TRUE, the local system:
		if f.delayOpen {
			//         - stops the ConnectRetryTimer (if running) and sets the
			//           ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//         - sets the DelayOpenTimer to the initial value, and
			f.delayOpenTimer.Reset(f.delayOpenTime)
			//         - stays in the Connect state.
		} else {
			//       If the DelayOpen attribute is set to FALSE, the local system:
			//         - stops the ConnectRetryTimer (if running) and sets the
			//           ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//         - completes BGP initialization
			//         - sends an OPEN message to its peer,
			f.open()
			//         - sets the HoldTimer to a large value, and
			f.holdTimer.Reset(largeHoldTimer)
			//         - changes its state to OpenSent.
			f.transition(openSent)
			return
		}

		//       A HoldTimer value of 4 minutes is suggested.

		//       If the TCP connection fails (Event 18), the local system checks
		//       the DelayOpenTimer.  If the DelayOpenTimer is running, the local
		//       system:
		if f.delayOpenTimer.Running() {
			//         - restarts the ConnectRetryTimer with the initial value,
			f.connectRetryTimer.Reset(f.connectRetryTime)
			//         - stops the DelayOpenTimer and resets its value to zero,
			f.delayOpenTimer.Stop()
			//         - continues to listen for a connection that may be initiated by
			//           the remote BGP peer, and
			//         - changes its state to Active.
			f.transition(active)
			return
		}
		//       If the DelayOpenTimer is not running, the local system:
		//         - stops the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - drops the TCP connection,
		f.drop()
		//         - releases all BGP resources, and
		f.release()
		//         - changes its state to Idle.
		f.transition(idle)
	case tcpConnectionFails:
		//       If the TCP connection fails (Event 18), the local system checks
		//       the DelayOpenTimer.  If the DelayOpenTimer is running, the local
		//       system:
		if f.delayOpenTimer.Running() {
			//         - restarts the ConnectRetryTimer with the initial value,
			f.connectRetryTimer.Reset(f.connectRetryTime)
			//         - stops the DelayOpenTimer and resets its value to zero,
			f.delayOpenTimer.Stop()
			//         - continues to listen for a connection that may be initiated by
			//           the remote BGP peer, and
			//         - changes its state to Active.
			f.transition(active)
			return
		}
		//       If the DelayOpenTimer is not running, the local system:
		//         - stops the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - drops the TCP connection,
		f.drop()
		//         - releases all BGP resources, and
		f.release()
		//         - changes its state to Idle.
		f.transition(idle)
	case bgpOpenWithDelayOpenTimerRunning:
		//       If an OPEN message is received while the DelayOpenTimer is running
		//       (Event 20), the local system:
		//         - stops the ConnectRetryTimer (if running) and sets the
		//           ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - completes the BGP initialization,
		//         - stops and clears the DelayOpenTimer (sets the value to zero),
		f.delayOpenTimer.Stop()
		//         - sends an OPEN message,
		f.open()
		//         - sends a KEEPALIVE message,
		f.keepalive()
		//         - if the HoldTimer initial value is non-zero,
		if f.initialHoldTime != 0 {
			//             - starts the KeepaliveTimer with the initial value and
			f.keepaliveTimer.Reset(f.keepaliveTime)
			//             - resets the HoldTimer to the negotiated value,
			f.holdTimer.Reset(f.holdTime)
		} else {
			//           else, if the HoldTimer initial value is zero,
			//             - resets the KeepaliveTimer and
			f.keepaliveTimer.Reset(f.keepaliveTime)
			//             - resets the HoldTimer value to zero,
			// Note: This seems redundant?
			f.holdTimer.Stop()
		}
		//         - and changes its state to OpenConfirm.
		f.transition(openConfirm)
		//       If the value of the autonomous system field is the same as the
		//       local Autonomous System number, set the connection status to an
		//       internal connection; otherwise it will be "external".
	case bgpHeaderErr:
		//       If BGP message header checking (Event 21) or OPEN message checking
		//       detects an error (Event 22) (see Section 6.2), the local system:
		//         - (optionally) If the SendNOTIFICATIONwithoutOPEN attribute is
		//           set to TRUE, then the local system first sends a NOTIFICATION
		//           message with the appropriate error code, and then
		if f.sendNotificationwithoutOpen {
			f.notification(messageHeaderError, noErrorSubcode, nil)
		}
		//         - stops the ConnectRetryTimer (if running) and sets the
		//           ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case bgpOpenMsgErr:
		//       If BGP message header checking (Event 21) or OPEN message checking
		//       detects an error (Event 22) (see Section 6.2), the local system:
		//         - (optionally) If the SendNOTIFICATIONwithoutOPEN attribute is
		//           set to TRUE, then the local system first sends a NOTIFICATION
		//           message with the appropriate error code, and then
		if f.sendNotificationwithoutOpen {
			f.notification(openMessageError, noErrorSubcode, nil)
		}
		//         - stops the ConnectRetryTimer (if running) and sets the
		//           ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case notifMsgVerErr:
		//       If a NOTIFICATION message is received with a version error (Event
		//       24), the local system checks the DelayOpenTimer.  If the
		//       DelayOpenTimer is running, the local system:
		if f.delayOpenTimer.Running() {
			//         - stops the ConnectRetryTimer (if running) and sets the
			//           ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//         - stops and resets the DelayOpenTimer (sets to zero),
			f.delayOpenTimer.Stop()
			//         - releases all BGP resources,
			f.release()
			//         - drops the TCP connection, and
			f.drop()
			//         - changes its state to Idle.
			f.transition(idle)
			return
		}
		//       If the DelayOpenTimer is not running, the local system:
		//         - stops the ConnectRetryTimer and sets the ConnectRetryTimer to
		//           zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - performs peer oscillation damping if the DampPeerOscillations
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//           attribute is set to True, and
		//         - changes its state to Idle.
		f.transition(idle)
	default:
		//       In response to any other events (Events 8, 10-11, 13, 19, 23,
		//       25-28), the local system:
		//         - if the ConnectRetryTimer is running, stops and resets the
		//           ConnectRetryTimer (sets to zero),
		f.connectRetryTimer.Stop()
		//         - if the DelayOpenTimer is running, stops and resets the
		//           DelayOpenTimer (sets to zero),
		f.delayOpenTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - performs peer oscillation damping if the DampPeerOscillations
		//           attribute is set to True, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	}
}

//    Active State:
//       In this state, BGP FSM is trying to acquire a peer by listening
//       for, and accepting, a TCP connection.
func (f *fsm) active(event int) {
	switch event {
	//       The start events (Events 1, 3-7) are ignored in the Active state.
	case manualStart:
	case automaticStart:
	case manualStartWithPassiveTCPEstablishment:
	case automaticStartWithPassiveTCPEstablishment:
	case automaticStartWithDampPeerOscillations:
	case automaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case manualStop:
		//       In response to a ManualStop event (Event 2), the local system:
		//         - If the DelayOpenTimer is running and the
		//           SendNOTIFICATIONwithoutOPEN session attribute is set, the
		//           local system sends a NOTIFICATION with a Cease,
		if f.delayOpenTimer.Running() && f.sendNotificationwithoutOpen {
			f.notification(cease, noErrorSubcode, nil)
		}
		//         - releases all BGP resources including stopping the
		//           DelayOpenTimer
		f.release()
		f.delayOpenTimer.Stop()
		//         - drops the TCP connection,
		f.drop()
		//         - sets ConnectRetryCounter to zero,
		f.connectRetryCounter = 0
		//         - stops the ConnectRetryTimer and sets the ConnectRetryTimer to
		//           zero, and
		f.connectRetryTimer.Stop()
		//         - changes its state to Idle.
		f.transition(idle)
	case connectRetryTimerExpires:
		//       In response to a ConnectRetryTimer_Expires event (Event 9), the
		//       local system:
		//         - restarts the ConnectRetryTimer (with initial value),
		f.connectRetryTimer.Reset(f.connectRetryTime)
		//         - initiates a TCP connection to the other BGP peer,
		f.dial()
		//         - continues to listen for a TCP connection that may be initiated
		//           by a remote BGP peer, and
		//         - changes its state to Connect.
		f.transition(connect)
	case delayOpenTimerExpires:
		//       If the local system receives a DelayOpenTimer_Expires event (Event
		//       12), the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - stops and clears the DelayOpenTimer (set to zero),
		f.delayOpenTimer.Stop()
		//         - completes the BGP initialization,
		//         - sends the OPEN message to its remote peer,
		f.open()
		//         - sets its hold timer to a large value, and
		//         - changes its state to OpenSent.
		f.transition(openSent)
		//       A HoldTimer value of 4 minutes is also suggested for this state
		//       transition.
	case tcpConnectionValid:
		//       If the local system receives a TcpConnection_Valid event (Event
		//       14), the local system processes the TCP connection flags and stays
		//       in the Active state.
	case tcpCRInvalid:
		//       If the local system receives a Tcp_CR_Invalid event (Event 15),
		//       the local system rejects the TCP connection and stays in the
		//       Active State.
	case tcpCRAcked:
		//       In response to the success of a TCP connection (Event 16 or Event
		//       17), the local system checks the DelayOpen optional attribute
		//       prior to processing.

		//         If the DelayOpen attribute is set to TRUE, the local system:
		if f.delayOpen {
			//           - stops the ConnectRetryTimer and sets the ConnectRetryTimer
			//             to zero,
			f.connectRetryTimer.Stop()
			//           - sets the DelayOpenTimer to the initial value
			//             (DelayOpenTime), and
			f.delayOpenTimer.Reset(f.delayOpenTime)
			//           - stays in the Active state.
		} else {
			//         If the DelayOpen attribute is set to FALSE, the local system:
			//           - sets the ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//           - completes the BGP initialization,
			//           - sends the OPEN message to its peer,
			f.open()
			//           - sets its HoldTimer to a large value, and
			f.holdTimer.Reset(largeHoldTimer)
			//           - changes its state to OpenSent.
			f.transition(openSent)
		}
		//       A HoldTimer value of 4 minutes is suggested as a "large value" for
		//       the HoldTimer.
	case tcpConnectionConfirmed:
		//       In response to the success of a TCP connection (Event 16 or Event
		//       17), the local system checks the DelayOpen optional attribute
		//       prior to processing.

		//         If the DelayOpen attribute is set to TRUE, the local system:
		if f.delayOpen {
			//           - stops the ConnectRetryTimer and sets the ConnectRetryTimer
			//             to zero,
			f.connectRetryTimer.Stop()
			//           - sets the DelayOpenTimer to the initial value
			//             (DelayOpenTime), and
			f.delayOpenTimer.Reset(f.delayOpenTime)
			//           - stays in the Active state.
		} else {
			//         If the DelayOpen attribute is set to FALSE, the local system:
			//           - sets the ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//           - completes the BGP initialization,
			//           - sends the OPEN message to its peer,
			f.open()
			//           - sets its HoldTimer to a large value, and
			f.holdTimer.Reset(largeHoldTimer)
			//           - changes its state to OpenSent.
			f.transition(openSent)
			return
		}
		//       A HoldTimer value of 4 minutes is suggested as a "large value" for
		//       the HoldTimer.
	case tcpConnectionFails:
		//       If the local system receives a TcpConnectionFails event (Event
		//       18), the local system:
		//         - restarts the ConnectRetryTimer (with the initial value),
		f.connectRetryTimer.Reset(f.connectRetryTime)
		//         - stops and clears the DelayOpenTimer (sets the value to zero),
		f.delayOpenTimer.Stop()
		//         - releases all BGP resource,
		f.release()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - optionally performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case bgpOpenWithDelayOpenTimerRunning:
		//       If an OPEN message is received and the DelayOpenTimer is running
		//       (Event 20), the local system:
		// TODO: How to check if OPEN message received?
		if f.delayOpenTimer.Running() {
			//         - stops the ConnectRetryTimer (if running) and sets the
			//           ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//         - stops and clears the DelayOpenTimer (sets to zero),
			f.delayOpenTimer.Stop()
			//         - completes the BGP initialization,
			//         - sends an OPEN message,
			f.open()
			//         - sends a KEEPALIVE message,
			f.keepalive()
			//         - if the HoldTimer value is non-zero,
			if f.holdTimer.Running() {
				//             - starts the KeepaliveTimer to initial value,
				//             - resets the HoldTimer to the negotiated value,
				f.holdTimer.Reset(f.holdTime)
			} else {
				//           else if the HoldTimer is zero
				//             - resets the KeepaliveTimer (set to zero),
				//             - resets the HoldTimer to zero, and
				f.holdTimer.Stop()
			}
			//         - changes its state to OpenConfirm.
			f.transition(openConfirm)
			//       If the value of the autonomous system field is the same as the
			//       local Autonomous System number, set the connection status to an
			//       internal connection; otherwise it will be external.
		}
	case bgpHeaderErr:
		//       If BGP message header checking (Event 21) or OPEN message checking
		//       detects an error (Event 22) (see Section 6.2), the local system:
		//         - (optionally) sends a NOTIFICATION message with the appropriate
		//           error code if the SendNOTIFICATIONwithoutOPEN attribute is set
		//           to TRUE,
		if f.sendNotificationwithoutOpen {
			f.notification(messageHeaderError, noErrorSubcode, nil)
		}
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case bgpOpenMsgErr:
		//       If BGP message header checking (Event 21) or OPEN message checking
		//       detects an error (Event 22) (see Section 6.2), the local system:
		//         - (optionally) sends a NOTIFICATION message with the appropriate
		//           error code if the SendNOTIFICATIONwithoutOPEN attribute is set
		//           to TRUE,
		if f.sendNotificationwithoutOpen {
			f.notification(openMessageError, noErrorSubcode, nil)
		}
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case notifMsgVerErr:
		//       If a NOTIFICATION message is received with a version error (Event
		//       24), the local system checks the DelayOpenTimer.  If the
		//       DelayOpenTimer is running, the local system:
		if f.delayOpenTimer.Running() {
			//         - stops the ConnectRetryTimer (if running) and sets the
			//           ConnectRetryTimer to zero,
			f.connectRetryTimer.Stop()
			//         - stops and resets the DelayOpenTimer (sets to zero),
			f.delayOpenTimer.Stop()
			//         - releases all BGP resources,
			f.release()
			//         - drops the TCP connection, and
			f.drop()
			//         - changes its state to Idle.
			f.transition(idle)
			return
		}
		//       If the DelayOpenTimer is not running, the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	default:
		//       In response to any other event (Events 8, 10-11, 13, 19, 23,
		//       25-28), the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by one,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	}
}

//    OpenSent:
//       In this state, BGP FSM waits for an OPEN message from its peer.
func (f *fsm) openSent(event int) {
	switch event {
	//       The start events (Events 1, 3-7) are ignored in the OpenSent state.
	case manualStart:
	case automaticStart:
	case manualStartWithPassiveTCPEstablishment:
	case automaticStartWithPassiveTCPEstablishment:
	case automaticStartWithDampPeerOscillations:
	case automaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case manualStop:
		//       If a ManualStop event (Event 2) is issued in the OpenSent state,
		//       the local system:
		//         - sends the NOTIFICATION with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - sets the ConnectRetryCounter to zero, and
		f.connectRetryCounter = 0
		//         - changes its state to Idle.
		f.transition(idle)
	case automaticStop:
		//       If an AutomaticStop event (Event 8) is issued in the OpenSent
		//       state, the local system:
		//         - sends the NOTIFICATION with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all the BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case holdTimerExpires:
		//       If the HoldTimer_Expires (Event 10), the local system:
		//         - sends a NOTIFICATION message with the error code Hold Timer
		//           Expired,
		f.notification(holdTimerExpired, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case tcpConnectionValid:
		//       If a TcpConnection_Valid (Event 14), Tcp_CR_Acked (Event 16), or a
		//       TcpConnectionConfirmed event (Event 17) is received, a second TCP
		//       connection may be in progress.  This second TCP connection is
		//       tracked per Connection Collision processing (Section 6.8) until an
		//       OPEN message is received.
	case tcpCRAcked:
		//       If a TcpConnection_Valid (Event 14), Tcp_CR_Acked (Event 16), or a
		//       TcpConnectionConfirmed event (Event 17) is received, a second TCP
		//       connection may be in progress.  This second TCP connection is
		//       tracked per Connection Collision processing (Section 6.8) until an
		//       OPEN message is received.
	case tcpConnectionConfirmed:
		//       If a TcpConnection_Valid (Event 14), Tcp_CR_Acked (Event 16), or a
		//       TcpConnectionConfirmed event (Event 17) is received, a second TCP
		//       connection may be in progress.  This second TCP connection is
		//       tracked per Connection Collision processing (Section 6.8) until an
		//       OPEN message is received.
	case tcpCRInvalid:
		//       A TCP Connection Request for an Invalid port (Tcp_CR_Invalid
		//       (Event 15)) is ignored.
	case tcpConnectionFails:
		//       If a TcpConnectionFails event (Event 18) is received, the local
		//       system:
		//         - closes the BGP connection,
		//         - restarts the ConnectRetryTimer,
		f.connectRetryTimer.Reset(f.connectRetryTime)
		//         - continues to listen for a connection that may be initiated by
		//           the remote BGP peer, and
		//         - changes its state to Active.
		f.transition(active)
	case bgpOpen:
		//       When an OPEN message is received, all fields are checked for
		//       correctness.  If there are no errors in the OPEN message (Event
		//       19), the local system:
		//         - resets the DelayOpenTimer to zero,
		f.delayOpenTimer.Reset(f.delayOpenTime)
		//         - sets the BGP ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - sends a KEEPALIVE message, and
		f.keepalive()
		//         - sets a KeepaliveTimer (via the text below)
		//         - sets the HoldTimer according to the negotiated value (see
		//           Section 4.2),
		if f.holdTime != 0 {
			f.holdTimer.Reset(f.holdTime)
		}
		//         - changes its state to OpenConfirm.
		f.transition(openConfirm)

		//       If the negotiated hold time value is zero, then the HoldTimer and
		//       KeepaliveTimer are not started.  If the value of the Autonomous
		//       System field is the same as the local Autonomous System number,
		//       then the connection is an "internal" connection; otherwise, it is
		//       an "external" connection.  (This will impact UPDATE processing as
		//       described below.)
	case bgpOpenMsgErr:
		//       If the BGP message header checking (Event 21) or OPEN message
		//       checking detects an error (Event 22)(see Section 6.2), the local
		//       system:
		//         - sends a NOTIFICATION message with the appropriate error code,
		f.notification(openMessageError, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
		//       Collision detection mechanisms (Section 6.8) need to be applied
		//       when a valid BGP OPEN message is received (Event 19 or Event 20).
		//       Please refer to Section 6.8 for the details of the comparison.  A
		//       CollisionDetectDump event occurs when the BGP implementation
		//       determines, by means outside the scope of this document, that a
		//       connection collision has occurred.
	case openCollisionDump:
		//       If a connection in the OpenSent state is determined to be the
		//       connection that must be closed, an OpenCollisionDump (Event 23) is
		//       signaled to the state machine.  If such an event is received in
		//       the OpenSent state, the local system:
		//         - sends a NOTIFICATION with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case notifMsgVerErr:
		//       If a NOTIFICATION message is received with a version error (Event
		//       24), the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection, and
		f.drop()
		//         - changes its state to Idle.
		f.transition(idle)
	default:
		//       In response to any other event (Events 9, 11-13, 20, 25-28), the
		//       local system:
		//         - sends the NOTIFICATION with the Error Code Finite State
		//           Machine Error,
		f.notification(finiteStateMachineError, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	}
}

//    OpenConfirm State:
//       In this state, BGP waits for a KEEPALIVE or NOTIFICATION message.
func (f *fsm) openConfirm(event int) {
	switch event {
	//       The start events (Events 1, 3-7) are ignored in the OpenConfirm state.
	case manualStart:
	case automaticStart:
	case manualStartWithPassiveTCPEstablishment:
	case automaticStartWithPassiveTCPEstablishment:
	case automaticStartWithDampPeerOscillations:
	case automaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case manualStop:
		//       In response to a ManualStop event (Event 2) initiated by the
		//       operator, the local system:
		//         - sends the NOTIFICATION message with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - sets the ConnectRetryCounter to zero,
		f.connectRetryCounter = 0
		//         - sets the ConnectRetryTimer to zero, and
		f.connectRetryTimer.Stop()
		//         - changes its state to Idle.
		f.transition(idle)
	case automaticStop:
		//       In response to the AutomaticStop event initiated by the system
		//       (Event 8), the local system:
		//         - sends the NOTIFICATION message with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case holdTimerExpires:
		//       If the HoldTimer_Expires event (Event 10) occurs before a
		//       KEEPALIVE message is received, the local system:
		//         - sends the NOTIFICATION message with the Error Code Hold Timer
		//           Expired,
		f.notification(holdTimerExpired, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case keepaliveTimerExpires:
		//       If the local system receives a KeepaliveTimer_Expires event (Event
		//       11), the local system:
		//         - sends a KEEPALIVE message,
		f.keepalive()
		//         - restarts the KeepaliveTimer, and
		//         - remains in the OpenConfirmed state.
	case tcpConnectionValid:
		//       In the event of a TcpConnection_Valid event (Event 14), or the
		//       success of a TCP connection (Event 16 or Event 17) while in
		//       OpenConfirm, the local system needs to track the second
		//       connection.
	case tcpCRAcked:
		//       In the event of a TcpConnection_Valid event (Event 14), or the
		//       success of a TCP connection (Event 16 or Event 17) while in
		//       OpenConfirm, the local system needs to track the second
		//       connection.
	case tcpConnectionConfirmed:
		//       In the event of a TcpConnection_Valid event (Event 14), or the
		//       success of a TCP connection (Event 16 or Event 17) while in
		//       OpenConfirm, the local system needs to track the second
		//       connection.
	case tcpCRInvalid:
		//       If a TCP connection is attempted with an invalid port (Event 15),
		//       the local system will ignore the second connection attempt.
	case tcpConnectionFails:
		//       If the local system receives a TcpConnectionFails event (Event 18)
		//       from the underlying TCP or a NOTIFICATION message (Event 25), the
		//       local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case notifMsg:
		//       If the local system receives a TcpConnectionFails event (Event 18)
		//       from the underlying TCP or a NOTIFICATION message (Event 25), the
		//       local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case notifMsgVerErr:
		//       If the local system receives a NOTIFICATION message with a version
		//       error (NotifMsgVerErr (Event 24)), the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection, and
		f.drop()
		//         - changes its state to Idle.
		f.transition(idle)
	case bgpOpen:
		//       If the local system receives a valid OPEN message (BGPOpen (Event
		//       19)), the collision detect function is processed per Section 6.8.
		//       If this connection is to be dropped due to connection collision,
		//       the local system:
		//         - sends a NOTIFICATION with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection (send TCP FIN),
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case bgpHeaderErr:
		//       If an OPEN message is received, all fields are checked for
		//       correctness.  If the BGP message header checking (BGPHeaderErr
		//       (Event 21)) or OPEN message checking detects an error (see Section
		//       6.2) (BGPOpenMsgErr (Event 22)), the local system:
		//         - sends a NOTIFICATION message with the appropriate error code,
		f.notification(messageHeaderError, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case bgpOpenMsgErr:
		//       If an OPEN message is received, all fields are checked for
		//       correctness.  If the BGP message header checking (BGPHeaderErr
		//       (Event 21)) or OPEN message checking detects an error (see Section
		//       6.2) (BGPOpenMsgErr (Event 22)), the local system:
		//         - sends a NOTIFICATION message with the appropriate error code,
		f.notification(openMessageError, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case openCollisionDump:
		//       If, during the processing of another OPEN message, the BGP
		//       implementation determines, by a means outside the scope of this
		//       document, that a connection collision has occurred and this
		//       connection is to be closed, the local system will issue an
		//       OpenCollisionDump event (Event 23).  When the local system
		//       receives an OpenCollisionDump event (Event 23), the local system:
		//         - sends a NOTIFICATION with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case keepAliveMsg:
		//       If the local system receives a KEEPALIVE message (KeepAliveMsg
		//       (Event 26)), the local system:
		//         - restarts the HoldTimer and
		f.holdTimer.Reset(f.holdTime)
		//         - changes its state to Established
		f.transition(established)
	default:
		//       In response to any other event (Events 9, 12-13, 20, 27-28), the
		//       local system:
		//         - sends a NOTIFICATION with a code of Finite State Machine
		//           Error,
		f.notification(finiteStateMachineError, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	}
}

//    Established State:
//       In the Established state, the BGP FSM can exchange UPDATE,
//       NOTIFICATION, and KEEPALIVE messages with its peer.
func (f *fsm) established(event int) {
	switch event {
	//       The start events (Events 1, 3-7) are ignored in the OpenConfirm state.
	case manualStart:
	case automaticStart:
	case manualStartWithPassiveTCPEstablishment:
	case automaticStartWithPassiveTCPEstablishment:
	case automaticStartWithDampPeerOscillations:
	case automaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case manualStop:
		//       In response to a ManualStop event (initiated by an operator)
		//       (Event 2), the local system:
		//         - sends the NOTIFICATION message with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - deletes all routes associated with this connection,
		//         - releases BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - sets the ConnectRetryCounter to zero, and
		f.connectRetryCounter = 0
		//          - changes its state to Idle.
		f.transition(idle)
	case automaticStop:
		//       In response to an AutomaticStop event (Event 8), the local system:
		//         - sends a NOTIFICATION with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero
		f.connectRetryTimer.Stop()
		//         - deletes all routes associated with this connection,
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)

		//       One reason for an AutomaticStop event is: A BGP receives an UPDATE
		//       messages with a number of prefixes for a given peer such that the
		//       total prefixes received exceeds the maximum number of prefixes
		//       configured.  The local system automatically disconnects the peer.
	case holdTimerExpires:
		//       If the HoldTimer_Expires event occurs (Event 10), the local
		//       system:
		//         - sends a NOTIFICATION message with the Error Code Hold Timer
		//           Expired,
		f.notification(holdTimerExpired, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case keepaliveTimerExpires:
		//       If the KeepaliveTimer_Expires event occurs (Event 11), the local
		//       system:
		//         - sends a KEEPALIVE message, and
		f.keepalive()
		//         - restarts its KeepaliveTimer, unless the negotiated HoldTime
		//           value is zero.

		//       Each time the local system sends a KEEPALIVE or UPDATE message, it
		//       restarts its KeepaliveTimer, unless the negotiated HoldTime value
		//       is zero.
	case tcpConnectionValid:
		//       A TcpConnection_Valid (Event 14), received for a valid port, will
		//       cause the second connection to be tracked.
	case tcpCRInvalid:
		//       An invalid TCP connection (Tcp_CR_Invalid event (Event 15)) will
		//       be ignored.
	case tcpCRAcked:
		//       In response to an indication that the TCP connection is
		//       successfully established (Event 16 or Event 17), the second
		//       connection SHALL be tracked until it sends an OPEN message.
	case tcpConnectionConfirmed:
		//       In response to an indication that the TCP connection is
		//       successfully established (Event 16 or Event 17), the second
		//       connection SHALL be tracked until it sends an OPEN message.
	case bgpOpen:
		//       If a valid OPEN message (BGPOpen (Event 19)) is received, and if
		//       the CollisionDetectEstablishedState optional attribute is TRUE,
		//       the OPEN message will be checked to see if it collides (Section
		//       6.8) with any other connection.  If the BGP implementation
		//       determines that this connection needs to be terminated, it will
		//       process an OpenCollisionDump event (Event 23).  If this connection
		//       needs to be terminated, the local system:
		//         - sends a NOTIFICATION with a Cease,
		f.notification(cease, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - deletes all routes associated with this connection,
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	case notifMsgVerErr:
		//       If the local system receives a NOTIFICATION message (Event 24 or
		//       Event 25) or a TcpConnectionFails (Event 18) from the underlying
		//       TCP, the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - deletes all routes associated with this connection,
		//         - releases all the BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - changes its state to Idle.
		f.transition(idle)
	case notifMsg:
		//       If the local system receives a NOTIFICATION message (Event 24 or
		//       Event 25) or a TcpConnectionFails (Event 18) from the underlying
		//       TCP, the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - deletes all routes associated with this connection,
		//         - releases all the BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - changes its state to Idle.
		f.transition(idle)
	case tcpConnectionFails:
		//       If the local system receives a NOTIFICATION message (Event 24 or
		//       Event 25) or a TcpConnectionFails (Event 18) from the underlying
		//       TCP, the local system:
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - deletes all routes associated with this connection,
		//         - releases all the BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - changes its state to Idle.
		f.transition(idle)
	case keepAliveMsg:
		//       If the local system receives a KEEPALIVE message (Event 26), the
		//       local system:
		//         - restarts its HoldTimer, if the negotiated HoldTime value is
		//           non-zero, and
		if f.holdTime != 0 {
			f.holdTimer.Reset(f.holdTime)
		}
		//         - remains in the Established state.
		f.transition(established)
	case updateMsg:
		//       If the local system receives an UPDATE message (Event 27), the
		//       local system:
		//         - processes the message,
		//         - restarts its HoldTimer, if the negotiated HoldTime value is
		//           non-zero, and
		if f.holdTime != 0 {
			f.holdTimer.Reset(f.holdTime)
		}
		//         - remains in the Established state.
		f.transition(established)
	case updateMsgErr:
		//       If the local system receives an UPDATE message, and the UPDATE
		//       message error handling procedure (see Section 6.3) detects an
		//       error (Event 28), the local system:
		//         - sends a NOTIFICATION message with an Update error,
		f.notification(updateMessageError, noErrorSubcode, nil)
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - deletes all routes associated with this connection,
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	default:
		//       In response to any other event (Events 9, 12-13, 20-22), the local
		//       system:
		//         - sends a NOTIFICATION message with the Error Code Finite State
		//           Machine Error,
		f.notification(finiteStateMachineError, noErrorSubcode, nil)
		//         - deletes all routes associated with this connection,
		//         - sets the ConnectRetryTimer to zero,
		f.connectRetryTimer.Stop()
		//         - releases all BGP resources,
		f.release()
		//         - drops the TCP connection,
		f.drop()
		//         - increments the ConnectRetryCounter by 1,
		f.connectRetryCounter++
		//         - (optionally) performs peer oscillation damping if the
		//           DampPeerOscillations attribute is set to TRUE, and
		if f.dampPeerOscillations {
			// TODO: Implement me
		}
		//         - changes its state to Idle.
		f.transition(idle)
	}
}

// 8.2.  Description of FSM

// 8.2.1.  FSM Definition

//    BGP MUST maintain a separate FSM for each configured peer.  Each BGP
//    peer paired in a potential connection will attempt to connect to the
//    other, unless configured to remain in the idle state, or configured
//    to remain passive.  For the purpose of this discussion, the active or
//    connecting side of the TCP connection (the side of a TCP connection
//    sending the first TCP SYN packet) is called outgoing.  The passive or
//    listening side (the sender of the first SYN/ACK) is called an
//    incoming connection.  (See Section 8.2.1.1 for information on the
//    terms active and passive used below.)

//    A BGP implementation MUST connect to and listen on TCP port 179 for
//    incoming connections in addition to trying to connect to peers.  For
//    each incoming connection, a state machine MUST be instantiated.
//    There exists a period in which the identity of the peer on the other
//    end of an incoming connection is known, but the BGP identifier is not
//    known.  During this time, both an incoming and outgoing connection
//    may exist for the same configured peering.  This is referred to as a
//    connection collision (see Section 6.8).

const port = 179

//    A BGP implementation will have, at most, one FSM for each configured
//    peering, plus one FSM for each incoming TCP connection for which the
//    peer has not yet been identified.  Each FSM corresponds to exactly
//    one TCP connection.

//    There may be more than one connection between a pair of peers if the
//    connections are configured to use a different pair of IP addresses.
//    This is referred to as multiple "configured peerings" to the same
//    peer.

// 8.2.1.1.  Terms "active" and "passive"

//    The terms active and passive have been in the Internet operator's
//    vocabulary for almost a decade and have proven useful.  The words
//    active and passive have slightly different meanings when applied to a
//    TCP connection or a peer.  There is only one active side and one
//    passive side to any one TCP connection, per the definition above and
//    the state machine below.  When a BGP speaker is configured as active,
//    it may end up on either the active or passive side of the connection
//    that eventually gets established.  Once the TCP connection is
//    completed, it doesn't matter which end was active and which was
//    passive.  The only difference is in which side of the TCP connection
//    has port number 179.

// 8.2.1.2.  FSM and Collision Detection

//    There is one FSM per BGP connection.  When the connection collision
//    occurs prior to determining what peer a connection is associated
//    with, there may be two connections for one peer.  After the
//    connection collision is resolved (see Section 6.8), the FSM for the
//    connection that is closed SHOULD be disposed.

// 8.2.1.3.  FSM and Optional Session Attributes

//    Optional Session Attributes specify either attributes that act as
//    flags (TRUE or FALSE) or optional timers.  For optional attributes
//    that act as flags, if the optional session attribute can be set to
//    TRUE on the system, the corresponding BGP FSM actions must be
//    supported.  For example, if the following options can be set in a BGP
//    implementation: AutoStart and PassiveTcpEstablishment, then Events 3,
//    4 and 5 must be supported.  If an Optional Session attribute cannot
//    be set to TRUE, the events supporting that set of options do not have
//    to be supported.

//    Each of the optional timers (DelayOpenTimer and IdleHoldTimer) has a
//    group of attributes that are:

//       - flag indicating support,
//       - Time set in Timer
//       - Timer.

//    The two optional timers show this format:

//       DelayOpenTimer: DelayOpen, DelayOpenTime, DelayOpenTimer
//       IdleHoldTimer:  DampPeerOscillations, IdleHoldTime,
//                       IdleHoldTimer

//    If the flag indicating support for an optional timer (DelayOpen or
//    DampPeerOscillations) cannot be set to TRUE, the timers and events
//    supporting that option do not have to be supported.

// 8.2.1.4.  FSM Event Numbers

//    The Event numbers (1-28) utilized in this state machine description
//    aid in specifying the behavior of the BGP state machine.
//    Implementations MAY use these numbers to provide network management
//    information.  The exact form of an FSM or the FSM events are specific
//    to each implementation.

// 8.2.1.5.  FSM Actions that are Implementation Dependent

//    At certain points, the BGP FSM specifies that BGP initialization will
//    occur or that BGP resources will be deleted.  The initialization of
//    the BGP FSM and the associated resources depend on the policy portion
//    of the BGP implementation.  The details of these actions are outside
//    the scope of the FSM document.

// 8.2.2.  Finite State Machine

func newFSM(remoteAS uint16, remoteIP net.IP) *fsm {
	f := &fsm{
		state: idle,
		acceptConnectionsUnconfiguredPeers: false,
		allowAutomaticStart:                true,
		allowAutomaticStop:                 true,
		collisionDetectEstablishedState:    false,
		dampPeerOscillations:               false,
		delayOpen:                          false,
		passiveTCPEstablishment:            false,
		sendNotificationwithoutOpen:        false,
		trackTCPState:                      false,
		peer:                               newPeer(remoteAS, remoteIP),
	}
	// Initialize timers
	f.delayOpenTimer = timer.New(defaultDelayOpenTime, f.sendEvent(delayOpenTimerExpires))
	f.delayOpenTimer.Stop()
	return f
}

func (f *fsm) transition(state int) {
	log.Printf("Transitioning from (%d) %s to (%d) %s", f.state, stateName[f.state], state, stateName[state])
	f.state = state
}
