# Golib Library Assurance

## Harden composition scope

The Harden additions exercise the existing public selections, without repinning
libraries or changing their APIs. Acceptance is actual loopback JSON-RPC/OpenRPC
discovery and JSON:API media/cursor/error responses; Wire JSON response decoding;
finite Scheduler/Clock/Calendar/Opening Hours/Temporal gap and fold composition;
and caller-driven validated configuration refresh through Wire JSON and the
Filesystem `io/fs` bridge, retaining the last good value after rejection.
JSON Schema validation is explicitly application-owned. RPC validation reuses
the Draft 7 schema retained in its OpenRPC metadata; configuration validates its
own Draft 7 policy before candidate publication. This is not an OpenAPI bridge
or automatic OpenRPC enforcement.
The selected ordinary JSON-RPC client reserves `rpc.*`; discovery therefore
uses its public raw HTTP transport and validates the response envelope/ID,
while application methods use the typed client.
DST retains the exact New York TZif from the Go 1.27.2 archive as an
independent pinned oracle. Actual canonical sibling loaders are preloaded and their finite
transition coordinates, due instants and point membership are compared with
that oracle. This proves outcome agreement, not byte identity of runtime-loaded
host data; test order requires no environment change or cache reset.
The oracle comes from Go source commit
`022c8636110ebac86a9b88724cda168ed713c21f`: `lib/time/zoneinfo.zip`
SHA-256 `b2d18a7c8fa8142097a48c99609fb3c92db5ee98bc740294e57eab8ae9f94779`,
with the 1744-byte New York entry SHA-256
`d7f2206b3a45989fc9ad63d558922532fa7352280d5f87176bf1db79cb1d1fa9`.

Named risks are schema dialect/nominal-major mismatch, loss of protocol error
status or discovery object shape, repeated civil times mistaken for identical
instants, invalid candidate publication, secret diagnostics and leaked owned
HTTP/filesystem resources. Minimal qualification is the new independent scenario
files, existing selection/refusal/source-documentation checks, and one complete
independent Tier D review. Each executed platform is reported separately.
`go-config/filesystem` is the format-source subpackage; sibling `go-filesystem/v2`
owns the actual storage/`io/fs` bridge. No implicit refresh watcher, forced
callback cancellation, durable storage, general DST expansion, service deployment,
whole conformance campaign or universal performance claim follows from these
finite scenarios.

This repository owns maintained cross-library composition scenarios. It is a
non-releasable test harness, not an application runtime, library umbrella or
second tooling CLI. Ordinary library CI does not depend on its checks or release.

## Current scenario

[`scenarios/compatibility-consumer`](scenarios/compatibility-consumer) carries
the eleven Go test files originally relocated from Tools commit
`47a291bfa957ed00e833b51cde867cc7b5bd831a`. The scenario module identity is
`github.com/faustbrian/go-library-assurance/scenarios/compatibility-consumer`;
current selections and dependency maintenance now belong to this repository.
Its Go minimum remains 1.27.0; ordinary qualification uses Go 1.27.2.
Manual CI dispatch can additionally select `minimum_compiler` to qualify the
unchanged current public dependency graph on native Go 1.27.0 across Darwin
arm64, Linux amd64 and Linux arm64. This is not a promise for every Go patch or
for independently minimized dependency versions; unexecuted platforms remain
unverified.

[`selection.json`](scenarios/compatibility-consumer/selection.json) is the sole
editable current selection. Historical compatibility sets and catalogs remain
immutable Tools release inputs, not a second maintained catalog here. Generic
verification, schemas, CLI, reusable actions and library workflows remain Tools.
Shared security authority and historical catalog publication also remain Tools.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the public-input bootstrap and
qualification boundary. No harness version, module release or library repinning
is required by this ownership change.

See [SECURITY.md](SECURITY.md) for dependency/action inventories, selection-owned
updates, scheduled advisory monitoring and shared incident-response ownership.

See [SOURCE_DOCUMENTATION.md](SOURCE_DOCUMENTATION.md) for source ownership,
composition proof limits and the generator-owned import inventory.

## Ecosystem navigation

See the [versioned Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.8.4/docs/ecosystem/README.md) for package
selection, explicit ownership, and compatibility guidance.
