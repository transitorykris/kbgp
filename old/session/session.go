package session

import (
	"net"

	"github.com/transitorykris/kbgp/bgp"
)

// peer implements bgp.Peer
type session struct {
	as      bgp.ASN
	address net.IP
}

// New creates a new BGP neighbor
func New(asn bgp.ASN, address net.IP) bgp.Session {
	p := &session{
		as:      asn,
		address: address,
	}
	return p
}

// Connect implements bgp.Peer
func (p *session) Connect(conn net.Conn) {}

// Shutdown implements bgp.Peer
func (p *session) Shutdown() {}

// Announce implements bgp.Peer
func (p *session) Announce(route bgp.Route) {}

// Withdraw implements bgp.Peer
func (p *session) Withdraw(nlri bgp.NLRI) {}
