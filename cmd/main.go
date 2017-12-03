package main

import (
	"log"

	"github.com/transitorykris/kbgp"
)

func main() {
	log.Println("Starting kBGP")

	// Config
	myAS := uint16(1234)
	id := uint32(123456789)

	// Start our router
	bgp := kbgp.New(myAS, id)
	bgp.Start()

	log.Println("Exiting")
}
