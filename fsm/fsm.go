package fsm

import (
	"github.com/transitorykris/kbgp/bgp"	
	"github.com/transitorykris/kbgp/counter"	
	"github.com/transitorykris/kbgp/timer"
)

type fsm struct {
	state bgp.State

	adjRIBIn bgp.RIB
	adjRIBOut bgp.RIB

	connectRetryCounter counter.Counter

	connectRetryTimer timer.Timer
	delayOpenTimer timer.Timer
	holdTimer timer.Timer
	idleHoldTimer timer.Timer
	keepaliveTimer timer.Timer
}

// New creates a new instance of the FSM
func New() bgp.FSM {
	return new(fsm)
}

// Send implements bgp.FSM
func (f *fsm) Send(event bgp.Event) {
	switch f.state {
	case bgp.Idle:
		f.idle(event)
	case bgp.Connect:
		f.connect(event)
	case bgp.Active:
		f.active(event)
	case bgp.OpenConfirm:
		f.openConfirm(event)
	case bgp.Established:
		f.established(event)
	}
}

// State implements bgp.FSM
func (f *fsm) State() bgp.State { return f.state }

func (f *fsm) transition(state bgp.State) { f.state = state }

// idle state - Initially, the BGP peer FSM is in the Idle state.
func (f *fsm) idle(event bgp.Event) {
	switch event {
	case bgp.ManualStart:
		// In this state, BGP FSM refuses all incoming BGP connections for
		// this peer.  No resources are allocated to the peer.  In response
		// to a bgp.ManualStart event (Event 1) or an bgp.AutomaticStart event (Event
		// 3), the local system:
		//   - initializes all BGP resources for the peer connection,
		//   - sets ConnectRetryCounter to zero,
		//   - starts the ConnectRetryTimer with the initial value,
		//   - initiates a TCP connection to the other BGP peer,
		//   - listens for a connection that may be initiated by the remote
		//     BGP peer, and
		//   - changes its state to Connect.
	case bgp.AutomaticStart:
		// In this state, BGP FSM refuses all incoming BGP connections for
		// this peer.  No resources are allocated to the peer.  In response
		// to a bgp.ManualStart event (Event 1) or an bgp.AutomaticStart event (Event
		// 3), the local system:
		//   - initializes all BGP resources for the peer connection,
		//   - sets ConnectRetryCounter to zero,
		//   - starts the ConnectRetryTimer with the initial value,
		//   - initiates a TCP connection to the other BGP peer,
		//   - listens for a connection that may be initiated by the remote
		//     BGP peer, and
		//   - changes its state to Connect.
	case bgp.ManualStop:
		// The bgp.ManualStop event (Event 2) and bgp.AutomaticStop (Event 8) event
		// are ignored in the Idle state.
	case bgp.AutomaticStop:
		// The bgp.ManualStop event (Event 2) and bgp.AutomaticStop (Event 8) event
		// are ignored in the Idle state.
	case bgp.ManualStartWithPassiveTCPEstablishment:
		// In response to a bgp.ManualStart_with_PassiveTcpEstablishment event
		// (Event 4) or bgp.AutomaticStart_with_PassiveTcpEstablishment event
		// (Event 5), the local system:
		//   - initializes all BGP resources,
		//   - sets the ConnectRetryCounter to zero,
		//   - starts the ConnectRetryTimer with the initial value,
		//   - listens for a connection that may be initiated by the remote
		//     peer, and
		//   - changes its state to Active.
		// The exact value of the ConnectRetryTimer is a local matter, but it
		// SHOULD be sufficiently large to allow TCP initialization.
	case bgp.AutomaticStartWithPassiveTCPEstablishment:
		// In response to a bgp.ManualStart_with_PassiveTcpEstablishment event
		// (Event 4) or bgp.AutomaticStart_with_PassiveTcpEstablishment event
		// (Event 5), the local system:
		//   - initializes all BGP resources,
		//   - sets the ConnectRetryCounter to zero,
		//   - starts the ConnectRetryTimer with the initial value,
		//   - listens for a connection that may be initiated by the remote
		//     peer, and
		// NOTE: Listening is already happening, but may need a better mechanism
		// e.g. We add a this peer to the listener, remove it from the listener when
		// we shut down. Right now the listener scans through all FSMs.
		//   - changes its state to Active.
		// The exact value of the ConnectRetryTimer is a local matter, but it
		// SHOULD be sufficiently large to allow TCP initialization.
	case bgp.AutomaticStartWithDampPeerOscillations:
		// If the DampPeerOscillations attribute is set to TRUE, the
		// following three additional events may occur within the Idle state:
		//   - bgp.AutomaticStart_with_DampPeerOscillations (Event 6),
		//   - bgp.AutomaticStart_with_DampPeerOscillations_and_
		//     PassiveTcpEstablishment (Event 7),
		//   - IdleHoldTimer_Expires (Event 13).

		// Upon receiving these 3 events, the local system will use these
		// events to prevent peer oscillations.  The method of preventing
		// persistent peer oscillation is outside the scope of this document.
	case bgp.AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
		// If the DampPeerOscillations attribute is set to TRUE, the
		// following three additional events may occur within the Idle state:
		//   - bgp.AutomaticStart_with_DampPeerOscillations (Event 6),
		//   - bgp.AutomaticStart_with_DampPeerOscillations_and_
		//     PassiveTcpEstablishment (Event 7),
		//   - IdleHoldTimer_Expires (Event 13).

		// Upon receiving these 3 events, the local system will use these
		// events to prevent peer oscillations.  The method of preventing
		// persistent peer oscillation is outside the scope of this document.
	case bgp.HoldTimerExpires:
		// If the DampPeerOscillations attribute is set to TRUE, the
		// following three additional events may occur within the Idle state:
		//   - bgp.AutomaticStart_with_DampPeerOscillations (Event 6),
		//   - bgp.AutomaticStart_with_DampPeerOscillations_and_
		//     PassiveTcpEstablishment (Event 7),
		//   - IdleHoldTimer_Expires (Event 13).

		// Upon receiving these 3 events, the local system will use these
		// events to prevent peer oscillations.  The method of preventing
		// persistent peer oscillation is outside the scope of this document.
	default:
		// Any other event (Events 9-12, 15-28) received in the Idle state
		// does not cause change in the state of the local system.
	}
}

// connect -  In this state, BGP FSM is waiting for the TCP connection to be completed.
func (f *fsm) connect(event bgp.Event) {
	switch event {
	// The start events (Events 1, 3-7) are ignored in the Connect state.
	case bgp.ManualStart:
	case bgp.AutomaticStart:
	case bgp.ManualStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithDampPeerOscillations:
	case bgp.AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case bgp.ManualStop:
		// In response to a bgp.ManualStop event (Event 2), the local system:
		//   - drops the TCP connection,
		//   - releases all BGP resources,
		//   - sets ConnectRetryCounter to zero,
		//   - stops the ConnectRetryTimer and sets ConnectRetryTimer to
		//     zero, and
		//   - changes its state to Idle.
	case bgp.ConnectRetryTimerExpires:
		// In response to the ConnectRetryTimer_Expires event (Event 9), the
		// local system:
		//   - drops the TCP connection,
		//   - restarts the ConnectRetryTimer,
		//   - stops the DelayOpenTimer and resets the timer to zero,
		//   - initiates a TCP connection to the other BGP peer,
		//   - continues to listen for a connection that may be initiated by
		//     the remote BGP peer, and
		//   - stays in the Connect state.
	case bgp.DelayOpenTimerExpires:
		// If the DelayOpenTimer_Expires event (Event 12) occurs in the
		// Connect state, the local system:
		//   - sends an OPEN message to its peer,
		//   - sets the HoldTimer to a large value, and
		//   - changes its state to OpenSent.
	case bgp.TCPConnectionValid:
		// If the BGP FSM receives a TcpConnection_Valid event (Event 14),
		// the TCP connection is processed, and the connection remains in the
		// Connect state.
	case bgp.TCPCRInvalid:
		// If the BGP FSM receives a Tcp_CR_Invalid event (Event 15), the
		// local system rejects the TCP connection, and the connection
		// remains in the Connect state.
	case bgp.TCPCRAcked:
		// If the TCP connection succeeds (Event 16 or Event 17), the local
		// system checks the DelayOpen attribute prior to processing.  If the
		// DelayOpen attribute is set to TRUE, the local system:
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - sets the DelayOpenTimer to the initial value, and
		//   - stays in the Connect state.
		// If the DelayOpen attribute is set to FALSE, the local system:
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - completes BGP initialization
		//   - sends an OPEN message to its peer,
		//   - sets the HoldTimer to a large value, and
		//   - changes its state to OpenSent.

		// A HoldTimer value of 4 minutes is suggested.

		// If the TCP connection fails (Event 18), the local system checks
		// the DelayOpenTimer.  If the DelayOpenTimer is running, the local
		// system:
		//   - restarts the ConnectRetryTimer with the initial value,
		//   - stops the DelayOpenTimer and resets its value to zero,
		//   - continues to listen for a connection that may be initiated by
		//     the remote BGP peer, and
		//   - changes its state to Active.
		// If the DelayOpenTimer is not running, the local system:
		//   - stops the ConnectRetryTimer to zero,
		//   - drops the TCP connection,
		//   - releases all BGP resources, and
		//   - changes its state to Idle.
	case bgp.TCPConnectionConfirmed:
		// If the TCP connection succeeds (Event 16 or Event 17), the local
		// system checks the DelayOpen attribute prior to processing.  If the
		// DelayOpen attribute is set to TRUE, the local system:
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - sets the DelayOpenTimer to the initial value, and
		//   - stays in the Connect state.
		// If the DelayOpen attribute is set to FALSE, the local system:
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - completes BGP initialization
		//   - sends an OPEN message to its peer,
		//   - sets the HoldTimer to a large value, and
		//   - changes its state to OpenSent.

		// A HoldTimer value of 4 minutes is suggested.
	case bgp.TCPConnectionFails:
		// If the TCP connection fails (Event 18), the local system checks
		// the DelayOpenTimer.  If the DelayOpenTimer is running, the local
		// system:
		//   - restarts the ConnectRetryTimer with the initial value,
		//   - stops the DelayOpenTimer and resets its value to zero,
		//   - continues to listen for a connection that may be initiated by
		//     the remote BGP peer, and
		//   - changes its state to Active.
		// If the DelayOpenTimer is not running, the local system:
		//   - stops the ConnectRetryTimer to zero,
		//   - drops the TCP connection,
		//   - releases all BGP resources, and
		//   - changes its state to Idle.
	case bgp.BGPOpenWithDelayOpenTimerRunning:
		// If an OPEN message is received while the DelayOpenTimer is running
		// (Event 20), the local system:
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - completes the BGP initialization,
		//   - stops and clears the DelayOpenTimer (sets the value to zero),
		//   - sends an OPEN message,
		//   - sends a KEEPALIVE message,
		//   - if the HoldTimer initial value is non-zero,
		//       - starts the KeepaliveTimer with the initial value and
		//       - resets the HoldTimer to the negotiated value,
		//     else, if the HoldTimer initial value is zero,
		//       - resets the KeepaliveTimer and
		//       - resets the HoldTimer value to zero,
		//   - and changes its state to OpenConfirm.
		// If the value of the autonomous system field is the same as the
		// local Autonomous System number, set the connection status to an
		// internal connection; otherwise it will be "external".
	case bgp.BGPHeaderErr:
		// If BGP message header checking (Event 21) or OPEN message checking
		// detects an error (Event 22) (see Section 6.2), the local system:
		//   - (optionally) If the SendNOTIFICATIONwithoutOPEN attribute is
		//     set to TRUE, then the local system first sends a NOTIFICATION
		//     message with the appropriate error code, and then
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.BGPOpenMsgErr:
		// If BGP message header checking (Event 21) or OPEN message checking
		// detects an error (Event 22) (see Section 6.2), the local system:
		//   - (optionally) If the SendNOTIFICATIONwithoutOPEN attribute is
		//     set to TRUE, then the local system first sends a NOTIFICATION
		//     message with the appropriate error code, and then
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.NotifMsgVerErr:
		// If a NOTIFICATION message is received with a version error (Event
		// 24), the local system checks the DelayOpenTimer.  If the
		// DelayOpenTimer is running, the local system:
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - stops and resets the DelayOpenTimer (sets to zero),
		//   - releases all BGP resources,
		//   - drops the TCP connection, and
		//   - changes its state to Idle.
		// If the DelayOpenTimer is not running, the local system:
		//   - stops the ConnectRetryTimer and sets the ConnectRetryTimer to
		//     zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - performs peer oscillation damping if the DampPeerOscillations
		//     attribute is set to True, and
		//   - changes its state to Idle.
	default:
		// In response to any other events (Events 8, 10-11, 13, 19, 23,
		// 25-28), the local system:
		//   - if the ConnectRetryTimer is running, stops and resets the
		//     ConnectRetryTimer (sets to zero),
		//   - if the DelayOpenTimer is running, stops and resets the
		//     DelayOpenTimer (sets to zero),
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - performs peer oscillation damping if the DampPeerOscillations
		//     attribute is set to True, and
		//   - changes its state to Idle.
	}
}

// active - In this state, BGP FSM is trying to acquire a peer by listening
// for, and accepting, a TCP connection.
func (f *fsm) active(event bgp.Event) {
	switch event {
	// The start events (Events 1, 3-7) are ignored in the Active state.
	case bgp.ManualStart:
	case bgp.AutomaticStart:
	case bgp.ManualStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithDampPeerOscillations:
	case bgp.AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case bgp.ManualStop:
		// In response to a bgp.ManualStop event (Event 2), the local system:
		//   - If the DelayOpenTimer is running and the
		//     SendNOTIFICATIONwithoutOPEN session attribute is set, the
		//     local system sends a NOTIFICATION with a Cease,
		//   - releases all BGP resources including stopping the
		//     DelayOpenTimer
		//   - drops the TCP connection,
		//   - sets ConnectRetryCounter to zero,
		//   - stops the ConnectRetryTimer and sets the ConnectRetryTimer to
		//     zero, and
		//   - changes its state to Idle.
	case bgp.ConnectRetryTimerExpires:
		// In response to a ConnectRetryTimer_Expires event (Event 9), the
		// local system:
		//   - restarts the ConnectRetryTimer (with initial value),
		//   - initiates a TCP connection to the other BGP peer,
		//   - continues to listen for a TCP connection that may be initiated
		//     by a remote BGP peer, and
		//   - changes its state to Connect.
	case bgp.DelayOpenTimerExpires:
		// If the local system receives a DelayOpenTimer_Expires event (Event
		// 12), the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - stops and clears the DelayOpenTimer (set to zero),
		//   - completes the BGP initialization,
		//   - sends the OPEN message to its remote peer,
		//   - sets its hold timer to a large value, and
		//   - changes its state to OpenSent.
		// A HoldTimer value of 4 minutes is also suggested for this state
		// transition.
	case bgp.TCPConnectionValid:
		// If the local system receives a TcpConnection_Valid event (Event
		// 14), the local system processes the TCP connection flags and stays
		// in the Active state.
	case bgp.TCPCRInvalid:
		// If the local system receives a Tcp_CR_Invalid event (Event 15),
		// the local system rejects the TCP connection and stays in the
		// Active State.
	case bgp.TCPCRAcked:
		// In response to the success of a TCP connection (Event 16 or Event
		// 17), the local system checks the DelayOpen optional attribute
		// prior to processing.

		//   If the DelayOpen attribute is set to TRUE, the local system:
		//     - stops the ConnectRetryTimer and sets the ConnectRetryTimer
		//       to zero,
		//     - sets the DelayOpenTimer to the initial value
		//       (DelayOpenTime), and
		//     - stays in the Active state.
		//   If the DelayOpen attribute is set to FALSE, the local system:
		//     - sets the ConnectRetryTimer to zero,
		//     - completes the BGP initialization,
		//     - sends the OPEN message to its peer,
		//     - sets its HoldTimer to a large value, and
		//     - changes its state to OpenSent.
		// A HoldTimer value of 4 minutes is suggested as a "large value" for
		// the HoldTimer.
	case bgp.TCPConnectionConfirmed:
		// In response to the success of a TCP connection (Event 16 or Event
		// 17), the local system checks the DelayOpen optional attribute
		// prior to processing.

		//   If the DelayOpen attribute is set to TRUE, the local system:
		//     - stops the ConnectRetryTimer and sets the ConnectRetryTimer
		//       to zero,
		//     - sets the DelayOpenTimer to the initial value
		//       (DelayOpenTime), and
		//     - stays in the Active state.
		//   If the DelayOpen attribute is set to FALSE, the local system:
		//     - sets the ConnectRetryTimer to zero,
		//     - completes the BGP initialization,
		//     - sends the OPEN message to its peer,
		//     - sets its HoldTimer to a large value, and
		//     - changes its state to OpenSent.
		// A HoldTimer value of 4 minutes is suggested as a "large value" for
		// the HoldTimer.
	case bgp.TCPConnectionFails:
		// If the local system receives a bgp.TCPConnectionFails event (Event
		// 18), the local system:
		//   - restarts the ConnectRetryTimer (with the initial value),
		//   - stops and clears the DelayOpenTimer (sets the value to zero),
		//   - releases all BGP resource,
		//   - increments the ConnectRetryCounter by 1,
		//   - optionally performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.BGPOpenWithDelayOpenTimerRunning:
		// If an OPEN message is received and the DelayOpenTimer is running
		// (Event 20), the local system:
		// TODO: How to check if OPEN message received?
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - stops and clears the DelayOpenTimer (sets to zero),
		//   - completes the BGP initialization,
		//   - sends an OPEN message,
		//   - sends a KEEPALIVE message,
		//   - if the HoldTimer value is non-zero,
		//       - starts the KeepaliveTimer to initial value,
		//       - resets the HoldTimer to the negotiated value,
		//     else if the HoldTimer is zero
		//       - resets the KeepaliveTimer (set to zero),
		//       - resets the HoldTimer to zero, and
		//   - changes its state to OpenConfirm.
		// If the value of the autonomous system field is the same as the
		// local Autonomous System number, set the connection status to an
		// internal connection; otherwise it will be external.
	case bgp.BGPHeaderErr:
		// If BGP message header checking (Event 21) or OPEN message checking
		// detects an error (Event 22) (see Section 6.2), the local system:
		//   - (optionally) sends a NOTIFICATION message with the appropriate
		//     error code if the SendNOTIFICATIONwithoutOPEN attribute is set
		//     to TRUE,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.BGPOpenMsgErr:
		// If BGP message header checking (Event 21) or OPEN message checking
		// detects an error (Event 22) (see Section 6.2), the local system:
		//   - (optionally) sends a NOTIFICATION message with the appropriate
		//     error code if the SendNOTIFICATIONwithoutOPEN attribute is set
		//     to TRUE,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.NotifMsgVerErr:
		// If a NOTIFICATION message is received with a version error (Event
		// 24), the local system checks the DelayOpenTimer.  If the
		// DelayOpenTimer is running, the local system:
		//   - stops the ConnectRetryTimer (if running) and sets the
		//     ConnectRetryTimer to zero,
		//   - stops and resets the DelayOpenTimer (sets to zero),
		//   - releases all BGP resources,
		//   - drops the TCP connection, and
		//   - changes its state to Idle.
		// If the DelayOpenTimer is not running, the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and

		//   - changes its state to Idle.
	default:
		// In response to any other event (Events 8, 10-11, 13, 19, 23,
		// 25-28), the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by one,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	}
}

// openSent - In this state, BGP FSM waits for an OPEN message from its peer.
func (f *fsm) openSent(event bgp.Event) {
	switch event {
	// The start events (Events 1, 3-7) are ignored in the OpenSent state.
	case bgp.ManualStart:
	case bgp.AutomaticStart:
	case bgp.ManualStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithDampPeerOscillations:
	case bgp.AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case bgp.ManualStop:
		// If a bgp.ManualStop event (Event 2) is issued in the OpenSent state,
		// the local system:
		//   - sends the NOTIFICATION with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - sets the ConnectRetryCounter to zero, and
		//   - changes its state to Idle.
	case bgp.AutomaticStop:
		// If an bgp.AutomaticStop event (Event 8) is issued in the OpenSent
		// state, the local system:
		//   - sends the NOTIFICATION with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all the BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.HoldTimerExpires:
		// If the HoldTimer_Expires (Event 10), the local system:
		//   - sends a NOTIFICATION message with the error code Hold Timer
		//     Expired,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.TCPConnectionValid:
		// If a TcpConnection_Valid (Event 14), Tcp_CR_Acked (Event 16), or a
		// bgp.TCPConnectionConfirmed event (Event 17) is received, a second TCP
		// connection may be in progress.  This second TCP connection is
		// tracked per Connection Collision processing (Section 6.8) until an
		// OPEN message is received.
	case bgp.TCPCRAcked:
		// If a TcpConnection_Valid (Event 14), Tcp_CR_Acked (Event 16), or a
		// bgp.TCPConnectionConfirmed event (Event 17) is received, a second TCP
		// connection may be in progress.  This second TCP connection is
		// tracked per Connection Collision processing (Section 6.8) until an
		// OPEN message is received.
	case bgp.TCPConnectionConfirmed:
		// If a TcpConnection_Valid (Event 14), Tcp_CR_Acked (Event 16), or a
		// bgp.TCPConnectionConfirmed event (Event 17) is received, a second TCP
		// connection may be in progress.  This second TCP connection is
		// tracked per Connection Collision processing (Section 6.8) until an
		// OPEN message is received.
	case bgp.TCPCRInvalid:
		// A TCP Connection Request for an Invalid port (Tcp_CR_Invalid
		// (Event 15)) is ignored.
	case bgp.TCPConnectionFails:
		// If a bgp.TCPConnectionFails event (Event 18) is received, the local
		// system:
		//   - closes the BGP connection,
		//   - restarts the ConnectRetryTimer,
		//   - continues to listen for a connection that may be initiated by
		//     the remote BGP peer, and
		//   - changes its state to Active.
	case bgp.BGPOpen:
		// When an OPEN message is received, all fields are checked for
		// correctness.  If there are no errors in the OPEN message (Event
		// 19), the local system:
		//   - resets the DelayOpenTimer to zero,
		//   - sets the BGP ConnectRetryTimer to zero,
		//   - sends a KEEPALIVE message, and
		//   - sets a KeepaliveTimer (via the text below)
		//   - sets the HoldTimer according to the negotiated value (see
		//     Section 4.2),
		//   - changes its state to OpenConfirm.
		// If the negotiated hold time value is zero, then the HoldTimer and
		// KeepaliveTimer are not started.  If the value of the Autonomous
		// System field is the same as the local Autonomous System number,
		// then the connection is an "internal" connection; otherwise, it is
		// an "external" connection.  (This will impact UPDATE processing as
		// described below.)
	case bgp.BGPOpenMsgErr:
		// If the BGP message header checking (Event 21) or OPEN message
		// checking detects an error (Event 22)(see Section 6.2), the local
		// system:
		//   - sends a NOTIFICATION message with the appropriate error code,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is TRUE, and
		//   - changes its state to Idle.
		// Collision detection mechanisms (Section 6.8) need to be applied
		// when a valid BGP OPEN message is received (Event 19 or Event 20).
		// Please refer to Section 6.8 for the details of the comparison.  A
		// CollisionDetectDump event occurs when the BGP implementation
		// determines, by means outside the scope of this document, that a
		// connection collision has occurred.
	case bgp.OpenCollisionDump:
		// If a connection in the OpenSent state is determined to be the
		// connection that must be closed, an bgp.OpenCollisionDump (Event 23) is
		// signaled to the state machine.  If such an event is received in
		// the OpenSent state, the local system:
		//   - sends a NOTIFICATION with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.NotifMsgVerErr:
		// If a NOTIFICATION message is received with a version error (Event
		// 24), the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection, and
		//   - changes its state to Idle.
	default:
		// In response to any other event (Events 9, 11-13, 20, 25-28), the
		// local system:
		//   - sends the NOTIFICATION with the Error Code Finite State
		//     Machine Error,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	}
}

// openConfirm - In this state, BGP waits for a KEEPALIVE or NOTIFICATION message.
func (f *fsm) openConfirm(event bgp.Event) {
	switch event {
	// The start events (Events 1, 3-7) are ignored in the OpenConfirm state.
	case bgp.ManualStart:
	case bgp.AutomaticStart:
	case bgp.ManualStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithDampPeerOscillations:
	case bgp.AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case bgp.ManualStop:
		// In response to a bgp.ManualStop event (Event 2) initiated by the
		// operator, the local system:
		//   - sends the NOTIFICATION message with a Cease,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - sets the ConnectRetryCounter to zero,
		//   - sets the ConnectRetryTimer to zero, and
		//   - changes its state to Idle.
	case bgp.AutomaticStop:
		// In response to the bgp.AutomaticStop event initiated by the system
		// (Event 8), the local system:
		//   - sends the NOTIFICATION message with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.HoldTimerExpires:
		// If the HoldTimer_Expires event (Event 10) occurs before a
		// KEEPALIVE message is received, the local system:
		//   - sends the NOTIFICATION message with the Error Code Hold Timer
		//     Expired,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.KeepaliveTimerExpires:
		// If the local system receives a KeepaliveTimer_Expires event (Event
		// 11), the local system:
		//   - sends a KEEPALIVE message,
		//   - restarts the KeepaliveTimer, and
		//   - remains in the OpenConfirmed state.
	case bgp.TCPConnectionValid:
		// In the event of a TcpConnection_Valid event (Event 14), or the
		// success of a TCP connection (Event 16 or Event 17) while in
		// OpenConfirm, the local system needs to track the second
		// connection.
	case bgp.TCPCRAcked:
		// In the event of a TcpConnection_Valid event (Event 14), or the
		// success of a TCP connection (Event 16 or Event 17) while in
		// OpenConfirm, the local system needs to track the second
		// connection.
	case bgp.TCPConnectionConfirmed:
		// In the event of a TcpConnection_Valid event (Event 14), or the
		// success of a TCP connection (Event 16 or Event 17) while in
		// OpenConfirm, the local system needs to track the second
		// connection.
	case bgp.TCPCRInvalid:
		// If a TCP connection is attempted with an invalid port (Event 15),
		// the local system will ignore the second connection attempt.
	case bgp.TCPConnectionFails:
		// If the local system receives a bgp.TCPConnectionFails event (Event 18)
		// from the underlying TCP or a NOTIFICATION message (Event 25), the
		// local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.NotifMsg:
		// If the local system receives a bgp.TCPConnectionFails event (Event 18)
		// from the underlying TCP or a NOTIFICATION message (Event 25), the
		// local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.NotifMsgVerErr:
		// If the local system receives a NOTIFICATION message with a version
		// error (bgp.NotifMsgVerErr (Event 24)), the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection, and
		//   - changes its state to Idle.
	case bgp.BGPOpen:
		// If the local system receives a valid OPEN message (bgp.BGPOpen (Event
		// 19)), the collision detect function is processed per Section 6.8.
		// If this connection is to be dropped due to connection collision,
		// the local system:
		//   - sends a NOTIFICATION with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection (send TCP FIN),
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.BGPHeaderErr:
		// If an OPEN message is received, all fields are checked for
		// correctness.  If the BGP message header checking (bgp.BGPHeaderErr
		// (Event 21)) or OPEN message checking detects an error (see Section
		// 6.2) (bgp.BGPOpenMsgErr (Event 22)), the local system:
		//   - sends a NOTIFICATION message with the appropriate error code,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.BGPOpenMsgErr:
		// If an OPEN message is received, all fields are checked for
		// correctness.  If the BGP message header checking (bgp.BGPHeaderErr
		// (Event 21)) or OPEN message checking detects an error (see Section
		// 6.2) (bgp.BGPOpenMsgErr (Event 22)), the local system:
		//   - sends a NOTIFICATION message with the appropriate error code,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.OpenCollisionDump:
		// If, during the processing of another OPEN message, the BGP
		// implementation determines, by a means outside the scope of this
		// document, that a connection collision has occurred and this
		// connection is to be closed, the local system will issue an
		// bgp.OpenCollisionDump event (Event 23).  When the local system
		// receives an bgp.OpenCollisionDump event (Event 23), the local system:
		//   - sends a NOTIFICATION with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.KeepAliveMsg:
		// If the local system receives a KEEPALIVE message (bgp.KeepAliveMsg
		// (Event 26)), the local system:
		//   - restarts the HoldTimer and
		//   - changes its state to Established
	default:
		// In response to any other event (Events 9, 12-13, 20, 27-28), the
		// local system:
		//   - sends a NOTIFICATION with a code of Finite State Machine
		//     Error,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	}
}

// established - In the Established state, the BGP FSM can exchange UPDATE,
// NOTIFICATION, and KEEPALIVE messages with its peer.
func (f *fsm) established(event bgp.Event) {
	switch event {
	// The start events (Events 1, 3-7) are ignored in the OpenConfirm state.
	case bgp.ManualStart:
	case bgp.AutomaticStart:
	case bgp.ManualStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithPassiveTCPEstablishment:
	case bgp.AutomaticStartWithDampPeerOscillations:
	case bgp.AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	case bgp.ManualStop:
		// In response to a bgp.ManualStop event (initiated by an operator)
		// (Event 2), the local system:
		//   - sends the NOTIFICATION message with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - deletes all routes associated with this connection,
		//   - releases BGP resources,
		//   - drops the TCP connection,
		//   - sets the ConnectRetryCounter to zero, and
		//    - changes its state to Idle.
	case bgp.AutomaticStop:
		// In response to an bgp.AutomaticStop event (Event 8), the local system:
		//   - sends a NOTIFICATION with a Cease,
		//   - sets the ConnectRetryTimer to zero
		//   - deletes all routes associated with this connection,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.

		// One reason for an bgp.AutomaticStop event is: A BGP receives an UPDATE
		// messages with a number of prefixes for a given peer such that the
		// total prefixes received exceeds the maximum number of prefixes
		// configured.  The local system automatically disconnects the peer.
	case bgp.HoldTimerExpires:
		// If the HoldTimer_Expires event occurs (Event 10), the local
		// system:
		//   - sends a NOTIFICATION message with the Error Code Hold Timer
		//     Expired,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.KeepaliveTimerExpires:
		// If the KeepaliveTimer_Expires event occurs (Event 11), the local
		// system:
		//   - sends a KEEPALIVE message, and
		//   - restarts its KeepaliveTimer, unless the negotiated HoldTime
		//     value is zero.

		// Each time the local system sends a KEEPALIVE or UPDATE message, it
		// restarts its KeepaliveTimer, unless the negotiated HoldTime value
		// is zero.
	case bgp.TCPConnectionValid:
		// A TcpConnection_Valid (Event 14), received for a valid port, will
		// cause the second connection to be tracked.
	case bgp.TCPCRInvalid:
		// An invalid TCP connection (Tcp_CR_Invalid event (Event 15)) will
		// be ignored.
	case bgp.TCPCRAcked:
		// In response to an indication that the TCP connection is
		// successfully established (Event 16 or Event 17), the second
		// connection SHALL be tracked until it sends an OPEN message.
	case bgp.TCPConnectionConfirmed:
		// In response to an indication that the TCP connection is
		// successfully established (Event 16 or Event 17), the second
		// connection SHALL be tracked until it sends an OPEN message.
	case bgp.BGPOpen:
		// If a valid OPEN message (bgp.BGPOpen (Event 19)) is received, and if
		// the CollisionDetectEstablishedState optional attribute is TRUE,
		// the OPEN message will be checked to see if it collides (Section
		// 6.8) with any other connection.  If the BGP implementation
		// determines that this connection needs to be terminated, it will
		// process an bgp.OpenCollisionDump event (Event 23).  If this connection
		// needs to be terminated, the local system:
		//   - sends a NOTIFICATION with a Cease,
		//   - sets the ConnectRetryTimer to zero,
		//   - deletes all routes associated with this connection,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations is set to TRUE, and
		//   - changes its state to Idle.
	case bgp.NotifMsgVerErr:
		// If the local system receives a NOTIFICATION message (Event 24 or
		// Event 25) or a bgp.TCPConnectionFails (Event 18) from the underlying
		// TCP, the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - deletes all routes associated with this connection,
		//   - releases all the BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - changes its state to Idle.
	case bgp.NotifMsg:
		// If the local system receives a NOTIFICATION message (Event 24 or
		// Event 25) or a bgp.TCPConnectionFails (Event 18) from the underlying
		// TCP, the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - deletes all routes associated with this connection,
		//   - releases all the BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - changes its state to Idle.
	case bgp.TCPConnectionFails:
		// If the local system receives a NOTIFICATION message (Event 24 or
		// Event 25) or a bgp.TCPConnectionFails (Event 18) from the underlying
		// TCP, the local system:
		//   - sets the ConnectRetryTimer to zero,
		//   - deletes all routes associated with this connection,
		//   - releases all the BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - changes its state to Idle.
	case bgp.KeepAliveMsg:
		// If the local system receives a KEEPALIVE message (Event 26), the
		// local system:
		//   - restarts its HoldTimer, if the negotiated HoldTime value is
		//     non-zero, and
		//   - remains in the Established state.
	case bgp.UpdateMsg:
		// If the local system receives an UPDATE message (Event 27), the
		// local system:
		//   - processes the message,
		//   - restarts its HoldTimer, if the negotiated HoldTime value is
		//     non-zero, and
		//   - remains in the Established state.
	case bgp.UpdateMsgErr:
		// If the local system receives an UPDATE message, and the UPDATE
		// message error handling procedure (see Section 6.3) detects an
		// error (Event 28), the local system:
		//   - sends a NOTIFICATION message with an Update error,
		//   - sets the ConnectRetryTimer to zero,
		//   - deletes all routes associated with this connection,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	default:
		// In response to any other event (Events 9, 12-13, 20-22), the local
		// system:
		//   - sends a NOTIFICATION message with the Error Code Finite State
		//     Machine Error,
		//   - deletes all routes associated with this connection,
		//   - sets the ConnectRetryTimer to zero,
		//   - releases all BGP resources,
		//   - drops the TCP connection,
		//   - increments the ConnectRetryCounter by 1,
		//   - (optionally) performs peer oscillation damping if the
		//     DampPeerOscillations attribute is set to TRUE, and
		//   - changes its state to Idle.
	}
}
