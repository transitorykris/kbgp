package main

import (
	"log"

	"github.com/transitorykris/kbgp"
)

func main() {
	log.Println("Starting kBGP")

	router, err := kbgp.New(1234, "1.2.3.4")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Configured speaker AS%s/%s", router.MyAS(), router.BGPIdentifier())

	log.Println("Exiting kBGP")
}
