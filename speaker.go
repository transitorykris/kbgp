package kbgp

import (
	"log"
	"net"
	"strings"
)

// Speaker is a BGP speaking router
type Speaker struct {
	as    asn
	addr  string
	peers []*Peer
}

// NewSpeaker creates a new BGP speaking router
func NewSpeaker(as asn, addr string) *Speaker {
	return &Speaker{as: as, addr: addr}
}

// Start the BGP speaker
func (s *Speaker) Start() {
	ln, err := net.Listen("tcp", ":179")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Speaker) handleConnection(conn net.Conn) {
	log.Println("handling connection from", conn.RemoteAddr())
	header, body, err := readHeader(conn)
	if err != nil {
		log.Println("header error")
		writeMessage(conn, notification, newNotification(err))
		conn.Close()
		return
	}
	if header.msgType != open {
		log.Println("expected an open message")
		writeMessage(conn, notification, newNotification(err))
		conn.Close()
		return
	}
	open, err := readOpen(body)
	if err != nil {
		log.Println("bad open message")
		writeMessage(conn, notification, newNotification(err))
		conn.Close()
		return
	}
	for _, p := range s.peers {
		if p.remoteAS == open.as && p.remoteIP.Equal(addrToIP(conn.RemoteAddr())) {
			log.Println("found a matching peer")
			go p.handleConnection(conn, open)
			return
		}
	}
	log.Println("no matching peer found for", open.as, conn.RemoteAddr())
	writeMessage(conn, notification, newNotification(newBGPError(openMessageError, badPeerAS, "")))
	conn.Close()
}

func addrToIP(addr net.Addr) net.IP {
	return net.ParseIP(strings.Split(addr.String(), ":")[0])
}

// Peer adds a BGP neighbor to the speaker
func (s *Speaker) Peer(p *Peer) {
	log.Println("adding peer to speaker", p)
	// Let this peer know who we are
	p.myAS = s.as
	s.peers = append(s.peers, p)
}
