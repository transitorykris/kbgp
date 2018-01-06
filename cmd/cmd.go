package main

import (
	"log"
	"net"

	"github.com/transitorykris/jbgp"
)

func main() {
	log.Println("Creating a new speaker")
	speaker := jbgp.NewSpeaker(1234, ":179")

	log.Println("Adding a peer")
	speaker.Peer(jbgp.NewPeer(1234, net.ParseIP("1.2.3.4")))

	log.Println("Starting the speaker")
	speaker.Start()

	log.Println("Exiting jBGP")
}
