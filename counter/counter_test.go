package counter

import "testing"

func TestNew(t *testing.T) {
	c := New()
	if c.Value() != 0 {
		t.Error("New counter has non-zero value", c.Value())
	}
}
