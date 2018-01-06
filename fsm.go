package jbgp

import "log"

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
	state state

	// reference back to our owner
	peer *Peer
}

func newFSM(p *Peer) *fsm {
	return &fsm{peer: p}
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

func (f *fsm) idle(e event) {
	switch e {
	case BGPOpen:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	default:
		log.Println("Ignoring event")
	}
}

func (f *fsm) connect(e event) {
	switch e {
	case BGPOpen:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	default:
		log.Println("Ignoring event")
	}
}

func (f *fsm) active(e event) {
	switch e {
	case BGPOpen:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	default:
		log.Println("Ignoring event")
	}
}

func (f *fsm) openSent(e event) {
	switch e {
	case BGPOpen:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	default:
		log.Println("Ignoring event")
	}
}

func (f *fsm) openConfirm(e event) {
	switch e {
	case BGPOpen:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	default:
		log.Println("Ignoring event")
	}
}

func (f *fsm) established(e event) {
	switch e {
	case BGPOpen:
	case BGPHeaderErr:
	case BGPOpenMsgErr:
	default:
		log.Println("Ignoring event")
	}
}
