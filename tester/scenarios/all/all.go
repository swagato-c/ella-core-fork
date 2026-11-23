// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package all blank-imports every scenario subpackage so their init()
// registrations execute when a binary imports this package for side effects.
package all

import (
	_ "github.com/ellanetworks/core/tester/scenarios/enb"
	_ "github.com/ellanetworks/core/tester/scenarios/gnb"
	_ "github.com/ellanetworks/core/tester/scenarios/ha"
	_ "github.com/ellanetworks/core/tester/scenarios/interworking"
	_ "github.com/ellanetworks/core/tester/scenarios/multi"
	_ "github.com/ellanetworks/core/tester/scenarios/s1enb"
	_ "github.com/ellanetworks/core/tester/scenarios/sms"
)
