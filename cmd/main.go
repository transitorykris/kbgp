package main

import (
	"log"
	"net"

	"github.com/transitorykris/kbgp/speaker"
)

var myAS = int16(12345)

func main() {
	log.Println("Starting")

	router := speaker.New(myAS)

	peer := router.Peer(
		1111,
		"1.1.1.1",
		speaker.PolicyInOption(myPolicyIn{}),
		speaker.PolicyOutOption(myPolicyOut{}),
	)
	peer.Enable()

	//router.Announce("1.2.3.0/24")
	//router.Withdraw("1.2.3.0/24")

	// A peer with the default inbound and outbound policies
	_ = router.Peer(3333, "3.3.3.3")

	peer.Disable()

	log.Println("Exiting")
}

// Implements policy for our peer
type myPolicyIn struct{}

func (m myPolicyIn) Apply(nlri *speaker.NLRI) bool {
	return true
}

type myPolicyOut struct{}

func (m myPolicyOut) Apply(nlri *speaker.NLRI) bool {
	return true
}

type fancyPolicy struct {
	deny []net.IPNet
}

func (f *fancyPolicy) Apple(nlri *speaker.NLRI) bool {
	for _, n := range f.deny {
		if n.String() == nlri.String() {
			return false
		}
	}
	return true
}
