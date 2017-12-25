package speaker

import (
	"fmt"
	"net"
)

const bgpPort = 179

// Speaker is a router that speaks BGP
type Speaker struct {
	asn   int16
	peers []Peer

	listener net.Listener
	dialer   net.Dialer
}

// New creates a new router speaking BGP
func New(asn int16) *Speaker {
	s := &Speaker{
		asn:   asn,
		peers: []Peer{},
	}
	// TODO: What is a sane way to restrict which IPs we listen on?
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", bgpPort))
	if err != nil {
		panic(err)
	}
	s.listener = l
	return s
}

// Peer is a remote BGP speaker
type Peer struct {
	asn     int32
	ip      net.IP
	enabled bool
	in      Policer
	out     Policer

	best BestPathSelecter
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
	if peer.best == nil {
		peer.best = &DefaultBestPathSelection{}
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

type BestPathSelecter interface {
	// Note: probably return a list of nlri for multipath?
	Compare(nlris ...NLRI) nlri
}

type DefaultBestPathSelection struct{}

func (d *DefaultBestPathSelection) Compare(nlris ...NLRI) nlri {
	// TODO: Implement me
	return NLRI{}
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

type NLRI struct {
	prefix net.IPNet
}

func (n *NLRI) String() string {
	return n.prefix.String()
}
