// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

package scenarios

import (
	"context"

	"github.com/ellanetworks/core/client"
	"github.com/ellanetworks/core/tester/provisioning"
	"github.com/ellanetworks/core/tester/provisioning/ellacore"
)

// EllaCoreClient dials the provisioning port and returns the Ella Core
// client behind it. Only scenarios that exercise Ella-Core-only surfaces
// (IMS, SMS settings, location) use it; everything else goes through
// provisioning.Port. It fails when the registered adapter is not Ella
// Core's.
func EllaCoreClient(ctx context.Context, address, token string) (*client.Client, error) {
	p, err := provisioning.Dial(ctx, provisioning.Target{Address: address, Token: token})
	if err != nil {
		return nil, err
	}

	return ellacore.Client(p)
}
