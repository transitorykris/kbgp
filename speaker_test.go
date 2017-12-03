package kbgp

import "testing"

func TestNew(t *testing.T) {
	as := uint16(1234)
	id := uint32(4567)
	s := New(as, id)
	if s.myAS != as {
		t.Errorf("Expected AS to be %d but got %d", as, s.myAS)
	}
	if s.bgpIdentifier != id {
		t.Errorf("Expected BGP identifier to be %d but got %d", id, s.bgpIdentifier)
	}
	if s.locRIB == nil {
		t.Errorf("Did not expect locRIB to be nil")
	}
}
