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

func (f *fsm) event(e event) {
	log.Println("routing event", e, "to state", f.state)
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
