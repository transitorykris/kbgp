package speaker

import "net"

// NLRI provides network layer reachability information
type NLRI struct {
	prefix net.IPNet
}

// String provides the common string format for prefixes
func (n *NLRI) String() string {
	return n.prefix.String()
}
