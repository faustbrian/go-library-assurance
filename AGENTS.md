# Assurance Ownership

The Golib coordination instructions and their streamlined workflow apply.

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
