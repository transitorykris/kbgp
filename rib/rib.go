package rib

// 3.1.  Routes: Advertisement and Storage

//    For the purpose of this protocol, a route is defined as a unit of
//    information that pairs a set of destinations with the attributes of a
//    path to those destinations.  The set of destinations are systems
//    whose IP addresses are contained in one IP address prefix that is
//    carried in the Network Layer Reachability Information (NLRI) field of
//    an UPDATE message, and the path is the information reported in the
//    path attributes field of the same UPDATE message.

//    Routes are advertised between BGP speakers in UPDATE messages.
//    Multiple routes that have the same path attributes can be advertised
//    in a single UPDATE message by including multiple prefixes in the NLRI
//    field of the UPDATE message.

//    Routes are stored in the Routing Information Bases (RIBs): namely,
//    the Adj-RIBs-In, the Loc-RIB, and the Adj-RIBs-Out, as described in
//    Section 3.2.

//    If a BGP speaker chooses to advertise a previously received route, it
//    MAY add to, or modify, the path attributes of the route before
//    advertising it to a peer.

//    BGP provides mechanisms by which a BGP speaker can inform its peers
//    that a previously advertised route is no longer available for use.
//    There are three methods by which a given BGP speaker can indicate
//    that a route has been withdrawn from service:

//       a) the IP prefix that expresses the destination for a previously
//          advertised route can be advertised in the WITHDRAWN ROUTES
//          field in the UPDATE message, thus marking the associated route
//          as being no longer available for use,

//       b) a replacement route with the same NLRI can be advertised, or

//       c) the BGP speaker connection can be closed, which implicitly
//          removes all routes the pair of speakers had advertised to each
//          other from service.

//    Changing the attribute(s) of a route is accomplished by advertising a
//    replacement route.  The replacement route carries new (changed)
//    attributes and has the same address prefix as the original route.

// 3.2.  Routing Information Base

//    The Routing Information Base (RIB) within a BGP speaker consists of
//    three distinct parts:

//       a) Adj-RIBs-In: The Adj-RIBs-In stores routing information learned
//          from inbound UPDATE messages that were received from other BGP
//          speakers.  Their contents represent routes that are available
//          as input to the Decision Process.

//       b) Loc-RIB: The Loc-RIB contains the local routing information the
//          BGP speaker selected by applying its local policies to the
//          routing information contained in its Adj-RIBs-In.  These are
//          the routes that will be used by the local BGP speaker.  The
//          next hop for each of these routes MUST be resolvable via the
//          local BGP speaker's Routing Table.

//       c) Adj-RIBs-Out: The Adj-RIBs-Out stores information the local BGP
//          speaker selected for advertisement to its peers.  The routing
//          information stored in the Adj-RIBs-Out will be carried in the
//          local BGP speaker's UPDATE messages and advertised to its
//          peers.

//    In summary, the Adj-RIBs-In contains unprocessed routing information
//    that has been advertised to the local BGP speaker by its peers; the
//    Loc-RIB contains the routes that have been selected by the local BGP
//    speaker's Decision Process; and the Adj-RIBs-Out organizes the routes
//    for advertisement to specific peers (by means of the local speaker's
//    UPDATE messages).

//    Although the conceptual model distinguishes between Adj-RIBs-In,
//    Loc-RIB, and Adj-RIBs-Out, this neither implies nor requires that an
//    implementation must maintain three separate copies of the routing
//    information.  The choice of implementation (for example, 3 copies of
//    the information vs 1 copy with pointers) is not constrained by the
//    protocol.

//    Routing information that the BGP speaker uses to forward packets (or
//    to construct the forwarding table used for packet forwarding) is
//    maintained in the Routing Table.  The Routing Table accumulates
//    routes to directly connected networks, static routes, routes learned
//    from the IGP protocols, and routes learned from BGP.  Whether a
//    specific BGP route should be installed in the Routing Table, and
//    whether a BGP route should override a route to the same destination
//    installed by another source, is a local policy decision, and is not
//    specified in this document.  In addition to actual packet forwarding,
//    the Routing Table is used for resolution of the next-hop addresses
//    specified in BGP updates (see Section 5.1.3).
