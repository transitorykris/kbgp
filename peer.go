package kbgp

import (
	"fmt"
	"log"
	"net"
	"time"
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
	p.fsm.event(TCPConnectionConfirmed)
	if err := p.validateOpen(open); err != nil {
		log.Println("failed to validate open message", err)
		p.fsm.event(BGPOpenMsgErr)
		writeMessage(p.conn, notification, newNotification(err))
		conn.Close()
		return
	}
	// TODO: should have a configured hold time, but hardcoding default for now
	offeredHoldTime := time.Duration(open.holdTime) * time.Second
	if offeredHoldTime < defaultHoldTime {
		p.fsm.holdTime = offeredHoldTime
	}
	p.fsm.keepaliveTime = p.fsm.holdTime / 3
	p.fsm.event(BGPOpen)
	// Go into our inbound message processing loop
	p.processInbound()
}

func (p *Peer) processInbound() {
	for {
		h, body, err := readHeader(p.conn)
		if err != nil {
			// Message header notification
		}
		switch h.msgType {
		case open:
			log.Println("Received an open")
			_, err := readOpen(body)
			if err != nil {
				// Open message notification
			}
			//TODO: Implement me
		case update:
			log.Println("Received an update")
			//TODO: Implement me
		case notification:
			log.Println("Received a notification")
			_, err := readNotification(body)
			if err != nil {
				//TODO: Handle me
			}
			//TODO: We're done, shut it down
		case keepalive:
			log.Println("Received a keepalive")
			//TODO: Implement me
		}
	}
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
	if o.version != version {
		// TODO: this should be a 2-octet unsigned int
		return newBGPError(openMessageError, unsupportedVersionNumber, "4")
	}
	// TODO:
	// If the version number in the Version field of the received OPEN
	// message is not supported, then the Error Subcode MUST be set to
	// Unsupported Version Number.  The Data field is a 2-octet unsigned
	// integer, which indicates the largest, locally-supported version
	// number less than the version the remote BGP peer bid (as indicated in
	// the received OPEN message), or if the smallest, locally-supported
	// version number is greater than the version the remote BGP peer bid,
	// then the smallest, locally-supported version number.

	if o.holdTime == 1 || o.holdTime == 2 {
		return newBGPError(openMessageError, unacceptableHoldTime,
			"hold time must be 0 or greater than 2")
	}
	if !o.bgpIdentifier.valid() {
		return newBGPError(openMessageError, badBGPIdentifier,
			"BGP identifier must be a unicast IP")
	}
	if p.fsm.state == idle {
		return newBGPError(0, 0, "peer is idle")
	}

	// TODO:
	// If one of the Optional Parameters in the OPEN message is not
	// recognized, then the Error Subcode MUST be set to Unsupported
	// Optional Parameters.
	// If one of the Optional Parameters in the OPEN message is recognized,
	// but is malformed, then the Error Subcode MUST be set to 0
	// (Unspecific).

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

// Returns true if the peer is iBGP
func (p *Peer) internal() bool {
	if p.remoteAS == p.myAS {
		return true
	}
	return false
}

// Returns true if the peer is eBGP
func (p *Peer) external() bool {
	return !p.internal()
}
