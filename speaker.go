package jbgp

import (
	"log"
	"net"
)

type speaker struct {
	addr  string
	peers []*peer
}

func newSpeaker(addr string) *speaker {
	return &speaker{addr: addr}
}

func (s *speaker) Start() {
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

func (s *speaker) handleConnection(conn net.Conn) {
	header, err := readHeader(conn)
	if err != nil {
		writeMessage(conn, notification, newNotification(err))
		conn.Close()
		return
	}
	if header.msgType != open {
		writeMessage(conn, notification, newNotification(err))
		conn.Close()
		return
	}
	open, err := readOpen(conn)
	if err != nil {
		writeMessage(conn, notification, newNotification(err))
		conn.Close()
		return
	}
	for _, p := range s.peers {
		if p.as == open.as && p.ip.String() == conn.RemoteAddr().String() {
			go p.handleConnection(conn, open)
			return
		}
	}
	// No matching peer was found
	writeMessage(conn, notification, newNotification(bgpError{0, 0, ""}))
	conn.Close()
}

func (s *speaker) peer(p *peer) {
	s.peers = append(s.peers, p)
}
