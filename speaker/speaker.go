package speaker

import (
	"net"
)

// Speaker is a router that speaks BGP
type Speaker struct {
	myAS int16

	peers []Peer
}

// New creates a new router speaking BGP
func New(asn int16) *Speaker {
	return &Speaker{asn, []Peer{}}
}

// Peer is a remote BGP speaker
type Peer struct {
	asn     int32
	ip      net.IP
	enabled bool
	in      Policer
	out     Policer
}

// Enable starts this peer
func (p *Peer) Enable() {
	p.enabled = true
}

// Disable stops this peer
func (p *Peer) Disable() {
	p.enabled = false
}

type PeerOption func(*Peer) error

// Peer adds a new peer to this speaker
func (s *Speaker) Peer(asn int32, ip string, opts ...PeerOption) *Peer {
	peer := &Peer{
		asn: asn,
		ip:  net.ParseIP(ip),
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
	return peer
}

// Remove deletes a peer from this speaker
func (s *Speaker) Remove(asn int32, ip string) {
}

// The Policer interface is implemented by clients to apply policy
// to an individual NLRI. Returning false indicates the NLRI must be
// denied from advertisement or injection into a RIB. Policies modify
// the NLRI in place.
type Policer interface {
	Apply(*NLRI) bool
}

type DefaultPolicy struct{}

func (d DefaultPolicy) Apply(n *NLRI) bool {
	return false
}

func PolicyInOption(policy Policer) PeerOption {
	return func(p *Peer) error {
		p.in = policy
		return nil
	}
}

func PolicyOutOption(policy Policer) PeerOption {
	return func(p *Peer) error {
		p.out = policy
		return nil
	}
}

// Other possible per-peer options
// Passive - Do not initiate connections to the peer
// Timers - Modify default timer values
// BestPathSelecter - Custom best path selection

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

type NLRI struct {
	prefix net.IPNet
}

func (n *NLRI) String() string {
	return n.prefix.String()
}
