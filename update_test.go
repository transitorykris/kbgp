package kbgp

import (
	"fmt"
	"net"
	"testing"
)

func TestHandleUpdate(t *testing.T) {
	f := newFSM(1, net.ParseIP("1.2.3.4"))

	f.state = established
	u := &updateMessage{}
	_, err := f.handleUpdate(u)
	if err != nil {
		fmt.Errorf("Expected nil error but got %v", err)
	}

	f.state = openSent
	_, err = f.handleUpdate(u)
	if err == nil {
		fmt.Errorf("Expected an error for not being in established state")
	}
}
