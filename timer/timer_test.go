package timer

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var ran bool
	f := func() {
		ran = true
	}
	ts := New(1*time.Second, f)
	if !ts.running {
		t.Errorf("Expected timer to be running but it's not")
	}
	time.Sleep(1100 * time.Millisecond)
	if !ran {
		t.Errorf("Timer did not call our function")
	}
}

func TestReset(t *testing.T) {
	var ran bool
	f := func() {
		ran = true
	}
	ts := New(1*time.Second, f)
	time.Sleep(500 * time.Millisecond)
	ts.Reset(1 * time.Second)
	time.Sleep(600 * time.Millisecond)
	if ran {
		t.Errorf("Timer called our function but it shouldn't have")
	}
	time.Sleep(500 * time.Millisecond)
	if !ran {
		t.Errorf("Timer did not call our function but should have")
	}
}

func TestStop(t *testing.T) {
	var ran bool
	f := func() {
		ran = true
	}
	ts := New(1*time.Second, f)
	ts.Stop()
	if ts.running {
		t.Errorf("Expected timer to be stopped but it's not")
	}
	time.Sleep(1100 * time.Millisecond)
	if ran {
		t.Errorf("Timer called our function but it shouldn't have")
	}
}

func TestRunning(t *testing.T) {
	f := func() {}
	ts := New(1*time.Second, f)
	if !ts.Running() {
		t.Errorf("Expected timer to be running but it's not")
	}
	ts.Stop()
	if ts.Running() {
		t.Errorf("Expected timer to be stopped but it's not")

	}
}
