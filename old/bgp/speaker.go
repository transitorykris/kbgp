package bgp

// New creates a new BGP speaking router
func New(asn ASN, id Identifier, locRIB RIB) *Speaker {
	s := &Speaker{
		ASN:    asn,
		ID:     id,
		LocRIB: locRIB,
	}
	return s
}

// Peer adds a new peer to the speaker
func (s *Speaker) Peer(p *Peer) {

}

// Remove removes a peer from the speaker
func (s *Speaker) Remove(p *Peer) {

}
