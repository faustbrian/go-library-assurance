# Golib Library Assurance

This repository owns maintained cross-library composition scenarios. It is a
non-releasable test harness, not an application runtime, library umbrella or
second tooling CLI. Ordinary library CI does not depend on its checks or release.

## Current scenario

[`scenarios/compatibility-consumer`](scenarios/compatibility-consumer) carries
the eleven existing Go test files and exact dependency selections from Tools
commit `47a291bfa957ed00e833b51cde867cc7b5bd831a`. Only the test module identity
changes to `github.com/faustbrian/go-library-assurance/scenarios/compatibility-consumer`.
Its Go minimum remains 1.27.0; qualification uses Go 1.27.2.

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
