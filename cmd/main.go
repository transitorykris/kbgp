package main

import (
	"log"
	"net"

	"github.com/transitorykris/kbgp/bgp"
	"github.com/transitorykris/kbgp/fsm"
	"github.com/transitorykris/kbgp/policy"
	"github.com/transitorykris/kbgp/rib"
	"github.com/transitorykris/kbgp/session"
)

const myAS = 1234
const id = 1234
const locRIB = rib.New()

func main() {
	log.Println("Starting kBGP")

	speaker := bgp.New(myAS, id, locRIB)
	peer1 := &Peer{
		Session: session.New(1111, net.ParseIP("1.1.1.1")),
		FSM:     fsm.New(),
	}
}

func NewPeer() bgp.Peer {
	p := &Peer{
		Session:   session.New(1111, net.ParseIP("1.1.1.1")),
		FSM:       fsm.New(),
		AdjRIBIn:  policy.DefaultEBGP,
		AdjRIBOut: policy.DefaultEBGP,
	}
	return p
}
