package jbgp

import (
	"log"
	"net"
)

// Peer is a BGP neighbor
type Peer struct {
	myAS     asn
	remoteAS asn
	remoteIP net.IP
	passive  bool // Do not initiate connections
	conn     net.Conn
	fsm      *fsm
}

// NewPeer creates a new BGP neighbor
func NewPeer(as asn, ip net.IP) *Peer {
	p := &Peer{
		remoteAS: as,
		remoteIP: ip,
	}
	p.fsm = newFSM(p)
	return p
}

func (p *Peer) handleConnection(conn net.Conn, open openMsg) {
	log.Println("handling connection for", open)
	if p.conn != nil {
		log.Println("connection collision detected")
		// We have a connection already! Collision detection time
	}
	p.conn = conn
	if err := p.validateOpen(open); err != nil {
		log.Println("failed to validate open message", err)
		p.fsm.event(BGPOpenMsgErr)
		writeMessage(p.conn, notification, newNotification(err))
		conn.Close()
		return
	}
	p.fsm.event(BGPOpen)
}

func (p *Peer) validateOpen(o openMsg) error {
	if p.fsm.state == idle {
		return newBGPError(0, 0, "peer is idle")
	}
	return nil
}
