package main

import (
	"log"
	"net"
	"time"

	"github.com/transitorykris/kbgp"
)

func main() {
	log.Println("Starting kBGP")

	listener, err := net.Listen("tcp", "0.0.0.0:8179")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	router, err := kbgp.New(1234, "1.2.3.4", listener)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Configured speaker AS%s/%s", router.MyAS(), router.BGPIdentifier())

	if err = router.Peer(1111, "1.1.1.1"); err != nil {
		log.Println(err)
	}

	if err = router.Peer(2222, "127.0.0.1"); err != nil {
		log.Println(err)
	}
	// TODO: Make this less ugly
	router.Enable(2222, "127.0.0.1")

	if err = router.Peer(3333, "1.1.1.1"); err != nil {
		log.Println(err)
	}

	go func() {
		if err := router.Speak(); err != nil {
			log.Println(err)
		}
	}()

	router.EnablePeer(1111, "1.1.1.1")

	time.Sleep(10 * time.Second)
	router.DisablePeer(1111, "1.1.1.1")

	select {}

	log.Println("Exiting kBGP")
}
