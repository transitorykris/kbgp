package speaker

// DefaultPolicy is a deny all policy
type DefaultPolicy struct{}

// Apply implement a Policer
func (d DefaultPolicy) Apply(n *NLRI) bool {
	return false
}

// PolicyInOption sets a custom inbound policy when creating a new peer
func PolicyInOption(policy Policer) PeerOption {
	return func(p *Peer) error {
		p.in = policy
		return nil
	}
}

// PolicyOutOption sets a custom outbound policy when creating a new peer
func PolicyOutOption(policy Policer) PeerOption {
	return func(p *Peer) error {
		p.out = policy
		return nil
	}
}
