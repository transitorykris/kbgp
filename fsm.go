package jbgp

import (
	"log"
	"time"

	"github.com/transitorykris/jbgp/counter"
	"github.com/transitorykris/jbgp/timer"
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
	// delayOpen                          bool
	// delayOpenTime                      time.Duration
	// delayOpenTimer                     timer.Timer
	// idleHoldTime                       time.Duration
	// idleHoldTimer                      timer.Timer
	// passiveTcpEstablishment            bool
	// sendNOTIFICATIONwithoutOPEN        bool
	// trackTcpState                      bool

	// reference back to our owner
	peer *Peer
}

// https://tools.ietf.org/html/rfc4271#section-10
const defaultConnectRetryTime = 120 * time.Second

func newFSM(p *Peer) *fsm {
	f := &fsm{peer: p}
	log.Println("Creating the connectRetryTimer")
	f.connectRetryTimer = timer.New(defaultConnectRetryTime, f.eventWrapper(ConnectRetryTimerExpires))
	f.connectRetryTimer.Stop()
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

// In this state, BGP FSM refuses all incoming BGP connections for
// this peer.  No resources are allocated to the peer.
func (f *fsm) idle(e event) {
	switch e {
	case ManualStart:
		f.start()
	case AutomaticStart:
		f.start()
	//case ManualStartWithPassiveTCPEstablishment:
	//case AutomaticStartWithPassiveTCPEstablishment:
	//case AutomaticStartWithDampPeerOscillations:
	//case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	//case IdleHoldTimerExpires:
	default:
		log.Println("Ignoring event")
	}
}

// Handle ManualStart and AutomaticStart in the idle state
func (f *fsm) start() {
	// TODO: initializes all BGP resources for the peer connection,
	f.connectRetryCounter.Reset()
	f.connectRetryTimer.Reset(defaultConnectRetryTime)
	// TODO: initiates a TCP connection to the other BGP peer,
	f.transition(connect)
}

func (f *fsm) connect(e event) {
	switch e {
	case ManualStart:
		// ignore
	case ManualStop:
		f.peer.conn.Close()
		// TODO(?): releases all BGP resources,
		f.connectRetryCounter.Reset()
		f.connectRetryTimer.Stop()
		f.transition(idle)
	case AutomaticStart:
		// ignore
	// case ManualStartWithPassiveTCPEstablishment:
	// ignore
	// case AutomaticStartWithPassiveTCPEstablishment:
	// 	ignore
	// case AutomaticStartWithDampPeerOscillations:
	// 	ignore
	// case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	// 	ignore
	case ConnectRetryTimerExpires:
		log.Println("ConnectRetryTimer expired")
	//case DelayOpenTimerExpires:
	//case TCPConnectionValid:
	//case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	//case BGPOpenWithDelayOpenTimerRunning:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	case NotifMsgVerErr:
	default:
		log.Println("Default handling of event")
	}
}

func (f *fsm) active(e event) {
	switch e {
	case ManualStart:
		// ignore
	case ManualStop:
	case AutomaticStart:
		// ignore
	//case ManualStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithDampPeerOscillations:
	// ignore
	//case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	// ignore
	case ConnectRetryTimerExpires:
	//case DelayOpenTimerExpires:
	//case TCPConnectionValid:
	//case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	//case BGPOpenWithDelayOpenTimerRunning:
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
		// ignore
	case ManualStop:
	case AutomaticStart:
		// ignore
	//case ManualStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithDampPeerOscillations:
	// ignore
	//case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	// ignore
	//case AutomaticStop:
	case HoldTimerExpires:
	//case TCPConnectionValid:
	//case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	case BGPOpen:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	//case OpenCollisionDump:
	case NotifMsgVerErr:
	default:
		log.Println("Default handling of event")
	}
}

func (f *fsm) openConfirm(e event) {
	switch e {
	case ManualStart:
		// ignore
	case ManualStop:
	case AutomaticStart:
		// ignore
	//case ManualStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithDampPeerOscillations:
	// ignore
	//case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	// ignore
	//case AutomaticStop:
	case HoldTimerExpires:
	case KeepaliveTimerExpires:
	//case TCPConnectionValid:
	//case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	//case OpenCollisionDump:
	case NotifMsgVerErr:
	case NotifMsg:
	case KeepAliveMsg:
	default:
		log.Println("Default handling of event")
	}
}

func (f *fsm) established(e event) {
	switch e {
	case ManualStart:
		// ignore
	case ManualStop:
	case AutomaticStart:
		// ignore
	//case ManualStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithPassiveTCPEstablishment:
	// ignore
	//case AutomaticStartWithDampPeerOscillations:
	// ignore
	//case AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment:
	// ignore
	//case AutomaticStop:
	case HoldTimerExpires:
	case KeepaliveTimerExpires:
	//case TCPConnectionValid:
	//case TCPCRInvalid:
	case TCPCRAcked:
	case TCPConnectionConfirmed:
	case TCPConnectionFails:
	case BGPOpen:
	//case OpenCollisionDump:
	case NotifMsgVerErr:
	case NotifMsg:
	case KeepAliveMsg:
	case UpdateMsg:
	case UpdateMsgErr:
	default:
		log.Println("Ignoring event")
	}
}
