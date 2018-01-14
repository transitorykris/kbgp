package kbgp

import (
	"log"
	"time"

	"github.com/transitorykris/kbgp/counter"
	"github.com/transitorykris/kbgp/timer"
)

type state int

const (
	idle = iota
	connect
	active
	openSent
	openConfirm
	established
)

var stateLookup = map[state]string{
	idle:        "Idle",
	connect:     "Connect",
	active:      "Active",
	openSent:    "OpenSent",
	openConfirm: "OpenConfirm",
	established: "Established",
}

// String implements string.Stringer
func (s state) String() string {
	return stateLookup[s]
}

type fsm struct {
	// Mandatory session attributes
	// https://tools.ietf.org/html/rfc4271#section-8
	state               state
	connectRetryCounter counter.Counter
	connectRetryTimer   *timer.Timer
	connectRetryTime    time.Duration
	holdTimer           *timer.Timer
	holdTime            time.Duration
	keepaliveTimer      *timer.Timer
	keepaliveTime       time.Duration

	// Optional session attributes
	// https://tools.ietf.org/html/rfc4271#section-8
	// acceptConnectionsUnconfiguredPeers bool
	// allowAutomaticStart                bool
	// allowAutomaticStop                 bool
	// collisionDetectEstablishedState    bool
	// dampPeerOscillations               bool
	delayOpen bool
	// delayOpenTime                      time.Duration
	// delayOpenTimer                     timer.Timer
	// idleHoldTime                       time.Duration
	// idleHoldTimer                      timer.Timer
	// passiveTcpEstablishment            bool
	sendNOTIFICATIONwithoutOPEN bool
	// trackTcpState                      bool

	// reference back to our owner
	peer *Peer
}

// https://tools.ietf.org/html/rfc4271#section-10
const defaultConnectRetryTime = 120 * time.Second
const defaultHoldTime = 90 * time.Second
const defaultKeepaliveTime = defaultHoldTime / 3

func newFSM(p *Peer) *fsm {
	f := &fsm{
		peer:          p,
		holdTime:      defaultHoldTime,
		keepaliveTime: defaultKeepaliveTime,
	}
	f.connectRetryTimer = timer.New(defaultConnectRetryTime, f.eventWrapper(ConnectRetryTimerExpires))
	return f
}

func (f *fsm) eventWrapper(e event) func() {
	return func() {
		f.event(ConnectRetryTimerExpires)
	}
}

type event int

// Administrative Events
// https://tools.ietf.org/html/rfc4271#section-8.1.2
const (
	_ = iota
	ManualStart
	ManualStop
	AutomaticStart
	ManualStartWithPassiveTCPEstablishment
	AutomaticStartWithPassiveTCPEstablishment
	AutomaticStartWithDampPeerOscillations
	AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment
	AutomaticStop
	ConnectRetryTimerExpires
	HoldTimerExpires
	KeepaliveTimerExpires
	DelayOpenTimerExpires
	IdleHoldTimerExpires
	TCPConnectionValid
	TCPCRInvalid
	TCPCRAcked
	TCPConnectionConfirmed
	TCPConnectionFails
	BGPOpen
	BGPOpenWithDelayOpenTimerRunning
	BGPHeaderErr
	BGPOpenMsgErr
	OpenCollisionDump
	NotifMsgVerErr
	NotifMsg
	KeepAliveMsg
	UpdateMsg
	UpdateMsgErr
)

var eventLookup = map[event]string{
	ManualStart:                                                      "ManualStart",
	ManualStop:                                                       "ManualStop",
	ManualStartWithPassiveTCPEstablishment:                           "ManualStart_with_PassiveTcpEstablishment",
	AutomaticStartWithPassiveTCPEstablishment:                        "AutomaticStart_with_PassiveTcpEstablishment",
	AutomaticStartWithDampPeerOscillations:                           "AutomaticStart_with_DampPeerOscillations",
	AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment: "AutomaticStart_with_DampPeerOscillations_and_PassiveTcpEstablishment",
	AutomaticStop:            "AutomaticStop",
	ConnectRetryTimerExpires: "ConnectRetryTimer_Expires",
	HoldTimerExpires:         "HoldTimer_Expires",
	KeepaliveTimerExpires:    "KeepaliveTimer_Expires",
	DelayOpenTimerExpires:    "DelayOpenTimer_Expires",
	IdleHoldTimerExpires:     "IdleHoldTimer_Expires",
	TCPConnectionValid:       "TcpConnection_Valid",
	TCPCRAcked:               "Tcp_CR_Acked",
	TCPConnectionConfirmed:   "TcpConnectionConfirmed",
	TCPConnectionFails:       "TcpConnectionFails",
	BGPOpen:                  "BGPOpen",
	BGPOpenWithDelayOpenTimerRunning: "BGPOpen with DelayOpenTimer running",
	BGPHeaderErr:                     "BGPHeaderErr",
	BGPOpenMsgErr:                    "BGPOpenMsgErr",
	OpenCollisionDump:                "OpenCollisionDump",
	NotifMsgVerErr:                   "NotifMsgVerErr",
	NotifMsg:                         "NotifMsg",
	KeepAliveMsg:                     "KeepAliveMsg",
	UpdateMsg:                        "UpdateMsg",
	UpdateMsgErr:                     "UpdateMsgErr",
}

// String implements string.Stringer
func (e event) String() string {
	return eventLookup[e]
}

func (f *fsm) event(e event) {
	log.Println("routing event", e, "to state", f.state)
	switch f.state {
	case idle:
		f.idle(e)
	case connect:
		f.connect(e)
	case active:
		f.active(e)
	case openSent:
		f.openSent(e)
	case openConfirm:
		f.openConfirm(e)
	case established:
		f.established(e)
	}
}

func (f *fsm) transition(s state) {
	log.Println("Transitioning from", f.state, "to", s)
	f.state = s
}

// Handle ManualStart and AutomaticStart in the idle state
func (f *fsm) start() {
	f.peer.initializeResources()
	f.connectRetryCounter.Reset()
	f.connectRetryTimer.Reset(defaultConnectRetryTime)
	// TODO: initiates a TCP connection to the other BGP peer,
	f.transition(connect)
}

func (f *fsm) stop() {
	writeMessage(f.peer.conn, notification, newNotification(newBGPError(cease, 0, "manual stop")))
	f.connectRetryTimer.Stop()
	f.peer.releaseResources()
	f.peer.conn.Close()
	f.connectRetryCounter.Reset()
	f.transition(idle)
}

func (f *fsm) tcpConnect() {
	// TODO: Implement when adding the delayOpen option
	if f.delayOpen {
		f.connectRetryTimer.Stop()
		//   - sets the DelayOpenTimer to the initial value, and
		return
	}
	f.connectRetryTimer.Stop()
	f.peer.initializeResources()
	log.Println("Sending OPEN message")
	writeMessage(f.peer.conn, open, newOpen(f.peer))
	// 	TODO: sets the HoldTimer to a large value (4 min recommended)
	f.transition(openSent)
}

func (f *fsm) ignore(e event) {
	log.Println("%s state ignoring %s event", f.state, e)
}

func (f *fsm) fsmErrorToIdle() {
	writeMessage(f.peer.conn, notification, newNotification(newBGPError(fsmError, 0, "invalid mesage")))
	f.connectRetryTimer.Stop()
	f.peer.releaseResources()
	f.peer.conn.Close()
	f.connectRetryCounter.Increment()
	// - (optionally) performs peer oscillation damping if the
	// DampPeerOscillations attribute is set to TRUE, and
	f.transition(idle)
}

// In this state, BGP FSM refuses all incoming BGP connections for
// this peer.  No resources are allocated to the peer.
func (f *fsm) idle(e event) {
	switch e {
	case ManualStart:
		f.start()
	case AutomaticStart:
		f.start()
	//TODO: case ManualStartWithPassiveTCPEstablishment:
	//TODO: case AutomaticStartWithPassiveTCPEstablishment:
	//TODO: case AutomaticStartWithDampPeerOscillations:
	//TODO: case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	//TODO: case IdleHoldTimerExpires:
	default:
		log.Println("Ignoring event")
	}
}

func (f *fsm) connect(e event) {
	switch e {
	case ManualStart:
		f.ignore(e)
	case ManualStop:
		f.peer.conn.Close()
		f.peer.releaseResources()
		f.connectRetryCounter.Reset()
		f.connectRetryTimer.Stop()
		f.transition(idle)
	case AutomaticStart:
		f.ignore(e)
	case ManualStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillations:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
		f.ignore(e)
	case ConnectRetryTimerExpires:
		f.peer.conn.Close()
		f.connectRetryTimer.Reset(defaultConnectRetryTime)
		// TODO: stops the DelayOpenTimer and resets the timer to zero,
		// TODO: initiates a TCP connection to the other BGP peer,
	//TODO: case DelayOpenTimerExpires:
	//TODO: case TCPConnectionValid:
	//TODO: case TCPCRInvalid:
	case TCPCRAcked:
		f.tcpConnect()
	case TCPConnectionConfirmed:
		f.tcpConnect()
	case TCPConnectionFails:
	//TODO: case BGPOpenWithDelayOpenTimerRunning:
	case BGPHeaderErr:
		if f.sendNOTIFICATIONwithoutOPEN {
			// TODO: How do we get the actual error here? Doesn't get passed in with the event
			writeMessage(f.peer.conn, notification, newNotification(newBGPError(messageHeaderError, 0, "")))
		}
		f.connectRetryTimer.Stop()
		f.peer.releaseResources()
		f.peer.conn.Close()
		f.connectRetryCounter.Increment()
		// TODO: (optionally) performs peer oscillation damping if the
		// DampPeerOscillations attribute is set to TRUE, and
		f.transition(idle)
	case BGPOpenMsgErr:
	case NotifMsgVerErr:
		f.connectRetryTimer.Stop()
		// TODO: stops and resets the DelayOpenTimer (sets to zero),
		f.peer.releaseResources()
		f.peer.conn.Close()
		f.transition(idle)
	default:
		log.Println("Default handling of event")
		f.connectRetryTimer.Stop()
		// TODO: if the DelayOpenTimer is running, stops and resets the
		// DelayOpenTimer (sets to zero),
		f.peer.releaseResources()
		f.peer.conn.Close()
		f.connectRetryCounter.Increment()
		// TODO: performs peer oscillation damping if the DampPeerOscillations
		// attribute is set to True, and
		f.transition(idle)
	}
}

func (f *fsm) active(e event) {
	switch e {
	case ManualStart:
		f.ignore(e)
	case ManualStop:
	case AutomaticStart:
		f.ignore(e)
	case ManualStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillations:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
		f.ignore(e)
	case ConnectRetryTimerExpires:
	//TODO: case DelayOpenTimerExpires:
	//TODO: case TCPConnectionValid:
	//TODO: case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	//TODO: case BGPOpenWithDelayOpenTimerRunning:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	case NotifMsgVerErr:
	case NotifMsg:
	case KeepAliveMsg:
	case UpdateMsg:
	case UpdateMsgErr:
	default:
		log.Println("Default handling of event")
	}
}

func (f *fsm) openSent(e event) {
	switch e {
	case ManualStart:
		f.ignore(e)
	case ManualStop:
		f.stop()
	case AutomaticStart:
		f.ignore(e)
	case ManualStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillations:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStop:
	case HoldTimerExpires:
		writeMessage(f.peer.conn, notification, newNotification(newBGPError(holdTimerExpiredError, 0, "hold timer expired")))
		f.connectRetryTimer.Stop()
		f.peer.releaseResources()
		f.peer.conn.Close()
		f.connectRetryCounter.Increment()
		// TODO: (optionally) performs peer oscillation damping if the
		//   DampPeerOscillations attribute is set to TRUE, and
		f.transition(idle)
	case TCPConnectionValid:
		// TODO: Collision handling!
	case TCPCRInvalid:
		f.ignore(e)
	case TCPCRAcked:
		// TODO: Collision handling!
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	case BGPOpen:
		f.connectRetryTimer.Stop()
		writeMessage(f.peer.conn, keepalive, newKeepalive())
		if f.holdTime != 0 {
			// 	TODO: sets a KeepaliveTimer (via the text below)
			// 	TODO: sets the HoldTimer according to the negotiated value (see
			// 	Section 4.2),
		}
		f.transition(openConfirm)
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	//TODO: case OpenCollisionDump:
	case NotifMsgVerErr:
	default:
		f.fsmErrorToIdle()
	}
}

func (f *fsm) openConfirm(e event) {
	switch e {
	case ManualStart:
		f.ignore(e)
	case ManualStop:
		f.stop()
	case AutomaticStart:
		f.ignore(e)
	case ManualStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillations:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStop:
	case HoldTimerExpires:
	case KeepaliveTimerExpires:
	//TODO: case TCPConnectionValid:
	//TODO: case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	//TODO: case OpenCollisionDump:
	case NotifMsgVerErr:
	case NotifMsg:
		f.connectRetryTimer.Stop()
		f.peer.releaseResources()
		f.peer.conn.Close()
		f.transition(idle)
	case KeepAliveMsg:
		//TODO: restarts the HoldTimer and
		f.transition(established)
	default:
		f.fsmErrorToIdle()
	}
}

func (f *fsm) established(e event) {
	switch e {
	case ManualStart:
		f.ignore(e)
	case ManualStop:
	case AutomaticStart:
		f.ignore(e)
	case ManualStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithPassiveTCPEstablishment:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillations:
		f.ignore(e)
	case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
		f.ignore(e)
		//TODO: case AutomaticStop:
	case HoldTimerExpires:
	case KeepaliveTimerExpires:
	//TODO: case TCPConnectionValid:
	//TODO: case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	case BGPOpen:
	//TODO: case OpenCollisionDump:
	case NotifMsgVerErr:
	case NotifMsg:
	case KeepAliveMsg:
	case UpdateMsg:
	case UpdateMsgErr:
	default:
		// TODO: deletes all routes associated with this connection,
		// can this just be done as part of releasing resources? RFC says
		// to do this after writing notification and before the rest.
		f.fsmErrorToIdle()
	}
}
