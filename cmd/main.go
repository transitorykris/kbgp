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

	if err = router.Peer(1111, "1.1.1.1"); err != nil {
		log.Println(err)
	}
	if err = router.Peer(2222, "2.2.2.2"); err != nil {
		log.Println(err)
	}
	if err = router.Peer(3333, "1.1.1.1"); err != nil {
		log.Println(err)
	}

	log.Println("Exiting kBGP")
}
