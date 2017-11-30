# kbgp

A silly exploration of BGP in Go. My plan is to follow the RFC closely, but split away from it as necessary to write idiomatic Go.

Comments will be liberally taken from the RFCs.

## Goals

* Implement RFC 4271
* Write idiomatic Go
* Re-learn BGP
* Have some fun

## Non-goals

* Implemented related RFCs unless necessary to complete RFC 4271
* Make the most efficient or optimized implementation

## RFC 4271

[A Border Gateway Protocol 4 (BGP-4)](https://tools.ietf.org/html/rfc4271)

The Border Gateway Protocol (BGP) is an inter-Autonomous System
routing protocol.

The primary function of a BGP speaking system is to exchange network
reachability information with other BGP systems.  This network
reachability information includes information on the list of
Autonomous Systems (ASes) that reachability information traverses.
This information is sufficient for constructing a graph of AS
connectivity for this reachability, from which routing loops may be
pruned and, at the AS level, some policy decisions may be enforced.

BGP-4 provides a set of mechanisms for supporting Classless Inter-
Domain Routing (CIDR) [RFC1518, RFC1519].  These mechanisms include
support for advertising a set of destinations as an IP prefix and
eliminating the concept of network "class" within BGP.  BGP-4 also
introduces mechanisms that allow aggregation of routes, including
aggregation of AS paths.

Routing information exchanged via BGP supports only the destination-
based forwarding paradigm, which assumes that a router forwards a
packet based solely on the destination address carried in the IP
header of the packet.  This, in turn, reflects the set of policy
decisions that can (and cannot) be enforced using BGP.  BGP can
support only those policies conforming to the destination-based
forwarding paradigm.

## Related RFCs

### BGP

* [Application of the Border Gateway Protocol in the Internet](https://tools.ietf.org/html/rfc1772)
* [Guidelines for creation, selection, and registration of an Autonomous System (AS)](https://tools.ietf.org/html/rfc1930)
* [BGP Communities Attribute](https://tools.ietf.org/html/rfc1997)
* [Multiprotocol Extensions for BGP-4](https://tools.ietf.org/html/rfc2858)
* [Route Refresh Capability for BGP-4](https://tools.ietf.org/html/rfc2918)
* [A Border Gateway Protocol 4 (BGP-4)](https://tools.ietf.org/html/rfc4271)

### CIDR

* [An Architecture for IP Address Allocation with CIDR](https://tools.ietf.org/html/rfc1518)
* [Classless Inter-Domain Routing (CIDR): an Address Assignment and Aggregation Strategy](https://tools.ietf.org/html/rfc1519)

### TCP

* [TRANSMISSION CONTROL PROTOCOL](https://tools.ietf.org/html/rfc793)