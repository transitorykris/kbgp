package kbgp

// Speaker implements BGP4
type Speaker struct {
	version       int
	myAS          uint16
	bgpIdentifier uint32
	locRIB        *locRIB
	fsm           []*fsm
}

// New creates a new BGP speaker
func New(myAS uint16, bgpIdentifier uint32) *Speaker {
	s := &Speaker{
		version:       version,
		myAS:          myAS,
		bgpIdentifier: bgpIdentifier,
		locRIB:        newLocRIB(),
	}
	return s
}

// Start sends an automatic start to all FSMs
func (s *Speaker) Start() {
	for _, f := range s.fsm {
		if f.allowAutomaticStart {
			f.sendEvent(automaticStart)
		}
	}
}

// Stop sends an automatic stop to all FSMs
func (s *Speaker) Stop() {
	for _, f := range s.fsm {
		if f.allowAutomaticStop {
			f.sendEvent(automaticStop)
		}
	}
}
