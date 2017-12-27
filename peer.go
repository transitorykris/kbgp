package kbgp

import (
	"fmt"
	"net"
)

type peer struct {
	as       uint16
	remoteIP net.IP
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
		as:       uint16(peerAS),
		remoteIP: ip,
	}
	return p, nil
}
