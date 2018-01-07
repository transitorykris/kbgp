package policy

import (
	"github.com/transitorykris/kbgp/bgp"
)

// DefaultEBGP is a policy that implements https://tools.ietf.org/html/rfc8212
type DefaultEBGP struct{}

// Apply implements bgp.Policer
func (d *DefaultEBGP) Apply(nlri bgp.NLRI, attributes []bgp.PathAttribute) bool {
	return false
}

// DefaultIBGP is a policy that allows all prefixes to give a consistent
// view of routes exterior to the AS https://tools.ietf.org/html/rfc4271
type DefaultIBGP struct{}

// Apply implements bgp.Policer
func (d *DefaultIBGP) Apply(nlri bgp.NLRI, attributes []bgp.PathAttribute) bool {
	return true
}
