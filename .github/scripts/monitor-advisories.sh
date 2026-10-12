#!/usr/bin/env bash
# Assurance-owned selected-graph orchestration, not a generic security verifier.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
task="$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/assurance-advisories.XXXXXX")"
# This invocation owns the whole directory. Cleanup failure changes the final
# verdict so a successful scanner cannot hide retained downloads or cache state.
cleanup() {
  local status=$?
  trap - EXIT
  chmod -R u+w "${task}" || status=1
  find "${task}" -depth -delete || status=1
  exit "${status}"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
mkdir -m 700 -p "${task}/cache" "${task}/mod" "${task}/tmp"
export GOCACHE="${task}/cache" GOMODCACHE="${task}/mod" GOTMPDIR="${task}/tmp"
export GOENV=off GOWORK=off GOTOOLCHAIN=local GOMAXPROCS=2
export GOFLAGS='-mod=readonly -p=1'
export GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org
export GOPRIVATE='' GONOPROXY='' GONOSUMDB=''
export GOVULNDB=https://vuln.go.dev
cd "${root}/scenarios/compatibility-consumer"
go mod download
go mod verify
# Module-granularity advisories use the current public database. They do not
# execute compositions or determine which vulnerable symbols are reachable.
# Version aligned with immutable Tools 75aeb45cf4fcedd0227782ea08bb9d690039b762
# internal/gates/runner.go. The advisory DB intentionally remains current.
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 -mode=source -scan=module -db=https://vuln.go.dev
