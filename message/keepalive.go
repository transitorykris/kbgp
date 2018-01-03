package message

import "time"

// BGP does not use any TCP-based, keep-alive mechanism to determine if
// peers are reachable.  Instead, KEEPALIVE messages are exchanged
// between peers often enough not to cause the Hold Timer to expire.  A
// reasonable maximum time between KEEPALIVE messages would be one third
// of the Hold Time interval.  KEEPALIVE messages MUST NOT be sent more
// frequently than one per second.  An implementation MAY adjust the
// rate at which it sends KEEPALIVE messages as a function of the Hold
// Time interval.
const minKeepaliveInterval = 1 * time.Second

// If the negotiated Hold Time interval is zero, then periodic KEEPALIVE
// messages MUST NOT be sent.

// A KEEPALIVE message consists of only the message header and has a
// length of 19 octets.
type keepaliveMessage struct{}

func newKeepaliveMessage() keepaliveMessage {
	return keepaliveMessage{}
}

func readKeepalive(message []byte) *keepaliveMessage {
	// Related events
	//if len(message) != 0 {
	// Send a notification
	//	return nil, newNotificationMessage(messageHeaderError, badMessageLength, nil)
	//}
	// Note: this should occur elsewhere
	//f.sendEvent(keepAliveMsg)
	return &keepaliveMessage{}
}

func (k keepaliveMessage) valid() (*notificationMessage, bool) {
	// TODO: Implement me
	return nil, true
}

func (k keepaliveMessage) bytes() []byte {
	return []byte{}
}
