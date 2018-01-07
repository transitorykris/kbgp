package bgp

// Reverse value lookup for event names
var eventName = map[Event]string{
	ManualStart:                                                      "ManualStart",
	ManualStop:                                                       "ManualStop",
	ManualStartWithPassiveTCPEstablishment:                           "ManualStart_with_PassiveTcpEstablishment",
	AutomaticStartWithPassiveTCPEstablishment:                        "AutomaticStart_with_PassiveTcpEstablishment",
	AutomaticStartWithDampPeerOscillations:                           "AutomaticStart_with_DampPeerOscillations",
	AutomaticStartWithDampPeerOscillationsAndPassiveTCPEstablishment: "AutomaticStart_with_DampPeerOscillations_and_PassiveTcpEstablishment",
	AutomaticStop:            "AutomaticStop",
	ConnectRetryTimerExpires: "ConnectRetryTimer_Expires",
	HoldTimerExpires:         "HoldTimer_Expires",
	KeepaliveTimerExpires:    "KeepaliveTimer_Expires",
	DelayOpenTimerExpires:    "DelayOpenTimer_Expires",
	IdleHoldTimerExpires:     "IdleHoldTimer_Expires",
	TCPConnectionValid:       "TcpConnection_Valid",
	TCPCRAcked:               "Tcp_CR_Acked",
	TCPConnectionConfirmed:   "TcpConnectionConfirmed",
	TCPConnectionFails:       "TcpConnectionFails",
	BGPOpen:                  "BGPOpen",
	BGPOpenWithDelayOpenTimerRunning: "BGPOpen with DelayOpenTimer running",
	BGPHeaderErr:                     "BGPHeaderErr",
	BGPOpenMsgErr:                    "BGPOpenMsgErr",
	OpenCollisionDump:                "OpenCollisionDump",
	NotifMsgVerErr:                   "NotifMsgVerErr",
	NotifMsg:                         "NotifMsg",
	KeepAliveMsg:                     "KeepAliveMsg",
	UpdateMsg:                        "UpdateMsg",
	UpdateMsgErr:                     "UpdateMsgErr",
}

// String implements string.Stringer
func (e Event) String() string {
	return eventName[e]
}

// Reverse value lookup for state names
var stateName = map[State]string{
	Idle:        "Idle",
	Connect:     "Connect",
	Active:      "Active",
	OpenConfirm: "OpenConfirm",
	Established: "Established",
}

// String implements string.Stringer
func (s State) String() string {
	return stateName[s]
}
