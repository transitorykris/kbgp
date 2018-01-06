package jbgp

import "net"

type peer struct {
	as      asn
	ip      net.IP
	passive bool // Do not initiate connections
	conn    net.Conn
	fsm     *fsm
}

func newPeer(as asn, ip net.IP) *peer {
	p := &peer{
		as: as,
		ip: ip,
	}
	p.fsm = newFSM(p)
	return &peer{as: as, ip: ip}
}

func (p *peer) handleConnection(conn net.Conn, open openMsg) {
	if p.conn != nil {
		// We have a connection already! Collision detection time
	}
	p.conn = conn
	if err := p.validateOpen(open); err != nil {
		p.fsm.event(BGPOpenMsgErr)
		writeMessage(p.conn, notification, newNotification(err))
		conn.Close()
		return
	}
	p.fsm.event(BGPOpen)
}

func (p *peer) validateOpen(o openMsg) error {
	if p.fsm.state == idle {
		return newBGPError(0, 0, "peer is idle")
	}
	return nil
}
