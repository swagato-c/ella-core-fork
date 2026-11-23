# Ella Core tester, flexicore fork

This branch (`flexicore` in [swagato-c/ella-core-fork](https://github.com/swagato-c/ella-core-fork))
is a fork of [Ella Core](https://github.com/ellanetworks/core) kept as the
Primary Driver of [flexicore](https://github.com/swagato-c/flexicore)'s
acceptance harness (flexicore ADR 0020). The module path stays
`github.com/ellanetworks/core`, so imports never change; flexicore's
`acceptance/go.mod` pins this fork with `replace` directives at a fork commit.

## Licence: BUSL-1.1, non-production use only

Everything here, the fork's own commits included, is Ella Core's code or
derived from it and is licensed under the **Business Source License 1.1**
(see `LICENSE`; Licensor: Ella Networks Inc.; Additional Use Grant: none).
BUSL-1.1 permits only **non-production use**. flexicore runs this tester
against cores in development, in CI and on staging Stamps, never against a
live network serving Subscribers. No file from this fork is copied into
flexicore's Apache-2.0 repository: the harness imports it as a Go module
dependency.

## Sources taken

| What | From | At |
| --- | --- | --- |
| The whole repository, the tester (`tester/`, formerly `internal/tester/`) and its codecs (`nas`, `ngap`, `s1ap`, `sctp`, `per`, `lppa`, `nrppa`, `diameter`) | `ellanetworks/core`, BUSL-1.1 | `142edd315a1453fe72c45b6fd508ccf763651035` (2026-11-21) |
| Milenage (`internal/milenage`), used directly by the 4G UE | the same repository | the same commit |
| The Ella Core adapter's fixture logic (`tester/provisioning/ellacore`) | `integration/fixture` of the same repository | the same commit |

Nothing else is taken from other projects.

## What the fork changes

The fork's commits sit on top of the upstream pin, bottom first:

1. **Move `internal/tester` to `tester`**, so other modules can import it. Paths only.
2. **Milenage for the 4G UE.** `s1enb/ue.go` derived RES, CK, IK and AK
   through `internal/udm`, which pulls Ella Core's UDM and database layer
   into the tester. It calls `internal/milenage` instead.
3. **Provisioning port.** `tester/provisioning` is a port of plain-data
   operations (apply a fixture, change a data network, policy, profile or
   subscriber, read a subscriber's registrations and sessions, framed
   routes). Programs register an adapter with `provisioning.Use`; scenarios
   open it with `provisioning.Dial`. `tester/provisioning/ellacore` wraps
   Ella's original REST `client` and is the only tester code importing it.
   `cmd/core-tester` and `integration/fixture` use it. Scenarios for
   Ella-only surfaces (IMS, SMS settings, location) take the client from the
   adapter through `scenarios.EllaCoreClient`, which refuses any other
   core's adapter. flexicore's own adapter arrives with its provisioning
   interface.
4. **This README, `tester/Dockerfile` and `tester/rebase-onto-upstream.sh`.**

Fixes go upstream as pull requests too, but the fork never waits for them.

## Rebasing onto upstream

Only at stage boundaries:

```bash
tester/rebase-onto-upstream.sh <ella-core-commit>   # or no argument for upstream/main
go build ./tester/... ./cmd/core-tester && go test ./tester/...
# run Ella's scenarios (below), then
git push --force-with-lease origin flexicore
```

Then bump the pin in flexicore's `acceptance/go.mod` (`acceptance/scripts/pin-fork.sh <fork-commit>`).

## Running Ella's own scenarios through the port

`integration/` applies every scenario's fixture through the Ella Core
adapter. On a Linux Docker host with SCTP:

```bash
docker pull ghcr.io/ellanetworks/ella-core:sha-142edd315a14
docker tag ghcr.io/ellanetworks/ella-core:sha-142edd315a14 ella-core:latest
docker build -f tester/Dockerfile -t ella-core-tester:latest .
mkdir -p ui/dist && touch ui/dist/.keep        # the UI is not needed for tests
INTEGRATION=1 IP_VERSION=ipv4 ELLA_CORE_ATTACH_MODE=xdp-generic \
  go test -run 'TestIntegrationTester$' -timeout 120m -v ./integration/
```
