// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package ellacore is the provisioning port's adapter to Ella Core. It
// wraps Ella Core's REST client; it is the only tester code that imports
// that client.
package ellacore

import (
	"context"
	"crypto/ecdh"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ellanetworks/core/client"
	"github.com/ellanetworks/core/tester/provisioning"
)

// Adapter implements provisioning.Port over Ella Core's REST API.
type Adapter struct {
	c *client.Client
}

var _ provisioning.Port = (*Adapter)(nil)

// New wraps an existing Ella Core client.
func New(c *client.Client) *Adapter { return &Adapter{c: c} }

// Open is a provisioning.Opener for Ella Core.
func Open(_ context.Context, t provisioning.Target) (provisioning.Port, error) {
	c, err := client.New(&client.Config{BaseURL: t.Address})
	if err != nil {
		return nil, err
	}

	if t.Token != "" {
		c.SetToken(t.Token)
	}

	return New(c), nil
}

// Client returns the Ella Core client behind p, for the scenarios that
// exercise Ella-Core-only surfaces (IMS, SMS settings, location) the
// provisioning port does not model. It fails when p is another core's
// adapter, so such scenarios report that they need Ella Core.
func Client(p provisioning.Port) (*client.Client, error) {
	a, ok := p.(*Adapter)
	if !ok {
		return nil, fmt.Errorf("this scenario needs Ella Core's own API, and the provisioning port is a %T", p)
	}

	return a.c, nil
}

// Bootstrap waits for a freshly started Ella Core at address to be ready,
// creates its first admin user and an API token, and returns an adapter
// that uses the token, together with the token.
func Bootstrap(ctx context.Context, address string, timeout time.Duration) (*Adapter, string, error) {
	c, err := client.New(&client.Config{BaseURL: address})
	if err != nil {
		return nil, "", err
	}

	deadline := time.Now().Add(timeout)

	for {
		status, err := c.GetStatus(ctx)
		if err == nil && status.Ready {
			break
		}

		if time.Now().After(deadline) {
			if err == nil {
				err = errors.New("API up but not ready")
			}

			return nil, "", fmt.Errorf("ella core at %s not ready after %s: %w", address, timeout, err)
		}

		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}

	if err := c.Initialize(ctx, &client.InitializeOptions{Email: "admin@ellanetworks.com", Password: "admin"}); err != nil {
		return nil, "", fmt.Errorf("initialize: %w", err)
	}

	resp, err := c.CreateMyAPIToken(ctx, &client.CreateAPITokenOptions{Name: "provisioning-port"})
	if err != nil {
		return nil, "", fmt.Errorf("create API token: %w", err)
	}

	c.SetToken(resp.Token)

	return New(c), resp.Token, nil
}

// Apply provisions f; see provisioning.Port.
func (a *Adapter) Apply(ctx context.Context, f provisioning.Fixture) (provisioning.Teardown, error) {
	var undo []func(context.Context) error

	teardown := func(ctx context.Context) error {
		var errs []error

		for i := len(undo) - 1; i >= 0; i-- {
			if err := undo[i](ctx); err != nil {
				errs = append(errs, err)
			}
		}

		undo = nil

		return errors.Join(errs...)
	}

	fail := func(err error) (provisioning.Teardown, error) {
		_ = teardown(ctx)
		return nil, err
	}

	if f.Operator != nil {
		if err := a.operator(ctx, *f.Operator); err != nil {
			return fail(err)
		}
	}

	for _, k := range f.HomeNetworkKeys {
		if err := a.homeNetworkKey(ctx, k); err != nil {
			return fail(err)
		}
	}

	for _, p := range f.Profiles {
		if err := a.c.CreateProfile(ctx, &client.CreateProfileOptions{Name: p.Name, UeAmbrUplink: p.UeAmbrUplink, UeAmbrDownlink: p.UeAmbrDownlink}); err != nil {
			return fail(fmt.Errorf("create profile %q: %w", p.Name, err))
		}

		name := p.Name
		undo = append(undo, func(ctx context.Context) error {
			return wrap("delete profile "+name, a.c.DeleteProfile(ctx, &client.DeleteProfileOptions{Name: name}))
		})
	}

	for _, s := range f.Slices {
		if err := a.c.CreateSlice(ctx, &client.CreateSliceOptions{Name: s.Name, Sst: s.SST, Sd: s.SD}); err != nil {
			return fail(fmt.Errorf("create slice %q: %w", s.Name, err))
		}

		name := s.Name
		undo = append(undo, func(ctx context.Context) error {
			return wrap("delete slice "+name, a.c.DeleteSlice(ctx, &client.DeleteSliceOptions{Name: name}))
		})
	}

	for _, dn := range f.DataNetworks {
		if err := a.c.CreateDataNetwork(ctx, &client.CreateDataNetworkOptions{
			Name: dn.Name, IPv4Pool: dn.IPv4Pool, IPv6Pool: dn.IPv6Pool, DNS: dn.DNS, Mtu: dn.MTU, IMS: dn.IMS,
		}); err != nil {
			return fail(fmt.Errorf("create data network %q: %w", dn.Name, err))
		}

		name := dn.Name
		undo = append(undo, func(ctx context.Context) error {
			return wrap("delete data network "+name, a.c.DeleteDataNetwork(ctx, &client.DeleteDataNetworkOptions{Name: name}))
		})
	}

	for _, p := range f.Policies {
		if err := a.serveDataNetwork(ctx, p.SliceName, p.DataNetworkName); err != nil {
			return fail(err)
		}

		if err := a.c.CreatePolicy(ctx, &client.CreatePolicyOptions{
			Name: p.Name, ProfileName: p.ProfileName, SliceName: p.SliceName, DataNetworkName: p.DataNetworkName,
			SessionAmbrUplink: p.SessionAmbrUplink, SessionAmbrDownlink: p.SessionAmbrDownlink,
			Var5qi: p.Var5qi, Arp: p.Arp,
			PreemptionCapability: p.PreemptionCapability, PreemptionVulnerability: p.PreemptionVulnerability,
		}); err != nil {
			return fail(fmt.Errorf("create policy %q: %w", p.Name, err))
		}

		name := p.Name
		undo = append(undo, func(ctx context.Context) error {
			return wrap("delete policy "+name, a.c.DeletePolicy(ctx, &client.DeletePolicyOptions{Name: name}))
		})
	}

	for _, s := range f.Subscribers {
		if err := a.c.CreateSubscriber(ctx, &client.CreateSubscriberOptions{
			Imsi: s.IMSI, Key: s.Key, SequenceNumber: s.SequenceNumber, ProfileName: s.ProfileName, OPc: s.OPc, Msisdn: s.MSISDN,
		}); err != nil {
			return fail(fmt.Errorf("create subscriber %q: %w", s.IMSI, err))
		}

		imsi := s.IMSI
		undo = append(undo, func(ctx context.Context) error {
			return wrap("delete subscriber "+imsi, a.c.DeleteSubscriber(ctx, &client.DeleteSubscriberOptions{ID: imsi}))
		})
	}

	for _, s := range f.StaticIPs {
		if err := a.c.CreateDataNetworkStaticIp(ctx, s.DataNetwork, &client.CreateStaticIPOptions{IMSI: s.IMSI, Address: s.Address}); err != nil {
			return fail(fmt.Errorf("create static IP %s for %s: %w", s.Address, s.IMSI, err))
		}

		dn, imsi, version := s.DataNetwork, s.IMSI, "ipv4"
		if strings.Contains(s.Address, ":") {
			version = "ipv6"
		}

		undo = append(undo, func(ctx context.Context) error {
			return wrap("delete static IP for "+imsi, a.c.DeleteDataNetworkStaticIp(ctx, dn, imsi, version))
		})
	}

	for _, r := range f.FramedRoutes {
		if err := a.CreateFramedRoute(ctx, r); err != nil {
			return fail(err)
		}

		dn, imsi := r.DataNetwork, r.IMSI
		undo = append(undo, func(ctx context.Context) error { return a.DeleteFramedRoute(ctx, dn, imsi) })
	}

	if len(f.PCSCFAddresses) > 0 {
		current, err := a.c.GetIMSSettings(ctx)
		if err != nil {
			return fail(fmt.Errorf("get IMS settings: %w", err))
		}

		if err := a.c.UpdateIMSSettings(ctx, &client.UpdateIMSSettingsOptions{
			PCSCFAddresses: f.PCSCFAddresses, HomeDomain: current.HomeDomain, AdvertiseVoice: current.AdvertiseVoice, RestrictPeers: current.RestrictPeers,
		}); err != nil {
			return fail(fmt.Errorf("set P-CSCF addresses %v: %w", f.PCSCFAddresses, err))
		}

		undo = append(undo, func(ctx context.Context) error {
			return wrap("clear P-CSCF addresses", a.c.UpdateIMSSettings(ctx, &client.UpdateIMSSettingsOptions{
				HomeDomain: current.HomeDomain, AdvertiseVoice: current.AdvertiseVoice, RestrictPeers: current.RestrictPeers,
			}))
		})
	}

	return teardown, nil
}

func (a *Adapter) operator(ctx context.Context, o provisioning.Operator) error {
	if err := a.c.UpdateOperatorID(ctx, &client.UpdateOperatorIDOptions{Mcc: o.MCC, Mnc: o.MNC}); err != nil {
		return fmt.Errorf("update operator ID: %w", err)
	}

	if len(o.SupportedTACs) > 0 {
		if err := a.c.UpdateOperatorTracking(ctx, &client.UpdateOperatorTrackingOptions{SupportedTacs: o.SupportedTACs}); err != nil {
			return fmt.Errorf("update operator tracking: %w", err)
		}
	}

	return nil
}

func (a *Adapter) homeNetworkKey(ctx context.Context, k provisioning.HomeNetworkKey) error {
	if k.PrivateKey == "" {
		return nil
	}

	raw, err := hex.DecodeString(k.PrivateKey)
	if err != nil {
		return fmt.Errorf("decode home network private key: %w", err)
	}

	priv, err := ecdh.X25519().NewPrivateKey(raw)
	if err != nil {
		return fmt.Errorf("construct X25519 private key: %w", err)
	}

	pubHex := hex.EncodeToString(priv.PublicKey().Bytes())

	op, err := a.c.GetOperator(ctx)
	if err != nil {
		return fmt.Errorf("get operator: %w", err)
	}

	for _, existing := range op.HomeNetworkKeys {
		if existing.PublicKey == pubHex {
			return nil
		}
	}

	if err := a.c.CreateHomeNetworkKey(ctx, &client.CreateHomeNetworkKeyOptions{
		KeyIdentifier: k.KeyIdentifier, Scheme: k.Scheme, PrivateKey: hex.EncodeToString(priv.Bytes()),
	}); err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("create home network key: %w", err)
	}

	return nil
}

func (a *Adapter) serveDataNetwork(ctx context.Context, slice, dataNetwork string) error {
	if _, err := a.c.GetSliceDataNetwork(ctx, slice, dataNetwork); err == nil {
		return nil
	}

	if err := a.c.CreateSliceDataNetwork(ctx, slice, &client.CreateSliceDataNetworkOptions{DataNetwork: dataNetwork}); err != nil {
		return fmt.Errorf("serve data network %q on slice %q: %w", dataNetwork, slice, err)
	}

	return nil
}

func (a *Adapter) UpdateDataNetwork(ctx context.Context, dn provisioning.DataNetwork) error {
	return a.c.UpdateDataNetwork(ctx, &client.UpdateDataNetworkOptions{
		Name: dn.Name, IPv4Pool: dn.IPv4Pool, IPv6Pool: dn.IPv6Pool, DNS: dn.DNS, Mtu: dn.MTU, IMS: dn.IMS,
	})
}

func (a *Adapter) UpdatePolicy(ctx context.Context, p provisioning.Policy) error {
	return a.c.UpdatePolicy(ctx, p.Name, &client.UpdatePolicyOptions{
		ProfileName: p.ProfileName, SliceName: p.SliceName, DataNetworkName: p.DataNetworkName,
		SessionAmbrUplink: p.SessionAmbrUplink, SessionAmbrDownlink: p.SessionAmbrDownlink,
		Var5qi: p.Var5qi, Arp: p.Arp,
		PreemptionCapability: p.PreemptionCapability, PreemptionVulnerability: p.PreemptionVulnerability,
	})
}

func (a *Adapter) UpdateProfile(ctx context.Context, p provisioning.Profile) error {
	return a.c.UpdateProfile(ctx, p.Name, &client.UpdateProfileOptions{UeAmbrUplink: p.UeAmbrUplink, UeAmbrDownlink: p.UeAmbrDownlink})
}

func (a *Adapter) UpdateSubscriber(ctx context.Context, u provisioning.SubscriberUpdate) error {
	return a.c.UpdateSubscriber(ctx, u.IMSI, &client.UpdateSubscriberOptions{ProfileName: u.ProfileName, Msisdn: u.MSISDN})
}

func (a *Adapter) DeleteSubscriber(ctx context.Context, imsi string) error {
	return a.c.DeleteSubscriber(ctx, &client.DeleteSubscriberOptions{ID: imsi})
}

func (a *Adapter) Subscriber(ctx context.Context, imsi string) (provisioning.SubscriberState, error) {
	sub, err := a.c.GetSubscriber(ctx, &client.GetSubscriberOptions{ID: imsi})
	if err != nil {
		return provisioning.SubscriberState{}, err
	}

	var st provisioning.SubscriberState

	for _, r := range sub.Registrations {
		st.Registrations = append(st.Registrations, provisioning.Registration{
			System: r.System, Registered: r.Registered, ConnectionState: r.ConnectionState,
		})
	}

	for _, s := range sub.Sessions {
		st.Sessions = append(st.Sessions, provisioning.Session{
			System: s.System, IPv4Address: s.IPv4Address, IPv6Prefix: s.IPv6Prefix, DataNetwork: s.DataNetwork,
		})
	}

	return st, nil
}

func (a *Adapter) CreateFramedRoute(ctx context.Context, r provisioning.FramedRoute) error {
	if err := a.c.CreateDataNetworkFramedRoute(ctx, r.DataNetwork, &client.CreateFramedRouteOptions{IMSI: r.IMSI, IPv4: r.IPv4, IPv6: r.IPv6}); err != nil {
		return fmt.Errorf("create framed routes for %s on %s: %w", r.IMSI, r.DataNetwork, err)
	}

	return nil
}

func (a *Adapter) DeleteFramedRoute(ctx context.Context, dataNetwork, imsi string) error {
	return wrap("delete framed routes for "+imsi+" on "+dataNetwork, a.c.DeleteDataNetworkFramedRoute(ctx, dataNetwork, imsi))
}

func wrap(what string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", what, err)
}
