package speaker

import (
	"net"
	"time"

	"github.com/transitorykris/kbgp"
)

const bgpPort = 179

// Speaker is a router that speaks BGP
type Speaker struct {
	asn   int16
	peers []Peer

	listeners map[net.Listener]struct{}
	conns     map[net.Conn]struct{}

	// These may be per-speaker or per-peer
	delayOpenTime time.Duration
	idleHoldTime  time.Duration
}

// New creates a new router speaking BGP
func New(asn int16) *Speaker {
	s := &Speaker{
		asn:           asn,
		peers:         []Peer{},
		delayOpenTime: kbgp.DefaultDelayOpenTime,
		idleHoldTime:  0, // What should default be?
	}
	return s
}

// Remove deletes a peer from this speaker
func (s *Speaker) Remove(asn int32, ip string) {
}

// Announce the given prefix
func (s *Speaker) Announce(prefix string) error {
	_, _, err := net.ParseCIDR(prefix)
	return err
}

// Withdraw the given prefix
func (s *Speaker) Withdraw(prefix string) error {
	_, _, err := net.ParseCIDR(prefix)
	return err
}

// PeerOption is used to pass arbitrary options when creating a new peer
type PeerOption func(*Peer) error

// Peer adds a new peer to this speaker
func (s *Speaker) Peer(asn int16, ip string, opts ...PeerOption) *Peer {
	peer := &Peer{
		asn: asn,
		ip:  net.ParseIP(ip),

		// Peer timers
		holdTime:          kbgp.DefaultHoldTime,
		keepAliveInterval: kbgp.DefaultKeepaliveTime,
		connectRetryTime:  kbgp.DefaultConnectRetryTime,
		initialHoldTime:   0, // Default??

		// These may be per-speaker or per-peer
		// TODO: Allow these to be overridden for this peer
		delayOpenTime: s.delayOpenTime,
		idleHoldTime:  s.idleHoldTime,
	}
	for _, opt := range opts {
		opt(peer)
	}
	if peer.in == nil {
		peer.in = &DefaultPolicy{}
	}
	if peer.out == nil {
		peer.out = &DefaultPolicy{}
	}
	if peer.best == nil {
		peer.best = &DefaultBestPathSelection{}
	}
	return peer
}
