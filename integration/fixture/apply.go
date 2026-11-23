// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

package fixture

import (
	"github.com/ellanetworks/core/tester/provisioning/ellacore"
	"github.com/ellanetworks/core/tester/scenarios"
)

// Apply provisions every scenario-owned resource in spec through the
// provisioning port's Ella Core adapter, and registers
// t.Cleanup handlers to delete them when the subtest ends. LIFO cleanup
// order gives the correct reverse-dependency teardown: subscribers,
// policies, data networks, slices, profiles.
//
// Scoped resources use strict-create semantics — a name collision fails
// the test so that stray state from a prior run surfaces immediately
// instead of silently mutating shared state.
//
// Non-scoped resources (operator singleton, home network keys) are
// applied idempotently and are not cleaned up; they belong to the
// baseline shared across all subtests.
//
// Rule for scenario authors: reference baseline names (default
// profile/slice/DN/policy) only when assertions match baseline values.
// Anything else must be declared under a scoped name.
func (f *F) Apply(spec scenarios.FixtureSpec) {
	f.t.Helper()

	teardown, err := ellacore.New(f.c).Apply(f.ctx, spec.Provisioning())
	if err != nil {
		f.fatalf("apply fixture through the provisioning port: %v", err)
	}

	f.t.Cleanup(func() {
		if err := teardown(f.ctx); err != nil {
			f.t.Logf("cleanup: %v", err)
		}
	})
}
