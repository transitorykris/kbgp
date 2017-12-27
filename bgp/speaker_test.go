package kbgp

import (
	"net"
	"testing"
)

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

func TestStart(t *testing.T) {
	s := New(uint16(1234), uint32(4567))
	s.Start()
	// Not much to check but hope we don't panic
}

func TestStop(t *testing.T) {
	s := New(uint16(1234), uint32(4567))
	s.Stop()
	// Not much to check but hope we don't panic
}

func TestAddPeer(t *testing.T) {
	s := New(uint16(1234), uint32(4567))
	err := s.AddPeer(1, net.ParseIP("1.2.3.4"))
	if err != nil {
		t.Errorf("Expected AddPeer to return nil but got %s", err.Error())
	}
}

func TestRemovePeer(t *testing.T) {
	s := New(uint16(1234), uint32(4567))
	err := s.RemovePeer()
	if err != nil {
		t.Errorf("Expected RemovePeer to return nil but got %s", err.Error())
	}
}
