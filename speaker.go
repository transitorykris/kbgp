package kbgp

import "net"

// Speaker implements BGP4
type Speaker struct {
	version       int
	myAS          uint16
	bgpIdentifier uint32
	locRIB        *locRIB
	fsm           []*fsm
}

// New creates a new BGP speaker
func New(myAS uint16, bgpIdentifier uint32) *Speaker {
	s := &Speaker{
		version:       version,
		myAS:          myAS,
		bgpIdentifier: bgpIdentifier,
		locRIB:        newLocRIB(),
	}
	return s
}

// Start sends an automatic start to all FSMs
func (s *Speaker) Start() {
	for _, f := range s.fsm {
		if f.allowAutomaticStart {
			f.sendEvent(manualStart)()
		}
	}
}

// Stop sends an automatic stop to all FSMs
func (s *Speaker) Stop() {
	for _, f := range s.fsm {
		if f.allowAutomaticStop {
			f.sendEvent(manualStop)()
		}
	}
}

// AddPeer configures a new BGP neighbor. Returns nil if successful.
func (s *Speaker) AddPeer() error {
	s.fsm = append(s.fsm, newFSM())
	return nil
}

// RemovePeer removes the BGP neighbor from the speaker. Returns nil if successful.
func (s *Speaker) RemovePeer() error {
	return nil
}

// listen handles incoming TCP connections and attempts to match them to
// a FSM or reject them if no such FSM exists or if they are in a state that
// forbids new connections.
func (s *Speaker) listener() (*net.Conn, error) {
	ln, err := net.Listen("tcp4", "")
	if err != nil {
		return nil, err
	}
	conn, err := ln.Accept()
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

// dial attempts to form a TCP connection with the peer
func (s *Speaker) dial(fsm *fsm) (*net.Conn, error) {
	conn, err := net.Dial("tcp4", fsm.peer.remoteIP.String())
	if err != nil {
		return nil, err
	}
	return &conn, nil
}
