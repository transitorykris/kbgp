package main

import (
	"log"
	"time"

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
	if err := bgp.AddPeer(); err != nil {
		log.Fatal(err)
	}
	bgp.Start()

	time.Sleep(5 * time.Second)

	bgp.Stop()
	log.Println("Exiting")
}
