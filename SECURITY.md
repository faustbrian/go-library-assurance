# Security and supply-chain ownership

Report suspected harness, dependency, action or bootstrap compromise privately
through [Tools security advisories](https://github.com/faustbrian/go-library-tools/security/advisories/new).
Tools owns ecosystem security policy and coordinated response; Assurance owns
its selection, composition assertions and CI configuration. Never publish
credentials or exploit-enabling details in issues or job artifacts.

Use the shared [security policy](https://github.com/faustbrian/go-library-tools/blob/main/SECURITY.md)
and [vulnerability-management process](https://github.com/faustbrian/go-library-tools/blob/main/docs/ecosystem/security/vulnerability-management.md)
for severity, triage, embargo, advisory ranges and affected-module remediation.
A harness check is not a release verdict for the selected libraries.
Dependency governance and compromise handling remain in the shared
[supply-chain policy](https://github.com/faustbrian/go-library-tools/blob/main/docs/supply-chain.md)
and [compromise playbook](https://github.com/faustbrian/go-library-tools/blob/main/docs/ecosystem/security/vulnerability-management.md#supply-chain-compromise-response).

## Authoritative dependency inventory

- `scenarios/compatibility-consumer/selection.json` owns current direct public
  selections and source revisions. Its generated requirements and imports are
  checked against immutable Tools policy/catalog inputs by
  `tools/compatibility/check-selection.py`; that file also inventories each
  fetched helper/fixture asset URL, revision and SHA-256. Frozen catalog license,
  purpose and provenance metadata remain Tools-owned historical evidence.
- Scenario `go.mod` records direct and indirect test dependencies; `go.sum`
  authenticates their resolved contents through public SumDB. There is no root
  module, vendor tree or independent runtime dependency graph.
- `.github/workflows/*.yml` inventory SHA-pinned checkout, setup-go and dependency
  review actions and the SHA-256-pinned Gitleaks release archive. Review upstream
  provenance and permissions when updating an action, not just its version.
- `.go-version` owns the qualification toolchain. The selected-graph monitor
  invokes `golang.org/x/vuln/cmd/govulncheck@v1.6.0`, aligned with Tools source
  `75aeb45cf4fcedd0227782ea08bb9d690039b762` (`internal/gates/runner.go`).
  Python 3.9+ and Bash run orchestration without additional package dependencies.

New direct dependencies require a reviewed selection change documenting purpose,
license, maintenance/security history and replacement cost; prefer the standard
library and preserve optional-integration isolation. Catalog metadata does not
establish current upstream maintenance, retraction or advisory status: inspect
those when selecting new versions and triaging updater/monitor findings.

## Updates and current advisories

`.github/dependabot.yml` is the only updater configuration. It proposes weekly
action updates and indirect-only scenario Go updates. Direct versions cannot be
updated by a go.mod-only bot PR: maintainers change `selection.json` together
with generated direct requirements/imports and affected composition proof.
Indirect updates retain the current matcher's allowance but must pass checksum,
selection, dependency review and composition CI. Review retractions, unexpected
module replacements and maintenance status before accepting an update.

The separate **Selected dependency advisories** workflow runs weekly and can be
dispatched on main. It verifies the selected module graph and checks the current
official Go vulnerability database with pinned govulncheck at module granularity;
it does not run composition tests or establish symbol reachability. Findings,
network/tool failures and timeouts fail its **Selected graph advisories** job.
It is not an ordinary PR-required status and cannot turn a prior composition
verdict into proof against newly published advisories. Maintainers monitor failed
runs and triage affected modules through the shared private process; no automatic
issue publication or privileged credential is needed.

## Suspected compromise

Stop accepting affected inputs and publication, retain immutable source/tag,
checksum and workflow-run evidence, and privately notify the security owner.
Identify affected selected versions and consumers before replacing the input.
Credential revocation/rotation, advisory publication, withdrawal/retraction and
communication require the responsible owner's authority; never delete tags or
rewrite history to hide evidence. Assurance publishes no module or binary, so
any affected library's owner publishes its fix and consumer upgrade guidance.
