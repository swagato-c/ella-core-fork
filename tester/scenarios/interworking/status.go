// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

package interworking

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ellanetworks/core/tester/provisioning"
	"github.com/ellanetworks/core/tester/scenarios"
)

const statusPoll = 200 * time.Millisecond

func coreClient(env scenarios.Env) (provisioning.Port, error) {
	cl, err := provisioning.Dial(context.Background(), provisioning.Target{Address: env.APIAddress, Token: env.APIToken})
	if err != nil {
		return nil, fmt.Errorf("open provisioning port: %w", err)
	}

	return cl, nil
}

func assertRegisteredOn(ctx context.Context, env scenarios.Env, want string) error {
	cl, err := coreClient(env)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(sessionSettle)

	last := make([]string, 0, 2)

	for {
		registered := false

		sub, err := cl.Subscriber(ctx, interworkingIMSI)
		if err == nil {
			last = last[:0]

			for _, reg := range sub.Registrations {
				if reg.ConnectionState == nil {
					continue
				}

				last = append(last, reg.System)
				registered = registered || reg.Registered
			}

			if registered && slices.Equal(last, []string{want}) {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("the subscriber is not registered on %s alone (systems with a context %v)", want, last)
		}

		time.Sleep(statusPoll)
	}
}
