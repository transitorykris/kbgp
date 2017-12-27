package bgp

import "fmt"

// 9.  UPDATE Message Handling

func (f *fsm) handleUpdate(u *updateMessage) (*notificationMessage, error) {
	//    An UPDATE message may be received only in the Established state.
	//    Receiving an UPDATE message in any other state is an error.  When an
	//    UPDATE message is received, each field is checked for validity, as
	//    specified in Section 6.3.
	if f.state != established {
		return nil, fmt.Errorf("UPDATE received in %s state", stateName[f.state])
	}

	if notif, ok := u.valid(); !ok {
		return notif, fmt.Errorf("UPDATE message is not valid")
	}

	for _, a := range u.pathAttributes {
		//    If an optional non-transitive attribute is unrecognized, it is
		//    quietly ignored.  If an optional transitive attribute is
		//    unrecognized, the Partial bit (the third high-order bit) in the
		//    attribute flags octet is set to 1, and the attribute is retained for
		//    propagation to other BGP speakers.
		if a.attributeType.optional() && a.attributeType.nonTransitive() {
			// TODO: what optional transitive attributes do we recognize?
			a.setPartial()
		}

		//    If an optional attribute is recognized and has a valid value, then,
		//    depending on the type of the optional attribute, it is processed
		//    locally, retained, and updated, if necessary, for possible
		//    propagation to other BGP speakers.
		if a.attributeType.optional() {
		}
	}

	//    If the UPDATE message contains a non-empty WITHDRAWN ROUTES field,
	//    the previously advertised routes, whose destinations (expressed as IP
	//    prefixes) are contained in this field, SHALL be removed from the
	//    Adj-RIB-In.  This BGP speaker SHALL run its Decision Process because
	//    the previously advertised route is no longer available for use.
	if u.withdrawnRoutesLength != 0 {
		for _, w := range u.withdrawnRoutes {
			f.peer.adjRIBIn.remove(w)
		}
		// TODO: Run the decision process
	}

	//    If the UPDATE message contains a feasible route, the Adj-RIB-In will
	//    be updated with this route as follows: if the NLRI of the new route
	//    is identical to the one the route currently has stored in the Adj-
	//    RIB-In, then the new route SHALL replace the older route in the Adj-
	//    RIB-In, thus implicitly withdrawing the older route from service.
	//    Otherwise, if the Adj-RIB-In has no route with NLRI identical to the
	//    new route, the new route SHALL be placed in the Adj-RIB-In.
	if len(u.nlris) > 0 {
		for _, n := range u.nlris {
			f.peer.adjRIBIn.add(n, u.pathAttributes)
		}
		//    Once the BGP speaker updates the Adj-RIB-In, the speaker SHALL run
		//    its Decision Process.
		// TODO: Run the decision process
	}

	return nil, nil
}

// 9.1.  Decision Process

//    The Decision Process selects routes for subsequent advertisement by
//    applying the policies in the local Policy Information Base (PIB) to
//    the routes stored in its Adj-RIBs-In.  The output of the Decision
//    Process is the set of routes that will be advertised to peers; the
//    selected routes will be stored in the local speaker's Adj-RIBs-Out,
//    according to policy.

//    The BGP Decision Process described here is conceptual, and does not
//    have to be implemented precisely as described, as long as the
//    implementations support the described functionality and they exhibit
//    the same externally visible behavior.

//    The selection process is formalized by defining a function that takes
//    the attribute of a given route as an argument and returns either (a)
//    a non-negative integer denoting the degree of preference for the
//    route, or (b) a value denoting that this route is ineligible to be
//    installed in Loc-RIB and will be excluded from the next phase of
//    route selection.

//    The function that calculates the degree of preference for a given
//    route SHALL NOT use any of the following as its inputs: the existence
//    of other routes, the non-existence of other routes, or the path
//    attributes of other routes.  Route selection then consists of the
//    individual application of the degree of preference function to each
//    feasible route, followed by the choice of the one with the highest
//    degree of preference.

//    The Decision Process operates on routes contained in the Adj-RIBs-In,
//    and is responsible for:

//       - selection of routes to be used locally by the speaker

//       - selection of routes to be advertised to other BGP peers

//       - route aggregation and route information reduction

//    The Decision Process takes place in three distinct phases, each
//    triggered by a different event:

//       a) Phase 1 is responsible for calculating the degree of preference
//          for each route received from a peer.

//       b) Phase 2 is invoked on completion of phase 1.  It is responsible
//          for choosing the best route out of all those available for each
//          distinct destination, and for installing each chosen route into
//          the Loc-RIB.

//       c) Phase 3 is invoked after the Loc-RIB has been modified.  It is
//          responsible for disseminating routes in the Loc-RIB to each
//          peer, according to the policies contained in the PIB.  Route
//          aggregation and information reduction can optionally be
//          performed within this phase.

// 9.1.1.  Phase 1: Calculation of Degree of Preference

// Note: this will not be the ultimate implementation, just a way to
// study this section
func (f *fsm) phase1(s *Speaker) {
	//    The Phase 1 decision function is invoked whenever the local BGP
	//    speaker receives, from a peer, an UPDATE message that advertises a
	//    new route, a replacement route, or withdrawn routes.

	//    The Phase 1 decision function is a separate process,f which completes
	//    when it has no further work to do.

	//    The Phase 1 decision function locks an Adj-RIB-In prior to operating
	//    on any route contained within it, and unlocks it after operating on
	//    all new or unfeasible routes contained within it.
	f.peer.adjRIBIn.Lock()
	defer f.peer.adjRIBIn.Unlock()

	//    For each newly received or replacement feasible route, the local BGP
	//    speaker determines a degree of preference as follows:
	if f.peer.remoteAS == s.myAS {
		//       If the route is learned from an internal peer, either the value of
		//       the LOCAL_PREF attribute is taken as the degree of preference, or
		//       the local system computes the degree of preference of the route
		//       based on preconfigured policy information.  Note that the latter
		//       may result in formation of persistent routing loops.
	} else {
		//       If the route is learned from an external peer, then the local BGP
		//       speaker computes the degree of preference based on preconfigured
		//       policy information.  If the return value indicates the route is
		//       ineligible, the route MAY NOT serve as an input to the next phase
		//       of route selection; otherwise, the return value MUST be used as
		//       the LOCAL_PREF value in any IBGP readvertisement.
	}

	//       The exact nature of this policy information, and the computation
	//       involved, is a local matter.
}

// 9.1.2.  Phase 2: Route Selection

// Note: this will not be the ultimate implementation, just a way to
// study this section
func (s *Speaker) phase2() {
	//    The Phase 2 decision function is invoked on completion of Phase 1.
	//    The Phase 2 function is a separate process, which completes when it
	//    has no further work to do.  The Phase 2 process considers all routes
	//    that are eligible in the Adj-RIBs-In.

	//    The Phase 2 decision function is blocked from running while the Phase
	//    3 decision function is in process.  The Phase 2 function locks all
	//    Adj-RIBs-In prior to commencing its function, and unlocks them on
	//    completion.

	s.phase3Mutex.Lock()
	defer s.phase3Mutex.Unlock()

	for _, f := range s.fsm {
		f.peer.adjRIBIn.Lock()
		defer f.peer.adjRIBIn.Unlock()
		// Todo: double check this doesn't immediately fall out of scope
	}

	//    If the NEXT_HOP attribute of a BGP route depicts an address that is
	//    not resolvable, or if it would become unresolvable if the route was
	//    installed in the routing table, the BGP route MUST be excluded from
	//    the Phase 2 decision function.

	//    If the AS_PATH attribute of a BGP route contains an AS loop, the BGP
	//    route should be excluded from the Phase 2 decision function.  AS loop
	//    detection is done by scanning the full AS path (as specified in the
	//    AS_PATH attribute), and checking that the autonomous system number of
	//    the local system does not appear in the AS path.  Operations of a BGP
	//    speaker that is configured to accept routes with its own autonomous
	//    system number in the AS path are outside the scope of this document.

	//    It is critical that BGP speakers within an AS do not make conflicting
	//    decisions regarding route selection that would cause forwarding loops
	//    to occur.

	//    For each set of destinations for which a feasible route exists in the
	//    Adj-RIBs-In, the local BGP speaker identifies the route that has:

	//       a) the highest degree of preference of any route to the same set
	//          of destinations, or

	//       b) is the only route to that destination, or

	//       c) is selected as a result of the Phase 2 tie breaking rules
	//          specified in Section 9.1.2.2.

	//    The local speaker SHALL then install that route in the Loc-RIB,
	//    replacing any route to the same destination that is currently being
	//    held in the Loc-RIB.  When the new BGP route is installed in the
	//    Routing Table, care must be taken to ensure that existing routes to
	//    the same destination that are now considered invalid are removed from
	//    the Routing Table.  Whether the new BGP route replaces an existing
	//    non-BGP route in the Routing Table depends on the policy configured
	//    on the BGP speaker.

	//    The local speaker MUST determine the immediate next-hop address from
	//    the NEXT_HOP attribute of the selected route (see Section 5.1.3).  If
	//    either the immediate next-hop or the IGP cost to the NEXT_HOP (where
	//    the NEXT_HOP is resolved through an IGP route) changes, Phase 2 Route
	//    Selection MUST be performed again.

	//    Notice that even though BGP routes do not have to be installed in the
	//    Routing Table with the immediate next-hop(s), implementations MUST
	//    take care that, before any packets are forwarded along a BGP route,
	//    its associated NEXT_HOP address is resolved to the immediate
	//    (directly connected) next-hop address, and that this address (or
	//    multiple addresses) is finally used for actual packet forwarding.

	//    Unresolvable routes SHALL be removed from the Loc-RIB and the routing
	//    table.  However, corresponding unresolvable routes SHOULD be kept in
	//    the Adj-RIBs-In (in case they become resolvable).

	// 9.1.2.1.  Route Resolvability Condition

	//    As indicated in Section 9.1.2, BGP speakers SHOULD exclude
	//    unresolvable routes from the Phase 2 decision.  This ensures that
	//    only valid routes are installed in Loc-RIB and the Routing Table.

	//    The route resolvability condition is defined as follows:

	//       1) A route Rte1, referencing only the intermediate network
	//          address, is considered resolvable if the Routing Table contains
	//          at least one resolvable route Rte2 that matches Rte1's
	//          intermediate network address and is not recursively resolved
	//          (directly or indirectly) through Rte1.  If multiple matching
	//          routes are available, only the longest matching route SHOULD be
	//          considered.

	//       2) Routes referencing interfaces (with or without intermediate
	//          addresses) are considered resolvable if the state of the
	//          referenced interface is up and if IP processing is enabled on
	//          this interface.

	//    BGP routes do not refer to interfaces, but can be resolved through
	//    the routes in the Routing Table that can be of both types (those that
	//    specify interfaces or those that do not).  IGP routes and routes to
	//    directly connected networks are expected to specify the outbound
	//    interface.  Static routes can specify the outbound interface, the
	//    intermediate address, or both.

	//    Note that a BGP route is considered unresolvable in a situation where
	//    the BGP speaker's Routing Table contains no route matching the BGP
	//    route's NEXT_HOP.  Mutually recursive routes (routes resolving each
	//    other or themselves) also fail the resolvability check.

	//    It is also important that implementations do not consider feasible
	//    routes that would become unresolvable if they were installed in the
	//    Routing Table, even if their NEXT_HOPs are resolvable using the
	//    current contents of the Routing Table (an example of such routes

	//    would be mutually recursive routes).  This check ensures that a BGP
	//    speaker does not install routes in the Routing Table that will be
	//    removed and not used by the speaker.  Therefore, in addition to local
	//    Routing Table stability, this check also improves behavior of the
	//    protocol in the network.

	//    Whenever a BGP speaker identifies a route that fails the
	//    resolvability check because of mutual recursion, an error message
	//    SHOULD be logged.

	// 9.1.2.2.  Breaking Ties (Phase 2)

	//    In its Adj-RIBs-In, a BGP speaker may have several routes to the same
	//    destination that have the same degree of preference.  The local
	//    speaker can select only one of these routes for inclusion in the
	//    associated Loc-RIB.  The local speaker considers all routes with the
	//    same degrees of preference, both those received from internal peers,
	//    and those received from external peers.

	//    The following tie-breaking procedure assumes that, for each candidate
	//    route, all the BGP speakers within an autonomous system can ascertain
	//    the cost of a path (interior distance) to the address depicted by the
	//    NEXT_HOP attribute of the route, and follow the same route selection
	//    algorithm.

	//    The tie-breaking algorithm begins by considering all equally
	//    preferable routes to the same destination, and then selects routes to
	//    be removed from consideration.  The algorithm terminates as soon as
	//    only one route remains in consideration.  The criteria MUST be
	//    applied in the order specified.

	//    Several of the criteria are described using pseudo-code.  Note that
	//    the pseudo-code shown was chosen for clarity, not efficiency.  It is
	//    not intended to specify any particular implementation.  BGP
	//    implementations MAY use any algorithm that produces the same results
	//    as those described here.

	//       a) Remove from consideration all routes that are not tied for
	//          having the smallest number of AS numbers present in their
	//          AS_PATH attributes.  Note that when counting this number, an
	//          AS_SET counts as 1, no matter how many ASes are in the set.

	//       b) Remove from consideration all routes that are not tied for
	//          having the lowest Origin number in their Origin attribute.

	//       c) Remove from consideration routes with less-preferred
	//          MULTI_EXIT_DISC attributes.  MULTI_EXIT_DISC is only comparable
	//          between routes learned from the same neighboring AS (the
	//          neighboring AS is determined from the AS_PATH attribute).
	//          Routes that do not have the MULTI_EXIT_DISC attribute are
	//          considered to have the lowest possible MULTI_EXIT_DISC value.

	//          This is also described in the following procedure:

	//        for m = all routes still under consideration
	//            for n = all routes still under consideration
	//                if (neighborAS(m) == neighborAS(n)) and (MED(n) < MED(m))
	//                    remove route m from consideration

	//          In the pseudo-code above, MED(n) is a function that returns the
	//          value of route n's MULTI_EXIT_DISC attribute.  If route n has
	//          no MULTI_EXIT_DISC attribute, the function returns the lowest
	//          possible MULTI_EXIT_DISC value (i.e., 0).

	//          Similarly, neighborAS(n) is a function that returns the
	//          neighbor AS from which the route was received.  If the route is
	//          learned via IBGP, and the other IBGP speaker didn't originate
	//          the route, it is the neighbor AS from which the other IBGP
	//          speaker learned the route.  If the route is learned via IBGP,
	//          and the other IBGP speaker either (a) originated the route, or
	//          (b) created the route by aggregation and the AS_PATH attribute
	//          of the aggregate route is either empty or begins with an
	//          AS_SET, it is the local AS.

	//          If a MULTI_EXIT_DISC attribute is removed before re-advertising
	//          a route into IBGP, then comparison based on the received EBGP
	//          MULTI_EXIT_DISC attribute MAY still be performed.  If an
	//          implementation chooses to remove MULTI_EXIT_DISC, then the
	//          optional comparison on MULTI_EXIT_DISC, if performed, MUST be
	//          performed only among EBGP-learned routes.  The best EBGP-
	//          learned route may then be compared with IBGP-learned routes
	//          after the removal of the MULTI_EXIT_DISC attribute.  If
	//          MULTI_EXIT_DISC is removed from a subset of EBGP-learned
	//          routes, and the selected "best" EBGP-learned route will not
	//          have MULTI_EXIT_DISC removed, then the MULTI_EXIT_DISC must be
	//          used in the comparison with IBGP-learned routes.  For IBGP-
	//          learned routes, the MULTI_EXIT_DISC MUST be used in route
	//          comparisons that reach this step in the Decision Process.
	//          Including the MULTI_EXIT_DISC of an EBGP-learned route in the
	//          comparison with an IBGP-learned route, then removing the
	//          MULTI_EXIT_DISC attribute, and advertising the route has been
	//          proven to cause route loops.

	//       d) If at least one of the candidate routes was received via EBGP,
	//          remove from consideration all routes that were received via
	//          IBGP.

	//       e) Remove from consideration any routes with less-preferred
	//          interior cost.  The interior cost of a route is determined by
	//          calculating the metric to the NEXT_HOP for the route using the
	//          Routing Table.  If the NEXT_HOP hop for a route is reachable,
	//          but no cost can be determined, then this step should be skipped
	//          (equivalently, consider all routes to have equal costs).

	//          This is also described in the following procedure.

	//          for m = all routes still under consideration
	//              for n = all routes in still under consideration
	//                  if (cost(n) is lower than cost(m))
	//                      remove m from consideration

	//          In the pseudo-code above, cost(n) is a function that returns
	//          the cost of the path (interior distance) to the address given
	//          in the NEXT_HOP attribute of the route.

	//       f) Remove from consideration all routes other than the route that
	//          was advertised by the BGP speaker with the lowest BGP
	//          Identifier value.

	//       g) Prefer the route received from the lowest peer address.
}

// 9.1.3.  Phase 3: Route Dissemination

// Note: this will not be the ultimate implementation, just a way to
// study this section
func (f *fsm) phase3() {
	//    The Phase 3 decision function is invoked on completion of Phase 2, or
	//    when any of the following events occur:

	//       a) when routes in the Loc-RIB to local destinations have changed

	//       b) when locally generated routes learned by means outside of BGP
	//          have changed

	//       c) when a new BGP speaker connection has been established

	//    The Phase 3 function is a separate process that completes when it has
	//    no further work to do.  The Phase 3 Routing Decision function is
	//    blocked from running while the Phase 2 decision function is in
	//    process.

	//    All routes in the Loc-RIB are processed into Adj-RIBs-Out according
	//    to configured policy.  This policy MAY exclude a route in the Loc-RIB
	//    from being installed in a particular Adj-RIB-Out.  A route SHALL NOT

	//    be installed in the Adj-Rib-Out unless the destination, and NEXT_HOP
	//    described by this route, may be forwarded appropriately by the
	//    Routing Table.  If a route in Loc-RIB is excluded from a particular
	//    Adj-RIB-Out, the previously advertised route in that Adj-RIB-Out MUST
	//    be withdrawn from service by means of an UPDATE message (see 9.2).

	//    Route aggregation and information reduction techniques (see Section
	//    9.2.2.1) may optionally be applied.

	//    Any local policy that results in routes being added to an Adj-RIB-Out
	//    without also being added to the local BGP speaker's forwarding table
	//    is outside the scope of this document.

	//    When the updating of the Adj-RIBs-Out and the Routing Table is
	//    complete, the local BGP speaker runs the Update-Send process of 9.2.
}

// 9.1.4.  Overlapping Routes

//    A BGP speaker may transmit routes with overlapping Network Layer
//    Reachability Information (NLRI) to another BGP speaker.  NLRI overlap
//    occurs when a set of destinations are identified in non-matching
//    multiple routes.  Because BGP encodes NLRI using IP prefixes, overlap
//    will always exhibit subset relationships.  A route describing a
//    smaller set of destinations (a longer prefix) is said to be more
//    specific than a route describing a larger set of destinations (a
//    shorter prefix); similarly, a route describing a larger set of
//    destinations is said to be less specific than a route describing a
//    smaller set of destinations.

//    The precedence relationship effectively decomposes less specific
//    routes into two parts:

//       - a set of destinations described only by the less specific route,
//         and

//       - a set of destinations described by the overlap of the less
//         specific and the more specific routes

//    The set of destinations described by the overlap represents a portion
//    of the less specific route that is feasible, but is not currently in
//    use.  If a more specific route is later withdrawn, the set of
//    destinations described by the overlap will still be reachable using
//    the less specific route.

//    If a BGP speaker receives overlapping routes, the Decision Process
//    MUST consider both routes based on the configured acceptance policy.
//    If both a less and a more specific route are accepted, then the
//    Decision Process MUST install, in Loc-RIB, either both the less and

//    the more specific routes or aggregate the two routes and install, in
//    Loc-RIB, the aggregated route, provided that both routes have the
//    same value of the NEXT_HOP attribute.

//    If a BGP speaker chooses to aggregate, then it SHOULD either include
//    all ASes used to form the aggregate in an AS_SET, or add the
//    ATOMIC_AGGREGATE attribute to the route.  This attribute is now
//    primarily informational.  With the elimination of IP routing
//    protocols that do not support classless routing, and the elimination
//    of router and host implementations that do not support classless
//    routing, there is no longer a need to de-aggregate.  Routes SHOULD
//    NOT be de-aggregated.  In particular, a route that carries the
//    ATOMIC_AGGREGATE attribute MUST NOT be de-aggregated.  That is, the
//    NLRI of this route cannot be more specific.  Forwarding along such a
//    route does not guarantee that IP packets will actually traverse only
//    ASes listed in the AS_PATH attribute of the route.

// 9.2.  Update-Send Process

//    The Update-Send process is responsible for advertising UPDATE
//    messages to all peers.  For example, it distributes the routes chosen
//    by the Decision Process to other BGP speakers, which may be located
//    in either the same autonomous system or a neighboring autonomous
//    system.

//    When a BGP speaker receives an UPDATE message from an internal peer,
//    the receiving BGP speaker SHALL NOT re-distribute the routing
//    information contained in that UPDATE message to other internal peers
//    (unless the speaker acts as a BGP Route Reflector [RFC2796]).

//    As part of Phase 3 of the route selection process, the BGP speaker
//    has updated its Adj-RIBs-Out.  All newly installed routes and all
//    newly unfeasible routes for which there is no replacement route SHALL
//    be advertised to its peers by means of an UPDATE message.

//    A BGP speaker SHOULD NOT advertise a given feasible BGP route from
//    its Adj-RIB-Out if it would produce an UPDATE message containing the
//    same BGP route as was previously advertised.

//    Any routes in the Loc-RIB marked as unfeasible SHALL be removed.
//    Changes to the reachable destinations within its own autonomous
//    system SHALL also be advertised in an UPDATE message.

//    If, due to the limits on the maximum size of an UPDATE message (see
//    Section 4), a single route doesn't fit into the message, the BGP
//    speaker MUST not advertise the route to its peers and MAY choose to
//    log an error locally.

// 9.2.1.  Controlling Routing Traffic Overhead

//    The BGP protocol constrains the amount of routing traffic (that is,
//    UPDATE messages), in order to limit both the link bandwidth needed to
//    advertise UPDATE messages and the processing power needed by the
//    Decision Process to digest the information contained in the UPDATE
//    messages.

// 9.2.1.1.  Frequency of Route Advertisement

//    The parameter MinRouteAdvertisementIntervalTimer determines the
//    minimum amount of time that must elapse between an advertisement
//    and/or withdrawal of routes to a particular destination by a BGP
//    speaker to a peer.  This rate limiting procedure applies on a per-
//    destination basis, although the value of
//    MinRouteAdvertisementIntervalTimer is set on a per BGP peer basis.

//    Two UPDATE messages sent by a BGP speaker to a peer that advertise
//    feasible routes and/or withdrawal of unfeasible routes to some common
//    set of destinations MUST be separated by at least
//    MinRouteAdvertisementIntervalTimer.  This can only be achieved by
//    keeping a separate timer for each common set of destinations.  This
//    would be unwarranted overhead.  Any technique that ensures that the
//    interval between two UPDATE messages sent from a BGP speaker to a
//    peer that advertise feasible routes and/or withdrawal of unfeasible
//    routes to some common set of destinations will be at least
//    MinRouteAdvertisementIntervalTimer, and will also ensure that a
//    constant upper bound on the interval is acceptable.

//    Since fast convergence is needed within an autonomous system, either
//    (a) the MinRouteAdvertisementIntervalTimer used for internal peers
//    SHOULD be shorter than the MinRouteAdvertisementIntervalTimer used
//    for external peers, or (b) the procedure describe in this section
//    SHOULD NOT apply to routes sent to internal peers.

//    This procedure does not limit the rate of route selection, but only
//    the rate of route advertisement.  If new routes are selected multiple
//    times while awaiting the expiration of
//    MinRouteAdvertisementIntervalTimer, the last route selected SHALL be
//    advertised at the end of MinRouteAdvertisementIntervalTimer.

// 9.2.1.2.  Frequency of Route Origination

//    The parameter MinASOriginationIntervalTimer determines the minimum
//    amount of time that must elapse between successive advertisements of
//    UPDATE messages that report changes within the advertising BGP
//    speaker's own autonomous systems.

// 9.2.2.  Efficient Organization of Routing Information

//    Having selected the routing information it will advertise, a BGP
//    speaker may avail itself of several methods to organize this
//    information in an efficient manner.

// 9.2.2.1.  Information Reduction

//    Information reduction may imply a reduction in granularity of policy
//    control - after information is collapsed, the same policies will
//    apply to all destinations and paths in the equivalence class.

//    The Decision Process may optionally reduce the amount of information
//    that it will place in the Adj-RIBs-Out by any of the following
//    methods:

//       a) Network Layer Reachability Information (NLRI):

//          Destination IP addresses can be represented as IP address
//          prefixes.  In cases where there is a correspondence between the
//          address structure and the systems under control of an
//          autonomous system administrator, it will be possible to reduce
//          the size of the NLRI carried in the UPDATE messages.

//       b) AS_PATHs:

//          AS path information can be represented as ordered AS_SEQUENCEs
//          or unordered AS_SETs.  AS_SETs are used in the route
//          aggregation algorithm described in Section 9.2.2.2.  They
//          reduce the size of the AS_PATH information by listing each AS
//          number only once, regardless of how many times it may have
//          appeared in multiple AS_PATHs that were aggregated.

//          An AS_SET implies that the destinations listed in the NLRI can
//          be reached through paths that traverse at least some of the
//          constituent autonomous systems.  AS_SETs provide sufficient
//          information to avoid routing information looping; however,
//          their use may prune potentially feasible paths because such
//          paths are no longer listed individually in the form of
//          AS_SEQUENCEs.  In practice, this is not likely to be a problem
//          because once an IP packet arrives at the edge of a group of
//          autonomous systems, the BGP speaker is likely to have more
//          detailed path information and can distinguish individual paths
//          from destinations.

// 9.2.2.2.  Aggregating Routing Information

//    Aggregation is the process of combining the characteristics of
//    several different routes in such a way that a single route can be
//    advertised.  Aggregation can occur as part of the Decision Process to
//    reduce the amount of routing information that will be placed in the
//    Adj-RIBs-Out.

//    Aggregation reduces the amount of information that a BGP speaker must
//    store and exchange with other BGP speakers.  Routes can be aggregated
//    by applying the following procedure, separately, to path attributes
//    of the same type and to the Network Layer Reachability Information.

//    Routes that have different MULTI_EXIT_DISC attributes SHALL NOT be
//    aggregated.

//    If the aggregated route has an AS_SET as the first element in its
//    AS_PATH attribute, then the router that originates the route SHOULD
//    NOT advertise the MULTI_EXIT_DISC attribute with this route.

//    Path attributes that have different type codes cannot be aggregated
//    together.  Path attributes of the same type code may be aggregated,
//    according to the following rules:

//       NEXT_HOP:
//          When aggregating routes that have different NEXT_HOP
//          attributes, the NEXT_HOP attribute of the aggregated route
//          SHALL identify an interface on the BGP speaker that performs
//          the aggregation.

//       ORIGIN attribute:
//          If at least one route among routes that are aggregated has
//          ORIGIN with the value INCOMPLETE, then the aggregated route
//          MUST have the ORIGIN attribute with the value INCOMPLETE.
//          Otherwise, if at least one route among routes that are
//          aggregated has ORIGIN with the value EGP, then the aggregated
//          route MUST have the ORIGIN attribute with the value EGP.  In
//          all other cases,, the value of the ORIGIN attribute of the
//          aggregated route is IGP.

//       AS_PATH attribute:
//          If routes to be aggregated have identical AS_PATH attributes,
//          then the aggregated route has the same AS_PATH attribute as
//          each individual route.

//          For the purpose of aggregating AS_PATH attributes, we model
//          each AS within the AS_PATH attribute as a tuple <type, value>,
//          where "type" identifies a type of the path segment the AS

//          belongs to (e.g., AS_SEQUENCE, AS_SET), and "value" identifies
//          the AS number.  If the routes to be aggregated have different
//          AS_PATH attributes, then the aggregated AS_PATH attribute SHALL
//          satisfy all of the following conditions:

//            - all tuples of type AS_SEQUENCE in the aggregated AS_PATH
//              SHALL appear in all of the AS_PATHs in the initial set of
//              routes to be aggregated.

//            - all tuples of type AS_SET in the aggregated AS_PATH SHALL
//              appear in at least one of the AS_PATHs in the initial set
//              (they may appear as either AS_SET or AS_SEQUENCE types).

//            - for any tuple X of type AS_SEQUENCE in the aggregated
//              AS_PATH, which precedes tuple Y in the aggregated AS_PATH,
//              X precedes Y in each AS_PATH in the initial set, which
//              contains Y, regardless of the type of Y.

//            - No tuple of type AS_SET with the same value SHALL appear
//              more than once in the aggregated AS_PATH.

//            - Multiple tuples of type AS_SEQUENCE with the same value may
//              appear in the aggregated AS_PATH only when adjacent to
//              another tuple of the same type and value.

//          An implementation may choose any algorithm that conforms to
//          these rules.  At a minimum, a conformant implementation SHALL
//          be able to perform the following algorithm that meets all of
//          the above conditions:

//            - determine the longest leading sequence of tuples (as
//              defined above) common to all the AS_PATH attributes of the
//              routes to be aggregated.  Make this sequence the leading
//              sequence of the aggregated AS_PATH attribute.

//            - set the type of the rest of the tuples from the AS_PATH
//              attributes of the routes to be aggregated to AS_SET, and
//              append them to the aggregated AS_PATH attribute.

//            - if the aggregated AS_PATH has more than one tuple with the
//              same value (regardless of tuple's type), eliminate all but
//              one such tuple by deleting tuples of the type AS_SET from
//              the aggregated AS_PATH attribute.

//            - for each pair of adjacent tuples in the aggregated AS_PATH,
//              if both tuples have the same type, merge them together, as
//              long as doing so will not cause a segment with a length
//              greater than 255 to be generated.

//          Appendix F, Section F.6 presents another algorithm that
//          satisfies the conditions and allows for more complex policy
//          configurations.

//       ATOMIC_AGGREGATE:
//          If at least one of the routes to be aggregated has
//          ATOMIC_AGGREGATE path attribute, then the aggregated route
//          SHALL have this attribute as well.

//       AGGREGATOR:
//          Any AGGREGATOR attributes from the routes to be aggregated MUST
//          NOT be included in the aggregated route.  The BGP speaker
//          performing the route aggregation MAY attach a new AGGREGATOR
//          attribute (see Section 5.1.7).

// 9.3.  Route Selection Criteria

//    Generally, additional rules for comparing routes among several
//    alternatives are outside the scope of this document.  There are two
//    exceptions:

//       - If the local AS appears in the AS path of the new route being
//         considered, then that new route cannot be viewed as better than
//         any other route (provided that the speaker is configured to
//         accept such routes).  If such a route were ever used, a routing
//         loop could result.

//       - In order to achieve a successful distributed operation, only
//         routes with a likelihood of stability can be chosen.  Thus, an
//         AS SHOULD avoid using unstable routes, and it SHOULD NOT make
//         rapid, spontaneous changes to its choice of route.  Quantifying
//         the terms "unstable" and "rapid" (from the previous sentence)
//         will require experience, but the principle is clear.  Routes
//         that are unstable can be "penalized" (e.g., by using the
//         procedures described in [RFC2439]).

// 9.4.  Originating BGP routes

//    A BGP speaker may originate BGP routes by injecting routing
//    information acquired by some other means (e.g., via an IGP) into BGP.
//    A BGP speaker that originates BGP routes assigns the degree of
//    preference (e.g., according to local configuration) to these routes
//    by passing them through the Decision Process (see Section 9.1).
//    These routes MAY also be distributed to other BGP speakers within the
//    local AS as part of the update process (see Section 9.2).  The
//    decision of whether to distribute non-BGP acquired routes within an
//    AS via BGP depends on the environment within the AS (e.g., type of
//    IGP) and SHOULD be controlled via configuration.
