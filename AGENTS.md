# Engineering Policy

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and
"OPTIONAL" in this document are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals, as
shown here.

This file is the canonical, self-contained repository policy. `CLAUDE.md` and
tool-specific instructions MUST point here. Historical plans remain product
requirements/evidence, not an alternate execution workflow or proof of completion.

## Shared Repository Development Workflow

This self-contained workflow is identical in every repository's `AGENTS.md`.
Repository additions MAY define concrete domain contracts and named risks, not
another lifecycle, gate sequence, review count or evidence framework. This
user-approved Golib workflow replaces conflicting inherited procedural defaults,
not correctness, security or mandatory CI. Explicit user authority and higher-
priority system/developer instructions remain controlling.

### Scope and verification

- Before writing, identify the outcome, owner, affected contracts/direct
  consumers, named risks and smallest sufficient checks in one compact plan.
  Preserve existing work; separate role factories, signed lifecycle receipts
  and recursive approval chains MUST NOT be routine stages.
- Classify the semantic change, not its filename or line count:
  - **Tier A:** non-runtime prose/metadata/registration: affected structure,
    links, examples/generation and final-diff inspection.
  - **Tier B:** internal behavior: focused regression, affected package/module
    tests, applicable format/static checks and one complete review.
  - **Tier C:** public API, lifecycle, security, persistence or concurrency:
    regression/characterization, affected integration/API/direct-consumer
    proof and one independent complete-final-diff review.
  - **Tier D:** release/milestone: relevant immutable-input compatibility,
    composition, public-consumer and release-integrity proof.
  Metadata, dependency, workflow and policy edits MUST retain their actual
  security/trust checks; names/extensions MUST NOT conceal executable changes.
  Actions and external tools MUST use immutable pinned identities. Preserve
  authenticated inputs, least privilege and untrusted-PR/release separation.
- Behavioral fixes MUST demonstrate the regression and corrected behavior.
  Tests assert outcomes, invariants, failures and cleanup, not just line counts.
  Behavioral PRs use affected tests, fuzz seeds/saved crashes and applicable
  static/security checks; changed hostile inputs add bounded active fuzzing.
  Cheap changed-content secret checks remain applicable to every change.
- Mutation campaigns MUST NOT run on PRs or become required PR statuses.
  Optional local mutation and bounded selected main/scheduled assurance remain
  separate. Race/leak, fuzz, service, conformance and performance checks require
  a named risk/release boundary. Material concurrency/lifetime changes MUST
  retain targeted race/leak proof; conformance claims MUST use applicable pinned
  official fixtures and independent oracles; performance claims MUST compare
  equivalent workloads with reproducible environment/corpus evidence.
  Selected timeouts, partial results and viable survivors are not passing
  proof; equivalent/invalid mutants need narrow independently reviewed proof.
  Coverage scores MUST NOT replace behavioral adequacy or waive required gates.

### Commands, review and evidence

- Choose actual commands from `CONTRIBUTING.md`, the Makefile and CI; run native
  focused tests in the affected module. Use the bounded local target where
  provided. `make check`/`make ci` MUST NOT be assumed cheap or made universal
  full qualification. Local/CI executions of the same selected gate preserve
  its scripts/thresholds. Missing prerequisites fail only an applicable gate.
  NilAway remains advisory with visible findings; mandatory failures stay fatal.
- One independent complete-final-diff review is the default for meaningful
  Tier C/D work; more reviews require a distinct named high-risk domain. Fix
  material findings, rerun affected checks and review the changed final diff,
  without restarting planning for clerical changes.
- Reuse evidence only when material source, dependency, fixture, tool and
  environment inputs still prove the same claim. Retain original identity and
  result; ambiguous/forged/partial evidence is not proof. A new SHA or mutable
  prose MUST NOT force broad reruns or provenance chains. Keep real integrity
  checksums, signatures, source locks and atomic publication at their actual
  supply-chain/reuse boundaries.
- Required aggregation MUST reject failed, cancelled, missing or skipped
  applicable work and classifier failures. Inapplicable work is not a passed
  check. Protection, scanner and dependency-authentication bypasses are forbidden.

### Delivery and resources

- Work/releases stay on main using the pinned Go 1.27 toolchain, required major
  module/import suffixes and Git tags, not versioned source branches/directories.
  CI-only protected-main carriers require applicable authority and advance the
  exact green reviewed source. Preserve user work, stage explicit owned paths
  and honor push/publication/destructive safeguards; carry approvals within scope.
  No routine amend, force-push, history rewrite or hook bypass is permitted.
- A changed repository is handled only on remote main with applicable exact-
  source CI green. Required releases also need published tags and actual clean
  public consumers. Pure internal policy/prose without consumer-visible contract
  impact MUST NOT force changelogs, repins or releases. Material consumer changes
  require the affected changelog before publication; preserve Unreleased entries.
- Agent Go runs MUST use task-owned `GOCACHE`, `GOMODCACHE` and `GOTMPDIR`,
  reused within a bounded local session rather than rebuilt per adjacent command.
  Clean exact owned caches/resources on success, failure and interruption;
  never remove user/shared resources. Serialize local heavy runs and honor
  hosted-only system-risk restrictions.
- Distinguish local, committed, pushed, CI-qualified and released states. Report
  failures, skipped checks and unverified boundaries; reused proof MUST NOT be
  described as newly executed.

## Assurance Ownership

- This repository owns current cross-library composition scenarios, selection
  data and their qualification jobs. It is a non-releasable harness, not a
  runtime library or second verifier product.
- Preserve scenario assertions, supported historical majors, exact public
  dependency selections and Go module minima. Qualification must use public
  resolution with `GOWORK=off`, no `replace` directives, and task-owned caches.
- Generic policy, validators, schemas and reusable execution helpers remain in
  Tools. Consume immutable verified released/source inputs; do not copy the
  compatibility generator or maintain a second mutable ecosystem catalog.
- `scenarios/compatibility-consumer/selection.json` is the sole editable current
  selection. Frozen Tools catalog assets are historical inputs, not evidence
  that the current selection has already passed.
- Ordinary libraries must not import this harness or wait for its release/CI.
  No automatic Tools-to-Assurance-to-library dispatch chain is permitted.
- Use focused selection/refusal checks and the affected composition tests.
  Keep qualification outcomes honest across Darwin arm64, Linux amd64 and
  Linux arm64; unavailable/failed platforms are not passing proof.
- Keep disposable downloads, Go caches and staged modules task-owned and remove
  them on success, failure and cancellation. Process-control/cancellation
  fixtures belong to capable hosted qualification, not local ad hoc execution.
- Preserve historical Tools assets, security authority and documentation links.
  New catalog-publication ownership requires a separately accepted scope.

[RFC2119]: https://www.rfc-editor.org/rfc/rfc2119
[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174
