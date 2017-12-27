package kbgp

var eventName = map[int]string{
	1:  "ManualStart",
	2:  "ManualStop",
	3:  "AutomaticStart",
	4:  "ManualStart_with_PassiveTcpEstablishment",
	5:  "AutomaticStart_with_PassiveTcpEstablishment",
	6:  "AutomaticStart_with_DampPeerOscillations",
	7:  "AutomaticStart_with_DampPeerOscillations_and_PassiveTcpEstablishment",
	8:  "AutomaticStop",
	9:  "ConnectRetryTimer_Expires",
	10: "HoldTimer_Expires",
	11: "KeepaliveTimer_Expires",
	12: "KeepaliveTimer_Expires",
	13: "DelayOpenTimer_Expires",
	14: "IdleHoldTimer_Expires",
	15: "TcpConnection_Valid",
	16: "Tcp_CR_Acked",
	17: "TcpConnectionConfirmed",
	18: "TcpConnectionFails",
	19: "BGPOpen",
	20: "BGPOpen with DelayOpenTimer running",
	21: "BGPHeaderErr",
	22: "BGPOpenMsgErr",
	23: "OpenCollisionDump",
	24: "NotifMsgVerErr",
	25: "NotifMsg",
	26: "KeepAliveMsg",
	27: "UpdateMsg",
	28: "UpdateMsgErr",
}

var stateName = map[int]string{
	0: "Idle",
	1: "Connect",
	2: "Active",
	3: "OpenConfirm",
	4: "Established",
}

var messageName = map[int]string{
	1: "OPEN",
	2: "UPDATE",
	3: "NOTIFICATION",
	4: "KEEPALIVE",
	//5: "ROUTE-REFRESH" // [RFC2918] defines one more type code.
}

var originName = map[int]string{
	0: "IGP",
	1: "EGP",
	2: "INCOMPLETE",
}

var pathAttributeName = map[int]string{
	1: "ORIGIN",
	2: "AS_PATH",
	3: "NEXT_HOP",
	4: "MULTI_EXIT_DISC",
	5: "LOCAL_PREF",
	6: "ATOMIC_AGGREGATE",
	7: "AGGREGATOR",
}

var asPathName = map[int]string{
	1: "AS_SET",
	2: "AS_SEQUENCE",
}

var errorCodeName = map[int]string{
	1: "Message Header Error",
	2: "OPEN Message Error",
	3: "UPDATE Message Error",
	4: "Hold Timer Expired",
	5: "Finite State Machine Error",
	6: "Cease",
}

var messageHeaderErrorSubcodeName = map[int]string{
	1: "Connection Not Synchronized",
	2: "Bad Message Length",
	3: "Bad Message Type",
}

var openMessageErrorSubcodeName = map[int]string{
	1: "Unsupported Version Number",
	2: "Bad Peer AS",
	3: "Bad BGP Identifier",
	4: "Unsupported Optional Parameter",
	// 5 is deprecated
	6: "Unacceptable Hold Time",
}

var updateMessageErrorSubcodeName = map[int]string{
	1: "Malformed Attribute List",
	2: "Unrecognized Well-known Attribute",
	3: "Missing Well-known Attribute",
	4: "Attribute Flags Error",
	5: "Attribute Length Error",
	6: "Invalid ORIGIN Attribute",
	// 7 is deprecated
	8:  "Invalid NEXT_HOP Attribute",
	9:  "Optional Attribute Error",
	10: "Invalid Network Field",
	11: "Malformed AS_PATH",
}
