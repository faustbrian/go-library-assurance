# Contributing

Use Python 3.9 or newer and Go 1.27.2. This checkout is standalone: no sibling
checkout, ecosystem goal ledger or source-built Tools CLI is required.

## Selection and bootstrap

```sh
PYTHONDONTWRITEBYTECODE=1 python3 tools/compatibility/check-selection.py
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools/compatibility -p test_selection.py
```

The adapter fetches the generic Python helper at Tools v2.2.0 source
`f0dd7cc2ed25abdb513b6364d7e160bf4b37c21b`, verifying its exact SHA-256
before execution. It also verifies the v1.8.4 released compatibility-set and
consumer/engineering catalog assets against their frozen release digests.
Downloads have byte limits, a shared 90-second preparation deadline and
at-most-15-second network waits; a blocked read may outlast that deadline by
one network wait. Failed fetches or digest mismatches are fatal, without a
mutable-ref or cached fallback. Inputs remain in memory, not a mutable catalog.

The existing Tools functions validate the current selection and render its
direct requirements/import inventory. The adapter changes only the exact
generated module header for this non-releasable scenario and checks the owned
files without rewriting them. Indirect requirements retain the generic
matcher's existing allowance. This check does **not** verify every selected
tag's current availability or certify successful scenario execution.

Only `scenarios/compatibility-consumer/selection.json` owns current selection;
update it deliberately with reviewed public versions and exact source revisions.
Preserve frozen base catalogs and historical verification identities. Changes
to generated requirements/imports still need normal Go dependency consistency
and composition qualification; no Tools release or library pin migration follows
merely from editing this harness.

## Go scenarios

Ordinary Go commands work inside the scenario module:

```sh
cd scenarios/compatibility-consumer
go test ./...
```

For agent or bounded qualification work, use private caches and authenticated
public dependencies instead of ambient workspaces/caches:

```sh
bash <<'SH'
set -euo pipefail
task="$(mktemp -d "${TMPDIR:-/tmp}/golib-assurance-local.XXXXXX")"
cleanup() {
  local status=$?
  trap - EXIT
  chmod -R u+w "$task" || status=1
  find "$task" -depth -delete || status=1
  exit "$status"
}
trap cleanup EXIT
mkdir -p "$task/cache" "$task/mod" "$task/tmp"
export GOCACHE="$task/cache" GOMODCACHE="$task/mod" GOTMPDIR="$task/tmp"
export GOENV=off GOWORK=off GOTOOLCHAIN=local GOMAXPROCS=2 GOFLAGS=-p=1
export GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org
export GOPRIVATE='' GONOPROXY='' GONOSUMDB=''
cd scenarios/compatibility-consumer
go test -mod=readonly -count=1 -timeout=20m ./...
SH
```

This module preserves compilation/import coverage and the existing eleven
composition test files; no new mutation/fuzz campaign or blanket race gate is
introduced. Native platform qualification remains Darwin arm64, Linux amd64
and Linux arm64 with fail-closed results and owned resource cleanup. A local
pass covers only the executed platform and tests, not that complete matrix.

PR and main CI select only changed independent scenario files. Dependency,
selection, toolchain, workflow and adapter changes select the complete matrix;
manual dispatch explicitly selects full qualification. Unrelated prose does
not rebuild compositions. Cheap selection/refusal checks and changed-content
secret scanning still run. The scanner is the same standard Gitleaks v8.30.1
used by Tools, fetched as an exact SHA256-pinned release binary rather than
compiled for every change. Required refuses failed, cancelled or missing
applicable work. Generic selection and execution consume the immutable Tools
helper source named in the workflow; they are not copied into this repository.

The scenario module is not released or imported by ordinary libraries.
Library-local test/security/release obligations remain with their owners;
Tools retains generic verification and historical catalog/security authority.

## Dependency maintenance

The single [Dependabot configuration](.github/dependabot.yml) updates actions and
only indirect scenario modules. Direct updates must change `selection.json`
alongside the generated direct `go.mod` requirements/imports; a go.mod-only
direct bump is intentionally rejected. Review versions, source identity,
retractions and module provenance before accepting either kind of update.

The separate scheduled/manual [advisory monitor](.github/workflows/advisories.yml)
refreshes the selected Go graph against current advisories without running
compositions. Its shell boundary is covered by process-double tests (not live
Go vulnerability evidence):

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s .github/scripts -p 'test_*.py'
actionlint
```

See [SECURITY.md](SECURITY.md) for scanner ownership, inventory sources and
compromise response. This harness remains non-releasable; these maintenance
changes do not require module tags, changelogs or downstream library repins.
