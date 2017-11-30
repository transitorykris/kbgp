package kbgp

import "testing"

func TestMarker(t *testing.T) {
	m := marker()
	if len(m) != markerLength {
		t.Errorf("Expected marker length %d but got %d", markerLength, len(m))
	}
	for i, v := range m {
		if v != 0xFF {
			t.Errorf("Expected all bits to be 1, got %d at position %d", v, i)
		}
	}
}
