package kbgp

import (
	"fmt"
	"log"
	"net"
)

type peer struct {
	remoteAS uint16
	remoteIP net.IP
	enabled  bool

	// the FSM will set this flag so the speaker will know to route
	// incoming connections to this peer
	listening bool

	// If passive is set, then we will not attempt to dial out
	passive bool

	fsm *fsm

	// The one true connection
	conn net.Conn

	// Temporary holding place for incoming and outgoing connections
	incoming net.Conn
	outgoing net.Conn
}

func newPeer(peerAS int, remoteIP string) (*peer, error) {
	if !validAS(peerAS) {
		return nil, fmt.Errorf("invalid autonomous system number %d", peerAS)
	}
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return nil, fmt.Errorf("remote IP address %s is invalid", remoteIP)
	}
	p := &peer{
		remoteAS: uint16(peerAS),
		remoteIP: ip,
	}
	p.fsm = newFSM(p)
	return p, nil
}

func (p *peer) connect(conn net.Conn) {
	if !p.enabled || !p.listening {
		conn.Close()
	}
	// TODO: Implement me
}

func (p *peer) enable() {
	if p.enabled {
		return
	}
	p.enabled = true
}

func (p *peer) disable() {
	if !p.enabled {
		return
	}
	p.enabled = false
}

func (p *peer) listen() {
	if p.listening {
		return
	}
	p.listening = true
}

func (p *peer) dialLoop() {
	if !p.enabled || p.passive {
		return
	}
	log.Println("Dialing", p.remoteIP, Port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.remoteIP, Port))
	if err != nil {
		// TODO: What's the correct interval here?
		log.Println(err)
		// TODO: Do we say something to the FSM?
		return
	}
	p.connect(conn)
}
