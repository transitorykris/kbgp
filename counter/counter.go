package counter

import (
	"fmt"
)

// Counter is a 64 bit counter
type Counter struct {
	count uint64
}

// New creates a new 64 bit counter
func New() *Counter {
	return new(Counter)
}

// Reset implements bgp.Counter
func (c *Counter) Reset() {
	c.count = 0
}

// Increment implements bgp.Counter
func (c *Counter) Increment() {
	c.count++
}

// Value implements bgp.Counter
func (c *Counter) Value() uint64 {
	return uint64(c.count)
}

// String implements strings.Stringer
func (c *Counter) String() string {
	return fmt.Sprintf("%d", c.count)
}
