#!/usr/bin/env bash
# Fixed retrospective release assurance; publication and generic tooling stay with owners.
set -euo pipefail
evidence="${RUNNER_TEMP:?}/assurance-release-integrity-evidence"
mkdir "${evidence}"
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then printf 'evidence=%s\n' "${evidence}" >>"${GITHUB_OUTPUT}"; fi
task="$(mktemp -d "${RUNNER_TEMP}/assurance-release-integrity.XXXXXX")"
cleanup() {
  local status=$?
  trap - EXIT
  chmod -R u+w "${task}" || status=1
  find "${task}" -depth -delete || status=1
  printf '%s\n' "${status}" >"${evidence}/exit-status.txt"
  exit "${status}"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
mkdir -p "${task}/cache" "${task}/mod" "${task}/tmp" "${task}/public" "${task}/build"
export GOCACHE="${task}/cache" GOMODCACHE="${task}/mod" GOTMPDIR="${task}/tmp"
export GOENV=off GOWORK=off GOTOOLCHAIN=local GOMAXPROCS=2 GOFLAGS='-mod=readonly -p=1'
export GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org
export GOPRIVATE='' GONOPROXY='' GONOSUMDB=''
export CGO_ENABLED=0 GOAMD64=v1 GOARM64=v8.0 GOEXPERIMENT=''
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
cd "${task}"
go env -json GOVERSION GOHOSTOS GOHOSTARCH GOOS GOARCH CGO_ENABLED GOAMD64 GOARM64 GOEXPERIMENT \
  GOFLAGS GOENV GOWORK GOTOOLCHAIN GOPROXY GOSUMDB GOPRIVATE GONOPROXY GONOSUMDB >"${evidence}/environment.json"
jq -e '.GOVERSION == "go1.27.2" and .GOHOSTOS == "linux" and .GOHOSTARCH == "amd64" and
  .GOOS == "linux" and .GOARCH == "amd64"' "${evidence}/environment.json" >/dev/null
gh --version >"${evidence}/github-cli.txt"
source_revision=09b43f49385650c44dd4151a12b0877ca6b48a88
repository=faustbrian/go-library-tools
platforms=(linux_amd64 linux_arm64 darwin_amd64 darwin_arm64)
gh release view v2.5.1 --repo "${repository}" --json tagName,isDraft,isPrerelease,assets >"${evidence}/tools-release.json"
jq -e '.tagName == "v2.5.1" and .isDraft == false and .isPrerelease == false' "${evidence}/tools-release.json" >/dev/null
assets=(checksums.txt release-manifest.json)
for platform in "${platforms[@]}"; do assets+=("golib_2.5.1_${platform}.tar.gz"); done
for asset in "${assets[@]}"; do
  gh release download v2.5.1 --repo "${repository}" --pattern "${asset}" --dir "${task}/public"
  gh attestation verify "${task}/public/${asset}" --repo "${repository}" \
    --signer-workflow "${repository}/.github/workflows/release.yml" \
    --source-digest "${source_revision}" --source-ref refs/tags/v2.5.1 \
    --deny-self-hosted-runners --format json >"${evidence}/${asset}.attestation.json"
done
cp "${task}/public/checksums.txt" "${task}/public/release-manifest.json" "${evidence}/"
jq -e --arg source "${source_revision}" '.schema_version == 1 and .version == "v2.5.1" and
  .source_commit == $source and .provenance == "github-artifact-attestation"' "${task}/public/release-manifest.json" >/dev/null
for asset in "${assets[@]:1}"; do
  digest="$(sha256sum "${task}/public/${asset}" | cut -d' ' -f1)"
  grep -Fx -- "${digest}  ${asset}" "${task}/public/checksums.txt" >/dev/null
  if [[ "${asset}" != release-manifest.json ]]; then
    jq -e --arg name "${asset}" --arg digest "${digest}" \
      '[.assets[] | select(.name == $name and .sha256 == $digest)] | length == 1' \
      "${task}/public/release-manifest.json" >/dev/null
  fi
done

generator=github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.12.0
go mod download -json github.com/CycloneDX/cyclonedx-gomod@v1.12.0 >"${evidence}/generator-resolution.json"
for library in jsonapi localized; do
  case "${library}" in
    jsonapi)
      module=github.com/faustbrian/go-jsonapi/v2; version=v2.0.1
      revision=6f44f8b7cd089d36bc36e67e6aec38faa5ad1aa0
      sum='h1:6kfqQ3swtoHvKcaHVqftn6YnakRxkxWFl7K7BvMKSjc='
      mod_sum='h1:MV6Dr+s3ii1yQK/FIZ9Z861Z+7ToQhiuxvZssu1nSEY='
      ;;
    localized)
      module=github.com/faustbrian/go-localized/v5; version=v5.0.2
      revision=534315549084994f3d2a2e088400ddc2bf3a6b68
      sum='h1:gTAPWcue3YTBZtHKNLAe4hwBSr8U4TecrXOFhz1nPa0='
      mod_sum='h1:tnQlj1cAbCt0tc0JpAfxkbd/G6RHVUBPGjUoxx14m2Q='
      ;;
  esac
  directory="${evidence}/${library}"
  mkdir "${directory}"
  go mod download -json "${module}@${version}" >"${directory}/module-resolution.json"
  jq -e --arg module "${module}" --arg version "${version}" --arg source "${revision}" \
    --arg sum "${sum}" --arg mod_sum "${mod_sum}" '
    .Path == $module and .Version == $version and .Sum == $sum and .GoModSum == $mod_sum and
    .Origin.Hash == $source and .Origin.Ref == ("refs/tags/" + $version) and
    .Origin.VCS == "git" and .Origin.URL == ("https://" + ($module | sub("/v[0-9]+$"; ""))) and (has("Error") | not)
  ' "${directory}/module-resolution.json" >/dev/null
  module_directory="$(jq -r .Dir "${directory}/module-resolution.json")"
  [[ "${module_directory}" == "${GOMODCACHE}/"* ]]
  (
    cd "${module_directory}"
    sha256sum go.mod go.sum >"${directory}/module-files.sha256"
    go list -mod=readonly -m -json all | jq -s . >"${directory}/dependencies.json"
    jq -e 'length > 0 and all(.[]; has("Replace") | not)' "${directory}/dependencies.json" >/dev/null
    go run "${generator}" mod -json -licenses -type library -noserial -notimestamp -output "${directory}/sbom.cdx.json" .
    jq -e '.bomFormat == "CycloneDX" and .specVersion == "1.6" and (.metadata.component | type == "object")' \
      "${directory}/sbom.cdx.json" >/dev/null
    go mod verify >"${directory}/module-verification.txt"
    sha256sum -c "${directory}/module-files.sha256"
  )
  jq -n --arg module "${module}" --arg version "${version}" --arg source "${revision}" \
    --arg source_root "${module_directory}" --arg generator "${generator}" --arg generated_at "$(date -u +%FT%TZ)" \
    --argjson bom_root_version "$(jq '.metadata.component.version // null' "${directory}/sbom.cdx.json")" \
    '{retrospective:true, original_release_pipeline:false, module:$module, release_version:$version,
      source:$source, source_root:$source_root, generator:$generator, generated_at:$generated_at,
      autodetected_bom_root_version:$bom_root_version}' >"${directory}/metadata.json"
  (cd "${directory}" && sha256sum sbom.cdx.json dependencies.json >digests.sha256)
done

git -c init.defaultBranch=main init --quiet "${task}/tools"
git -C "${task}/tools" -c credential.helper= fetch --quiet --depth=1 "https://github.com/${repository}.git" "${source_revision}"
git -C "${task}/tools" checkout --quiet --detach FETCH_HEAD
test "$(git -C "${task}/tools" rev-parse HEAD)" = "${source_revision}"
cd "${task}/tools"
sha256sum go.mod go.sum >"${task}/tools-module-files.sha256"
go list -mod=readonly -m -json all | jq -s . >"${evidence}/tools-dependencies.json"
jq -e 'length > 0 and all(.[]; has("Replace") | not)' "${evidence}/tools-dependencies.json" >/dev/null
for platform in "${platforms[@]}"; do
  output="${task}/build/${platform}"
  mkdir "${output}"
  GOOS="${platform%_*}" GOARCH="${platform#*_}" go build -trimpath -buildvcs=false \
    -ldflags='-s -w -X github.com/faustbrian/go-library-tools/v2/internal/buildinfo.Version=v2.5.1 -X github.com/faustbrian/go-library-tools/v2/internal/buildinfo.DesignLanguageSourceIdentity=v2.5.1' \
    -o "${output}/golib" ./cmd/golib
  archive="golib_2.5.1_${platform}.tar.gz"
  GOOS=linux GOARCH=amd64 go run ./cmd/golib release package --binary "${output}/golib" --license LICENSE --output "${output}/${archive}"
  for candidate in "${task}/public/${archive}" "${output}/${archive}"; do
    GOOS=linux GOARCH=amd64 go run ./cmd/golib release archive-check --file "${candidate}"
  done
  tar -xOzf "${task}/public/${archive}" golib >"${output}/published-golib"
  cmp "${output}/published-golib" "${output}/golib"
  cmp "${task}/public/${archive}" "${output}/${archive}"
  go version -m "${output}/golib" >"${evidence}/${platform}.buildinfo.txt"
  jq -n --arg platform "${platform}" --arg source "${source_revision}" \
    --arg archive_sha256 "$(sha256sum "${output}/${archive}" | cut -d' ' -f1)" \
    --arg binary_sha256 "$(sha256sum "${output}/golib" | cut -d' ' -f1)" \
    '{platform:$platform, source:$source, version:"v2.5.1", binary_and_archive_byte_equal:true,
      archive_sha256:$archive_sha256, binary_sha256:$binary_sha256}' >"${evidence}/${platform}.comparison.json"
done
go mod verify >"${evidence}/tools-module-verification.txt"
sha256sum -c "${task}/tools-module-files.sha256"
test -z "$(git status --porcelain)"
(cd "${evidence}" && find . -type f ! -name exit-status.txt -print0 | sort -z | xargs -0 sha256sum >"${task}/evidence.sha256")
mv "${task}/evidence.sha256" "${evidence}/evidence.sha256"
