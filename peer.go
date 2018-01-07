package jbgp

import (
	"fmt"
	"log"
	"net"
)

// Peer is a BGP neighbor
type Peer struct {
	myAS     asn
	remoteAS asn
	remoteIP net.IP
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

// String implements string.Stringer
func (p *Peer) String() string {
	return fmt.Sprintf("AS%d/%s", p.remoteAS, p.remoteIP)
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

// Up sends a ManualStart event to the FSM
func (p *Peer) Up() {
	p.fsm.event(ManualStart)
}

// Down sends a ManualStop event to the FSM
func (p *Peer) Down() {
	p.fsm.event(ManualStop)
}

func (p *Peer) close() {
	log.Println("Closing connection to", p)
	p.conn.Close()
}

func (p *Peer) validateOpen(o openMsg) error {
	if p.fsm.state == idle {
		return newBGPError(0, 0, "peer is idle")
	}
	return nil
}

// initializeResources initializes all BGP resources for this peer
func (p *Peer) initializeResources() {
	// TODO: implement me
}

// releaseResources releases all BGP resources held by this peer
func (p *Peer) releaseResources() {
	// TODO: Implement me
}
