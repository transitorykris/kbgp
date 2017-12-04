package main

import (
	"log"
	"net"

	"github.com/transitorykris/kbgp"
)

func main() {
	log.Println("Starting kBGP")

	// Config
	myAS := uint16(1234)
	id, err := kbgp.FindBGPIdentifier()
	if err != nil {
		panic(err)
	}

	// Start our router
	log.Printf("AS: %d ID: %s", myAS, kbgp.Uint32ToIP(id).String())
	bgp := kbgp.New(myAS, id)

	remoteAS := uint16(1001)
	remoteIP := net.ParseIP("127.0.0.1")
	if err := bgp.AddPeer(remoteAS, remoteIP); err != nil {
		log.Fatal(err)
	}

	bgp.Start()

	select {}

	//time.Sleep(5 * time.Second)

	//bgp.Stop()
	//log.Println("Exiting")
}
