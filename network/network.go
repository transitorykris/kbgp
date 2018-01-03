package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// FindBGPIdentifier tries to find the best possible identifier from all
// interfaces configured on the host
func FindBGPIdentifier() (uint32, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return 0, err
	}
	// Note: this selection process is arbitrary
	for _, v := range ifs {
		addrs, err := v.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			// Make sure we have an IPv4 address
			if ip.To4() == nil {
				continue
			}
			// If it's routable, we have a winner!
			if ip.IsGlobalUnicast() {
				return ipToUint32(ip), nil
			}
		}
	}
	return 0, fmt.Errorf("No valid BGP identifier found")
}

func ipToUint32(ip net.IP) uint32 {
	// ip could be 4 or 16 bytes, let's be sure it's 4
	ip4 := ip.To4()
	u := binary.BigEndian.Uint32(ip4)
	return u
}

// Uint32ToIP converts a uint32 to a net.IP
func Uint32ToIP(i uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, i)
	return ip
}

// parseAddr splits the IP and port
func parseAddr(a net.Addr) (string, uint16) {
	addr := strings.Split(a.String(), ":")
	ip := addr[0]
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		port = 0
	}
	return ip, uint16(port)
}

// resolvable returns true if the given IP address is resolvable by the
// local system
func resolvable(ip net.IP) bool {
	// TODO: Implement me
	return true
}

// Appendix E.  TCP Options that May Be Used with BGP

//    If a local system TCP user interface supports the TCP PUSH function,
//    then each BGP message SHOULD be transmitted with PUSH flag set.
//    Setting PUSH flag forces BGP messages to be transmitted to the
//    receiver promptly.

//    If a local system TCP user interface supports setting the DSCP field
//    [RFC2474] for TCP connections, then the TCP connection used by BGP
//    SHOULD be opened with bits 0-2 of the DSCP field set to 110 (binary).

//    An implementation MUST support the TCP MD5 option [RFC2385].

// Security Considerations

//    A BGP implementation MUST support the authentication mechanism
//    specified in RFC 2385 [RFC2385].  The authentication provided by this
//    mechanism could be done on a per-peer basis.

//    BGP makes use of TCP for reliable transport of its traffic between
//    peer routers.  To provide connection-oriented integrity and data
//    origin authentication on a point-to-point basis, BGP specifies use of
//    the mechanism defined in RFC 2385.  These services are intended to
//    detect and reject active wiretapping attacks against the inter-router
//    TCP connections.  Absent the use of mechanisms that effect these
//    security services, attackers can disrupt these TCP connections and/or
//    masquerade as a legitimate peer router.  Because the mechanism
//    defined in the RFC does not provide peer-entity authentication, these
//    connections may be subject to some forms of replay attacks that will
//    not be detected at the TCP layer.  Such attacks might result in
//    delivery (from TCP) of "broken" or "spoofed" BGP messages.

//    The mechanism defined in RFC 2385 augments the normal TCP checksum
//    with a 16-byte message authentication code (MAC) that is computed
//    over the same data as the TCP checksum.  This MAC is based on a one-
//    way hash function (MD5) and use of a secret key.  The key is shared
//    between peer routers and is used to generate MAC values that are not
//    readily computed by an attacker who does not have access to the key.
//    A compliant implementation must support this mechanism, and must
//    allow a network administrator to activate it on a per-peer basis.

//    RFC 2385 does not specify a means of managing (e.g., generating,
//    distributing, and replacing) the keys used to compute the MAC.  RFC
//    3562 [RFC3562] (an informational document) provides some guidance in
//    this area, and provides rationale to support this guidance.  It notes
//    that a distinct key should be used for communication with each
//    protected peer.  If the same key is used for multiple peers, the
//    offered security services may be degraded, e.g., due to an increased
//    risk of compromise at one router that adversely affects other
//    routers.

//    The keys used for MAC computation should be changed periodically, to
//    minimize the impact of a key compromise or successful cryptanalytic
//    attack.  RFC 3562 suggests a crypto period (the interval during which
//    a key is employed) of, at most, 90 days.  More frequent key changes
//    reduce the likelihood that replay attacks (as described above) will
//    be feasible.  However, absent a standard mechanism for effecting such
//    changes in a coordinated fashion between peers, one cannot assume
//    that BGP-4 implementations complying with this RFC will support
//    frequent key changes.

//    Obviously, each should key also be chosen to be difficult for an
//    attacker to guess.  The techniques specified in RFC 1750 for random
//    number generation provide a guide for generation of values that could
//    be used as keys.  RFC 2385 calls for implementations to support keys
//    "composed of a string of printable ASCII of 80 bytes or less."  RFC
//    3562 suggests keys used in this context be 12 to 24 bytes of random
//    (pseudo-random) bits.  This is fairly consistent with suggestions for
//    analogous MAC algorithms, which typically employ keys in the range of
//    16 to 20 bytes.  To provide enough random bits at the low end of this
//    range, RFC 3562 also observes that a typical ACSII text string would
//    have to be close to the upper bound for the key length specified in
//    RFC 2385.
