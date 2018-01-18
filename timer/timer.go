package timer

import (
	"math/rand"
	"time"
)

// Timer provides a fancier timer than time.Timer
type Timer struct {
	timer   *time.Timer
	running bool
}

// New creates a new timer that will call the given function after
// the interval has elapsed
func New(d time.Duration, f func()) *Timer {
	t := &Timer{
		running: true,
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
func (t *Timer) Reset(d time.Duration) {
	if !t.timer.Stop() {
		<-t.timer.C
	}
	t.timer.Reset(d)
}

// Stop cancels the timer
func (t *Timer) Stop() {
	if !t.running {
		return
	}
	t.timer.Stop()
	t.running = false
}

// Running returns true if the timer is counting down, false otherwise
func (t *Timer) Running() bool {
	return t.running
}

//    To minimize the likelihood that the distribution of BGP messages by a
//    given BGP speaker will contain peaks, jitter SHOULD be applied to the
//    timers associated with MinASOriginationIntervalTimer, KeepaliveTimer,
//    MinRouteAdvertisementIntervalTimer, and ConnectRetryTimer.  A given
//    BGP speaker MAY apply the same jitter to each of these quantities,
//    regardless of the destinations to which the updates are being sent;
//    that is, jitter need not be configured on a per-peer basis.

//    The suggested default amount of jitter SHALL be determined by
//    multiplying the base value of the appropriate timer by a random
//    factor, which is uniformly distributed in the range from 0.75 to 1.0.
//    A new random value SHOULD be picked each time the timer is set.  The
//    range of the jitter's random value MAY be configurable.
func jitter() time.Duration {
	v := ((rand.Float64() / 4.0) + .75) * 1000
	j := time.Duration(v) * time.Millisecond
	return j
}
