package main

import (
	"log"

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
	bgp.Start()

	log.Println("Exiting")
}
