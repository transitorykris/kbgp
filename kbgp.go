package kbgp

import (
	"fmt"
	"net"
)

// Version is always 4 for BGP4
const Version = 4

// Speaker is a BGP router
type Speaker struct {
	// This 2-octet unsigned integer indicates the Autonomous System
	// number of the sender.
	myAS uint16

	// This 4-octet unsigned integer indicates the BGP Identifier of
	// the sender.  A given BGP speaker sets the value of its BGP
	// Identifier to an IP address that is assigned to that BGP
	// speaker.  The value of the BGP Identifier is determined upon
	// startup and is the same for every local interface and BGP peer.
	bgpIdentifier net.IP

	peers []*peer
}

// New creates a new BGP speaker
func New(myAS int, bgpIdentifier string) (*Speaker, error) {
	if !validAS(myAS) {
		return nil, fmt.Errorf("invalid autonomous system number %d", myAS)
	}
	ip := net.ParseIP(bgpIdentifier).To4()
	if ip == nil {
		return nil, fmt.Errorf("identifier %s is not a valid IPv4 address", bgpIdentifier)
	}
	s := &Speaker{
		myAS:          uint16(myAS),
		bgpIdentifier: ip,
	}
	return s, nil
}

func validAS(as int) bool {
	if 0 > as || as > 65535 {
		return false
	}
	return true
}

// MyAS returns a string representation of this speakers autonomous system number
func (s *Speaker) MyAS() string {
	return fmt.Sprintf("%d", s.myAS)
}

// BGPIdentifier returns a string representation of this speakers BGP identifier
func (s *Speaker) BGPIdentifier() string {
	return s.bgpIdentifier.String()
}

// Peer creates a new peer for this speaker
func (s *Speaker) Peer(peerAS int, remoteIP string) error {
	p, err := newPeer(peerAS, remoteIP)
	if err != nil {
		return err
	}
	for _, v := range s.peers {
		if v.remoteIP.Equal(net.ParseIP(remoteIP)) {
			return fmt.Errorf("Peer with remote IP %s already configured", remoteIP)
		}
	}
	s.peers = append(s.peers, p)
	return nil
}
