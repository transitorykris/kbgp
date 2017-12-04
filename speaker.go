package kbgp

import (
	"fmt"
	"log"
	"net"
)

// Speaker implements BGP4
type Speaker struct {
	version       int
	myAS          uint16
	bgpIdentifier uint32
	locRIB        *locRIB
	fsm           []*fsm
}

// New creates a new BGP speaker
func New(myAS uint16, bgpIdentifier uint32) *Speaker {
	s := &Speaker{
		version:       version,
		myAS:          myAS,
		bgpIdentifier: bgpIdentifier,
		locRIB:        newLocRIB(),
	}
	return s
}

// Start sends an automatic start to all FSMs
func (s *Speaker) Start() {
	for _, f := range s.fsm {
		if f.allowAutomaticStart {
			f.sendEvent(manualStart)()
		}
	}
	// We'll have just one listener while hacking through this
	go s.listener()
}

// Stop sends an automatic stop to all FSMs
func (s *Speaker) Stop() {
	for _, f := range s.fsm {
		if f.allowAutomaticStop {
			f.sendEvent(manualStop)()
		}
	}
}

// AddPeer configures a new BGP neighbor. Returns nil if successful.
func (s *Speaker) AddPeer(remoteAS int, remoteIP net.IP) error {
	s.fsm = append(s.fsm, newFSM(remoteAS, remoteIP))
	return nil
}

// RemovePeer removes the BGP neighbor from the speaker. Returns nil if successful.
func (s *Speaker) RemovePeer() error {
	return nil
}

// listen handles incoming TCP connections and attempts to match them to
// a FSM or reject them if no such FSM exists or if they are in a state that
// forbids new connections.
func (s *Speaker) listener() {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
		// TODO: Figure out how to handle errors here
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
			// TODO: Figure out how to handle errors here
		}
		log.Printf("Connection from %s to %s", conn.RemoteAddr(), conn.LocalAddr())
		// TODO:
		// - Figure out what peer this is
		var f *fsm
		for _, v := range s.fsm {
			// Note: this probably won't work, RemoteAddr()'s format could have things
			// like the port attached to it. There may be a better way..
			if v.peer.remoteIP.String() == conn.RemoteAddr().String() {
				f = v
				break
			}
		}

		// TODO: Allow binding listeners to specific IP addresses
		// If the remote address does not match the expected remote address
		// we would reject the connection. In this case we would check if
		// TrackTcpState is true and if so send a Tcp_CR_Invalid message to
		// the FSM.

		// If we found no associated FSM, close the connection.
		if f == nil {
			log.Println("No matching peer found, closing TCP connection")
			// Tell the state machine about the failure?
			// Is ther any error to send here before closing?
			conn.Close()
			continue // Accept the next connection
		}
		// - Check if the FSM is okay to take a new connection

		// - Set s.fsm.peer.conn to this conn
		f.peer.conn = &conn

		// Event 14: TcpConnection_Valid

		// 		 Definition: Event indicating the local system reception of a
		// 					 TCP connection request with a valid source IP
		// 					 address, TCP port, destination IP address, and TCP
		// 					 Port.  The definition of invalid source and invalid
		// 					 destination IP address is determined by the
		// 					 implementation.

		// 					 BGP's destination port SHOULD be port 179, as
		// 					 defined by IANA.

		// 					 TCP connection request is denoted by the local
		// 					 system receiving a TCP SYN.

		// 		 Status:     Optional

		// 		 Optional
		// 		 Attribute
		// 		 Status:     1) The TrackTcpState attribute SHOULD be set to
		// 						TRUE if this event occurs.
		if f.trackTCPState {
			f.sendEvent(tcpConnectionValid)
		}

		// Event 17: TcpConnectionConfirmed

		// 		 Definition: Event indicating that the local system has received
		// 					 a confirmation that the TCP connection has been
		// 					 established by the remote site.

		// 					 The remote peer's TCP engine sent a TCP SYN.  The
		// 					 local peer's TCP engine sent a SYN, ACK message and
		// 					 now has received a final ACK.

		// 		 Status:     Mandatory
		f.sendEvent(tcpConnectionConfirmed)
	}
}

// dial attempts to form a TCP connection with the peer
func (s *Speaker) dial(f *fsm) {
	conn, err := net.Dial("tcp4", f.peer.remoteIP.String())
	if err != nil {
		log.Fatal(err)
		// TODO: Figure out how to handle errors here
	}

	// Event 16: Tcp_CR_Acked

	// 		 Definition: Event indicating the local system's request to
	// 					 establish a TCP connection to the remote peer.

	// 					 The local system's TCP connection sent a TCP SYN,
	// 					 received a TCP SYN/ACK message, and sent a TCP ACK.

	// 		 Status:     Mandatory
	f.peer.conn = &conn
	f.sendEvent(tcpCRAcked)
}
