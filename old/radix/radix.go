package radix

import (
	"fmt"
	"net"
)

// Trie is a generic trie for managing IP networks and their
// associated next hops
type Trie interface {
	Insert(net.IPNet, net.IP)
	Delete(net.IPNet) bool
	Lookup(net.IPNet) (net.IPNet, net.IP, error)
}

// Radix is a radix trie implementation specifically for managing
// IPv4 networks. We'll do this with IP networks as edges, it won't
// be efficient, but hopefully clear.
type Radix struct {
	root *node
}

// New creates an empty radix trie
func New() *Radix {
	r := &Radix{
		root: new(node),
	}
	return r
}

type edge struct {
	target  *node
	network net.IPNet
	nextHop net.IP
}

type node struct {
	edges []*edge
}

// leaf returns true if this node is a leaf
func (n *node) leaf() bool {
	if len(n.edges) == 0 {
		return true
	}
	return false
}

// Insert a new node into the trie
func (r *Radix) Insert(network net.IPNet, nextHop net.IP) {
	// Try to lookup the best match edge
	bestEdge := r.lookup(r.root, network)
	var bestNode *node
	if bestEdge == nil {
		bestNode = r.root
	} else if bestEdge.network.String() == network.String() {
		// Replace the data, we're done
		return
	} else {
		bestNode = bestEdge.target
	}
	// Otherwise this is more specific. Add a new edge.
	freshEdge := newEdge(network, nextHop)
	bestNode.edges = append(bestNode.edges, freshEdge)
	// Now check to see if we need to move any other edges under this new one
	for i, e := range bestNode.edges {
		if contains(network, e.network) {
			freshEdge.target.edges = append(freshEdge.target.edges, e)
			bestNode.edges = removeEdge(bestNode.edges, i)
		}
	}
}

func removeEdge(e []*edge, p int) []*edge {
	return append(e[:p], e[p+1:]...)
}

func contains(a net.IPNet, b net.IPNet) bool {
	if a.String() != b.String() {
		if a.Contains(b.IP) {
			return true
		}
	} else {
		fmt.Println(a.String(), "==", b.String())
	}
	return false
}

func newEdge(n net.IPNet, nextHop net.IP) *edge {
	e := &edge{
		target:  newNode(),
		network: n,
		nextHop: nextHop,
	}
	return e
}

func newNode() *node {
	return new(node)
}

// lookup finds the best edge for this address
func (r *Radix) lookup(n *node, network net.IPNet) *edge {
	var best *edge
	for _, e := range n.edges {
		if e.network.Contains(network.IP) {
			best = e
			if next := r.lookup(e.target, network); next == nil {
				return best
			}
		}
	}
	return best
}

// Delete a node from the trie. Returns true if a node was deleted.
func (r *Radix) Delete(n net.IPNet) bool {
	fmt.Println("Deleting", n)
	return true
}

// Lookup a node in the trie
func (r *Radix) Lookup(network net.IPNet) (net.IPNet, net.IP, error) {
	fmt.Println("Looking up", network)
	e := r.lookup(r.root, network)
	if e == nil {
		return net.IPNet{}, net.IP{}, fmt.Errorf("Route not found")
	}
	return e.network, e.nextHop, nil
}

// Dump prints the trie to stdout
func (r *Radix) Dump(n *node) {
	fmt.Println("------")
	for _, e := range n.edges {
		fmt.Println("Edge", e)
		r.Dump(e.target)
	}
}
