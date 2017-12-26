package speaker

import (
	"net"
	"time"
)

// Peer is a remote BGP speaker
type Peer struct {
	asn     int16
	ip      net.IP
	enabled bool
	in      Policer
	out     Policer

	best BestPathSelecter

	// Peer timers
	holdTime          time.Duration
	keepAliveInterval time.Duration
	connectRetryTime  time.Duration
	initialHoldTime   time.Duration

	// These may be per-speaker or per-peer
	delayOpenTime time.Duration
	idleHoldTime  time.Duration
}

// The Policer interface is implemented by clients to apply policy
// to an individual NLRI. Returning false indicates the NLRI must be
// denied from advertisement or injection into a RIB. Policies modify
// the NLRI in place.
type Policer interface {
	Apply(*NLRI) bool
}

// The BestPathSelecter interface is implemented by clients to create a
// custom best path selection procedure.
type BestPathSelecter interface {
	// Note: probably return a list of nlri for multipath?
	Compare(nlris ...NLRI) NLRI
}

// Enable starts this peer
func (p *Peer) Enable() {
	p.enabled = true
}

// Disable stops this peer
func (p *Peer) Disable() {
	p.enabled = false
}
