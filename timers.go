package kbgp

import (
	"math/rand"
	"time"
)

// 10.  BGP Timers

//    BGP employs five timers: ConnectRetryTimer (see Section 8), HoldTimer
//    (see Section 4.2), KeepaliveTimer (see Section 8),
//    MinASOriginationIntervalTimer (see Section 9.2.1.2), and
//    MinRouteAdvertisementIntervalTimer (see Section 9.2.1.1).

//    Two optional timers MAY be supported: DelayOpenTimer, IdleHoldTimer
//    by BGP (see Section 8).  Section 8 describes their use.  The full
//    operation of these optional timers is outside the scope of this
//    document.

//    ConnectRetryTime is a mandatory FSM attribute that stores the initial
//    value for the ConnectRetryTimer.  The suggested default value for the
//    ConnectRetryTime is 120 seconds.

const defaultConnectRetryTime = 120 * time.Second

//    HoldTime is a mandatory FSM attribute that stores the initial value
//    for the HoldTimer.  The suggested default value for the HoldTime is
//    90 seconds.

const defaultHoldTime = 90 * time.Second

//    During some portions of the state machine (see Section 8), the
//    HoldTimer is set to a large value.  The suggested default for this
//    large value is 4 minutes.

const defaultLargeHoldTimer = 4 * time.Minute

//    The KeepaliveTime is a mandatory FSM attribute that stores the
//    initial value for the KeepaliveTimer.  The suggested default value
//    for the KeepaliveTime is 1/3 of the HoldTime.

const defaultKeepaliveTime = defaultHoldTime / 3

//    The suggested default value for the MinASOriginationIntervalTimer is
//    15 seconds.

const minASOriginationIntervalTimer = 15 * time.Second

//    The suggested default value for the
//    MinRouteAdvertisementIntervalTimer on EBGP connections is 30 seconds.

const minRouteAdvertisementIntervalTimerEBGP = 30 * time.Second

//    The suggested default value for the
//    MinRouteAdvertisementIntervalTimer on IBGP connections is 5 seconds.

const minRouteAdvertisementIntervalTimerIBGP = 5 * time.Second

//    An implementation of BGP MUST allow the HoldTimer to be configurable
//    on a per-peer basis, and MAY allow the other timers to be
//    configurable.

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
