package main

import (
	"log"
	"net"

	"github.com/transitorykris/kbgp"
)

func main() {
	log.Println("Creating a new speaker")
	speaker := kbgp.NewSpeaker(1234, ":179")

	log.Println("Adding a peer")
	myPeer := kbgp.NewPeer(1234, net.ParseIP("192.168.86.30"))
	speaker.Peer(myPeer)
	myPeer.Up()

	log.Println("Starting the speaker")
	speaker.Start()

	log.Println("Exiting  kbgp")
}
