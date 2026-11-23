// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

package provisioning

// Fixture describes everything a scenario needs provisioned before it
// runs. Plain data only.
type Fixture struct {
	// Operator, when non-nil, overrides the baseline operator config.
	Operator *Operator

	HomeNetworkKeys []HomeNetworkKey

	Profiles     []Profile
	Slices       []Slice
	DataNetworks []DataNetwork
	Policies     []Policy

	Subscribers []Subscriber

	// StaticIPs are pinned after subscribers and data networks exist.
	StaticIPs []StaticIP

	// FramedRoutes are provisioned after subscribers and data networks exist.
	FramedRoutes []FramedRoute

	PCSCFAddresses []string
}

type Operator struct {
	MCC           string
	MNC           string
	SupportedTACs []string
}

type HomeNetworkKey struct {
	KeyIdentifier int
	Scheme        string
	// PrivateKey is hex-encoded X25519; the core derives and serves the public key.
	PrivateKey string
}

type Profile struct {
	Name           string
	UeAmbrUplink   string
	UeAmbrDownlink string
}

type Slice struct {
	Name string
	SST  int
	SD   string
}

type DataNetwork struct {
	Name     string
	IPv4Pool string
	IPv6Pool string
	DNS      string
	MTU      int32
	IMS      *bool
}

type Policy struct {
	Name                    string
	ProfileName             string
	SliceName               string
	DataNetworkName         string
	SessionAmbrUplink       string
	SessionAmbrDownlink     string
	Var5qi                  int32
	Arp                     int32
	PreemptionCapability    string
	PreemptionVulnerability string
}

type Subscriber struct {
	IMSI           string
	Key            string
	OPc            string
	SequenceNumber string
	ProfileName    string
	MSISDN         string
}

// SubscriberUpdate changes a subscriber's profile and, when MSISDN is
// set, its MSISDN.
type SubscriberUpdate struct {
	IMSI        string
	ProfileName string
	MSISDN      string
}

// StaticIP pins an address to a subscriber for a data network. The IP
// version is inferred from the address family.
type StaticIP struct {
	IMSI        string
	DataNetwork string
	Address     string
}

// FramedRoute is a subscriber's framed routes on a data network
// (TS 23.501 §5.6.14). Families are inferred from each CIDR.
type FramedRoute struct {
	IMSI        string
	DataNetwork string
	IPv4        []string
	IPv6        []string
}

// SubscriberState is what the core knows about a subscriber right now.
type SubscriberState struct {
	Registrations []Registration
	Sessions      []Session
}

type Registration struct {
	System          string // "5G" or "4G"
	Registered      bool
	ConnectionState *string
}

// Session is a 5G PDU session or a 4G PDN connection.
type Session struct {
	System      string
	IPv4Address string
	IPv6Prefix  string
	DataNetwork string
}
