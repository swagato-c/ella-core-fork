// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package provisioning is the port through which scenarios and test
// harnesses set and read subscriber state on the core under test.
//
// Scenarios never talk to a core's management API directly. They dial a
// Port, and the adapter registered with Use decides which core it speaks
// to: the ellacore adapter wraps Ella Core's REST client, and other cores
// bring their own adapter. Every type here is plain data, so no adapter's
// SDK leaks into a scenario.
package provisioning

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Port sets and reads the state a scenario needs on the core under test.
type Port interface {
	// Apply provisions every resource in f and returns a Teardown that
	// removes the scoped ones again, in reverse dependency order. Scoped
	// resources (profiles, slices, data networks, policies, subscribers,
	// static IPs, framed routes) are created strictly: a name already in
	// use is an error. The operator and home network keys are applied
	// idempotently and are not torn down.
	Apply(ctx context.Context, f Fixture) (Teardown, error)

	UpdateDataNetwork(ctx context.Context, dn DataNetwork) error
	UpdatePolicy(ctx context.Context, p Policy) error
	UpdateProfile(ctx context.Context, p Profile) error
	UpdateSubscriber(ctx context.Context, u SubscriberUpdate) error
	DeleteSubscriber(ctx context.Context, imsi string) error

	// Subscriber reads what the core currently knows about one
	// subscriber's registrations and sessions.
	Subscriber(ctx context.Context, imsi string) (SubscriberState, error)

	CreateFramedRoute(ctx context.Context, r FramedRoute) error
	DeleteFramedRoute(ctx context.Context, dataNetwork, imsi string) error
}

// Teardown undoes an Apply. Call it once.
type Teardown func(ctx context.Context) error

// Target says where a core's provisioning interface is and how to
// authenticate to it.
type Target struct {
	Address string
	Token   string
}

// Opener builds a Port for a Target.
type Opener func(ctx context.Context, t Target) (Port, error)

var (
	mu     sync.Mutex
	opener Opener
)

// ErrNoAdapter is returned by Dial when no adapter has been registered.
var ErrNoAdapter = errors.New("provisioning: no adapter registered (call provisioning.Use)")

// Use registers the adapter that Dial opens. Programs call it once at
// start-up: core-tester registers the Ella Core adapter, and a harness
// registers the adapter for the core it tests.
func Use(o Opener) {
	mu.Lock()
	defer mu.Unlock()

	opener = o
}

// Dial opens a Port on t with the registered adapter.
func Dial(ctx context.Context, t Target) (Port, error) {
	mu.Lock()
	o := opener
	mu.Unlock()

	if o == nil {
		return nil, ErrNoAdapter
	}

	p, err := o(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("provisioning: open %s: %w", t.Address, err)
	}

	return p, nil
}
