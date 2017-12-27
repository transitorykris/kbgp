package kbgp

import (
	"fmt"
	"log"
	"net"
	"time"
)

type peer struct {
	as       uint16
	remoteIP net.IP
	enabled  bool
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
	go p.dialLoop()
	return p, nil
}

func (p *peer) connect(conn net.Conn) {
	if !p.enabled {
		conn.Close()
	}
	// TODO: Implement me
}

func (p *peer) enable() {
	if p.enabled {
		return
	}
	p.enabled = true
}

func (p *peer) disable() {
	if !p.enabled {
		return
	}
	p.enabled = false
}

func (p *peer) dialLoop() {
	for {
		// TODO: Replace with a channel
		if !p.enabled {
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Dialing", p.remoteIP, Port)
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.remoteIP, Port))
		if err != nil {
			// TODO: What's the correct interval here?
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		p.connect(conn)
		break
	}
}
