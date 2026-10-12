# Source ownership and audit

## Baseline inventory

The tracked-source audit at main
`85735f1dfebc9a2f73a16bcb187415d9558d09fb` found one module,
`scenarios/compatibility-consumer`, and eleven Go files, all in the external
`compatibilityconsumer_test` namespace. There is no runtime Go package, public
library API, root module, vendor source, build constraint, platform-specific
file, example, fuzz target, benchmark or configured API documentation snapshot.
Test names and private collaborator methods are not a separately published API.

Reproduce the source boundary with:

```sh
git ls-files '*.go' 'go.mod' '**/go.mod'
```

| Source | Responsibility and audit focus |
| --- | --- |
| `composition_test.go` | Nominal types, canonical/retained adapters, immutable inputs, finite limits, issuer-bound reusable grants, idempotent replay and explicit logging trust. |
| `consumer_test.go` | Generator-owned blank imports for selected public packages; not behavioral evidence for every imported adapter. |
| `openapi2_composition_test.go` | Exact numeric-token retention, OAS 3.1 dialect bridge, bounded parsing and authorization admission before a decision. |
| `outbox2_composition_test.go` | Routing/command identity, copied delivery input and disagreement between envelope metadata and encoded event identity. |
| `postgres2_composition_test.go` | Explicit resolution, private error causes, cancellation before collaborator calls and caller-owned pool lifetime. |
| `qcp3_composition_test.go` | Static authentication/authorization composition, anonymous API access, tenant isolation and actor binding without real control-plane operations. |
| `state_machine2_composition_test.go` | Whole-plan admission before effects/records, finite compile allowance, default diagnostic privacy and caller-provided database dependencies. |
| `tabular_xlsx_composition_test.go` | One in-memory spreadsheet corpus across supported historical/current majors, with reader cleanup on every subtest outcome. |
| `wire3_composition_test.go` | Canonical/retained codec identity, copy ownership, present-empty locales, typed format/size failures and schedule endpoint semantics. |
| `workflow_composition_test.go` | Portable event fields versus workflow-owned side state, explicit reverse-conversion loss, copied payloads/steps and one deadline-bound activity. |
| `wsdl2_composition_test.go` | WSDL 1.1 model round trip, parser/compiler cancellation, compiler reuse and input-independent default errors. |

Three private helpers are deliberately narrow: the idempotency clock freezes
lease time, the borrowed Postgres resource counts lifecycle calls without a
database, and the state-machine collaborator counts synchronous effects and
records without durable storage. They do not establish concurrent-use safety,
crash recovery or external-service interoperability.

## Generated input and qualification boundaries

`consumer_test.go` is the exact import rendering of the pinned Tools helper
selected by [`check-selection.py`](tools/compatibility/check-selection.py).
`selection.json` owns current choices; the helper owns rendering. The adapter
validates the generated direct requirements and imports without rewriting them.
Its existing direct `test_source != imports` comparison intentionally refuses
even an extra header comment. The file currently has no generated-code header
or local `go:generate` directive; adding one requires a reviewed producer change,
not a hand edit or a relaxation of this integrity check. The generator identity,
source digest and frozen catalog digests are recorded in that adapter.

For selection changes, use the pinned helper's `current_consumer_inputs` and
`render_clean_consumer` functions through the adapter's verified `load_inputs`
boundary; retain the adapter's exact Tools-to-Assurance module-header mapping.
The maintained [selection check](CONTRIBUTING.md#selection-and-bootstrap) is
read-only, not a regeneration command. Do not call the helper's unrelated
default main path or duplicate its policy here.

The local bootstrap helper downloads only bounded, digest-verified public
inputs before execution. A shared preparation deadline and bounded network
waits are not a guarantee of instantaneous cancellation. File admission assumes
a stable checkout; pre-open checks do not create an atomic filesystem snapshot.

The CI selector and scenario runner remain immutable Tools-owned helpers.
[`check-required.sh`](.github/scripts/check-required.sh) treats selected scenarios
as mandatory and inapplicable scenarios as skipped; invalid classification,
missing mandatory work, cancellation and failure cannot become success.
The selected-graph advisory monitor owns its caches and reads the current
vulnerability database without composing libraries or proving symbol reachability.
Its process-double tests are orchestration proof, not a live vulnerability verdict.

Source comments preserve all assertions, literals, dependency selections and
executable template bytes. These tests qualify the exercised compositions,
not every library contract or platform. No module release or runtime-library
changelog is applicable to this non-releasable harness.

## Outstanding producer adoption

This harness intentionally has no `.golib.yaml`. The Tools source-docs route
must use Git-root discovery rather than requiring library framework metadata.
The candidate Git-only correction inventories this module's eleven tracked Go
files and passes its objective strict check without configuration scaffolding;
that local result is not final Tools CI, release or harness-CI adoption proof.

The producer-owned generated-header correction remains pending. Update this
harness only through the verified renderer and its exact native guard once
that producer change is delivered. A structurally passing test-source inventory
does not discover this absent header's producer ownership, and maintained-file
prose/token proof does not fulfill that dependency or qualify the complete goal.
