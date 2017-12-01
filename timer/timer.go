package timer

import "time"

// Timer provides a fancier timer than time.Timer
type Timer struct {
	timer    *time.Timer
	interval time.Duration
	running  bool
}

// New creates a new timer that will call the given function after
// the interval has elapsed
func New(d time.Duration, f func()) *Timer {
	t := &Timer{
		interval: d,
		running:  true,
	}
	t.timer = time.AfterFunc(d, t.preflight(f))
	return t
}

// preflight takes care of any housekeeping before calling the user's function
func (t *Timer) preflight(f func()) func() {
	p := func() {
		t.running = false
		f()
	}
	return p
}

// Reset starts the timer at its initial value
func (t *Timer) Reset() {
	if !t.timer.Stop() {
		<-t.timer.C
	}
	t.timer.Reset(t.interval)
}

// Stop cancels the timer
func (t *Timer) Stop() {
	if !t.timer.Stop() {
		<-t.timer.C
	}
	t.running = false
}

// Running returns true if the timer is counting down, false otherwise
func (t *Timer) Running() bool {
	return t.running
}
