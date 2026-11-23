// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

package scenarios

import "github.com/ellanetworks/core/tester/provisioning"

// FixtureSpec describes everything a scenario needs provisioned on the core
// before it runs, plus how the runner should invoke and check it. Plain data
// only; the provisioning port applies it (see Provisioning).
type FixtureSpec struct {
	// Operator, when non-nil, overrides the baseline operator config that the
	// integration test otherwise applies from scenarios.Default* values.
	Operator *OperatorSpec

	HomeNetworkKeys []HomeNetworkKeySpec

	// Profiles/Slices/DataNetworks/Policies add to the default set.
	// Idempotent: names already provisioned are left alone.
	Profiles     []ProfileSpec
	Slices       []SliceSpec
	DataNetworks []DataNetworkSpec
	Policies     []PolicySpec

	Subscribers []SubscriberSpec

	// StaticIPs are pinned after subscribers and data networks exist.
	StaticIPs []StaticIPSpec

	// FramedRoutes are provisioned after subscribers and data networks exist.
	FramedRoutes []FramedRouteSpec

	PCSCFAddresses []string

	// ExtraArgs are passed verbatim to `core-tester run <scenario>`, for the
	// scenario-specific flags it declares via BindFlags.
	ExtraArgs []string

	// AssertUsageForIMSIs lists IMSIs whose uplink+downlink byte counters the
	// integration test polls and asserts > 0 post-scenario. Set only when the
	// scenario exercised the data plane.
	AssertUsageForIMSIs []string
}

// The fixture's parts are the provisioning port's own types, so a
// FixtureSpec converts to a provisioning.Fixture without copying fields.
type (
	OperatorSpec       = provisioning.Operator
	HomeNetworkKeySpec = provisioning.HomeNetworkKey
	ProfileSpec        = provisioning.Profile
	SliceSpec          = provisioning.Slice
	DataNetworkSpec    = provisioning.DataNetwork
	PolicySpec         = provisioning.Policy
	SubscriberSpec     = provisioning.Subscriber
	StaticIPSpec       = provisioning.StaticIP
	FramedRouteSpec    = provisioning.FramedRoute
)

// Provisioning returns the part of the spec the provisioning port applies.
func (s FixtureSpec) Provisioning() provisioning.Fixture {
	return provisioning.Fixture{
		Operator:        s.Operator,
		HomeNetworkKeys: s.HomeNetworkKeys,
		Profiles:        s.Profiles,
		Slices:          s.Slices,
		DataNetworks:    s.DataNetworks,
		Policies:        s.Policies,
		Subscribers:     s.Subscribers,
		StaticIPs:       s.StaticIPs,
		FramedRoutes:    s.FramedRoutes,
		PCSCFAddresses:  s.PCSCFAddresses,
	}
}

func DefaultSubscriber() SubscriberSpec {
	return SubscriberSpec{
		IMSI:           DefaultIMSI,
		Key:            DefaultKey,
		OPc:            DefaultOPC,
		SequenceNumber: DefaultSequenceNumber,
		ProfileName:    DefaultProfileName,
	}
}

func DefaultSubscriberWith(imsi string, profile string) SubscriberSpec {
	s := DefaultSubscriber()
	s.IMSI = imsi

	if profile != "" {
		s.ProfileName = profile
	}

	return s
}
