# Contributing

Use Python 3.9 or newer and Go 1.27.2. This checkout is standalone: no sibling
checkout, ecosystem goal ledger or source-built Tools CLI is required.

## Selection and bootstrap

```sh
PYTHONDONTWRITEBYTECODE=1 python3 tools/compatibility/check-selection.py
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools/compatibility -p test_selection.py
```

The adapter fetches the generic Python helper at immutable Tools source
`0e8ff6fee42e88858069a6f6ae6f1080d95c5eae`, verifying its exact SHA-256
before execution. It also verifies the v1.8.4 released compatibility-set and
consumer/engineering catalog assets against their frozen release digests.
Downloads have byte limits, a shared 90-second preparation deadline and
at-most-15-second network waits; a blocked read may outlast that deadline by
one network wait. Failed fetches or digest mismatches are fatal, without a
mutable-ref or cached fallback. Inputs remain in memory, not a mutable catalog.

The Tools functions validate the current selection and render its direct
requirements/import inventory. The adapter maps only the scenario module
identity and supplies the actual owner-relative generator path. Read-only checks
require exact renderer source bytes, including its generated header; handwritten
or stale headers are refused. Indirect requirements retain the generic matcher's
existing allowance. This check does **not** verify every selected
tag's current availability or certify successful scenario execution.

To refresh the generated import file deliberately:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 tools/compatibility/check-selection.py --write
PYTHONDONTWRITEBYTECODE=1 python3 tools/compatibility/check-selection.py
```

`--write` validates the same authenticated selection and existing direct
requirements before replacing only `consumer_test.go`. It never rewrites
`go.mod`, `go.sum`, selection or frozen catalogs. It requires bounded, regular,
single-link input files and real scenario directories; symlinks and unsafe
paths are refused. Keep the checkout stable during validation and writing:
these guards do not provide a concurrent filesystem transaction.

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
applicable work. Source documentation is always mandatory: its separate CI job
uses the pinned setup-golib action's explicit source bootstrap at Tools
`0e8ff6fee42e88858069a6f6ae6f1080d95c5eae`, then runs
`golib source-docs check --json`. Git-root discovery needs no `.golib.yaml` in
this test-only harness. The job always invokes guarded cleanup of its exact
owned source installation, including after a check failure; unexpected paths
or aliased/nonregular contents fail closed rather than deleting foreign files.
Generic selection and execution consume the immutable Tools helper source named
in the workflow; they are not copied into this repository.

The scenario module is not released or imported by ordinary libraries.
Library-local test/security/release obligations remain with their owners;
Tools retains generic verification and historical catalog/security authority.

### Declared minimum compiler

CI's optional `minimum_compiler` manual-dispatch input defaults to false. When
selected, all three native platforms also use exact Go 1.27.0 to test the entire
current selected public module, including its import inventory and composition
contracts. Required rejects failed, cancelled, skipped or missing selected
minimum-compiler work; when unselected, that work must remain skipped. The
ordinary Go 1.27.2 qualification is unchanged.

This bounded scenario uses direct Go commands rather than the shared Tools
runner, whose current qualification contract requires Go 1.27.2 and tidies a
disposable copy. Minimum-compiler qualification instead disables automatic
toolchain/workspace selection, requires public authenticated module resolution,
rejects replacements and runs readonly tests with private cleaned caches. It
checks the declared 1.27.0 floor and unchanged checked-in `go.mod`/`go.sum`; it
does not tidy, repin or raise the floor to obtain a pass. Failures remain evidence
against that selected compiler/platform claim. It does not qualify every patch
compiler or a separately minimized dependency graph.

## Fixed release-integrity recipe

The default-off manual CI input `release_integrity` selects one serial Ubuntu
24.04/Go 1.27.2 job. It authenticates the public Tools v2.5.1 checksums, manifest
and four distribution archives through GitHub attestations bound to its release
workflow, tag and source `09b43f49385650c44dd4151a12b0877ca6b48a88`. Independent
builds use that exact source and release flags, then its existing `release package`
and `release archive-check` commands compare both binary and archive bytes for
Linux/Darwin amd64/arm64. These are cross-builds on Linux amd64, not native
execution on all four targets or a universal cross-environment promise.

The same recipe produces retrospective CycloneDX 1.6 SBOMs and module inventories
for JSONAPI/v2 v2.0.1 at `6f44f8b7cd089d36bc36e67e6aec38faa5ad1aa0` and
Localized/v5 v5.0.2 at `534315549084994f3d2a2e088400ddc2bf3a6b68`, using public
authenticated modules and `cyclonedx-gomod@v1.12.0`. Expected release identity is
recorded separately from the generator's untouched root-version autodetection.
The evidence is not an original release-pipeline SBOM or an assertion of SBOM
byte reproducibility. No release is created or amended.

Readonly module execution rejects replacement graphs, checks Origin/sums and
preserves source `go.mod`/`go.sum` hashes without tidying. Private caches and
downloads are removed; the 30-day `fixed-release-integrity` artifact retains
environment, authentication results, module graphs, SBOMs, source/root metadata,
digests, comparisons and exit status, including partial failure evidence.
Required fails for any selected failed, cancelled, skipped or missing job.
Publication and deployment are outside this recipe.

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
