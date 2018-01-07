package radix

import (
	"net"
	"testing"
)

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Errorf("Expected a new trie to not be nil")
	}
}

func TestNewEdge(t *testing.T) {
	_, network, _ := net.ParseCIDR("10.1.1.0/24")
	nextHop := net.ParseIP("1.2.3.4")
	e := newEdge(*network, nextHop)
	if !e.nextHop.Equal(nextHop) {
		t.Error("Expected", e.nextHop, "to equal", nextHop)
	}
	if e.network.String() != network.String() {
		t.Error("Expected", e.network.String(), "to equal", network.String())
	}
}

func TestNewNode(t *testing.T) {
	n := newNode()
	if n == nil {
		t.Error("Did not expect node to be nil")
	}
}

func TestInsert(t *testing.T) {
	r := New()

	_, n, _ := net.ParseCIDR("10.1.1.0/24")
	r.Insert(*n, net.ParseIP("1.1.1.1"))

	_, n, _ = net.ParseCIDR("10.1.1.2/32")
	r.Insert(*n, net.ParseIP("1.1.1.2"))

	_, n, _ = net.ParseCIDR("10.1.1.1/32")
	r.Insert(*n, net.ParseIP("1.1.1.3"))

	_, n, _ = net.ParseCIDR("10.1.1.0/25")
	r.Insert(*n, net.ParseIP("1.1.1.4"))

	_, n, _ = net.ParseCIDR("10.1.2.2/24")
	r.Insert(*n, net.ParseIP("1.1.1.5"))

	_, n, _ = net.ParseCIDR("10.2.1.0/24")
	r.Insert(*n, net.ParseIP("1.1.1.6"))

	_, n, _ = net.ParseCIDR("10.2.0.0/16")
	r.Insert(*n, net.ParseIP("1.1.1.7"))

	// Try replacing the next hop
	r.Insert(*n, net.ParseIP("1.1.1.8"))
}

func TestDelete(t *testing.T) {
	r := New()
	_, n, _ := net.ParseCIDR("10.1.1.0/24")
	r.Delete(*n)
}

func TestLookup(t *testing.T) {
	r := New()

	_, n, _ := net.ParseCIDR("10.1.1.0/24")
	r.Insert(*n, net.ParseIP("1.1.1.1"))
	_, n, _ = net.ParseCIDR("10.1.1.2/32")
	r.Insert(*n, net.ParseIP("1.1.1.2"))
	_, n, _ = net.ParseCIDR("10.1.1.1/32")
	r.Insert(*n, net.ParseIP("1.1.1.3"))
	_, n, _ = net.ParseCIDR("10.1.1.0/25")
	r.Insert(*n, net.ParseIP("1.1.1.4"))
	_, n, _ = net.ParseCIDR("10.1.2.2/24")
	r.Insert(*n, net.ParseIP("1.1.1.5"))
	_, n, _ = net.ParseCIDR("10.2.1.0/24")
	r.Insert(*n, net.ParseIP("1.1.1.6"))
	_, n, _ = net.ParseCIDR("10.2.0.0/16")
	r.Insert(*n, net.ParseIP("1.1.1.7"))

	_, n, _ = net.ParseCIDR("10.1.2.2/32")
	_, nextHop, err := r.Lookup(*n)
	if err != nil {
		t.Error("Did not expect an error, got", err)
	}
	if !nextHop.Equal(net.ParseIP("1.1.1.5")) {
		t.Error("Expected next hop to be", net.ParseIP("1.1.1.5"), "but got", nextHop)
	}

	_, n, _ = net.ParseCIDR("192.2.2.2/32")
	_, _, err = r.Lookup(*n)
	if err == nil {
		t.Error("Did not expect error to be nil")
	}
}
