#!/usr/bin/env bash
# Match the reusable Tools source-bootstrap ownership boundary. Refuse foreign
# paths and aliases before recursive chmod/deletion; cleanup is not a general rm.
set -euo pipefail
if [[ -z "${GOLIB_SOURCE_INSTALL_ROOT:-}" ]]; then
  exit 0
fi
root="${GOLIB_SOURCE_INSTALL_ROOT}"
case "${root}" in
  "${RUNNER_TEMP}"/golib-source-install.*/*) echo 'unexpected source bootstrap workspace' >&2; exit 1 ;;
  "${RUNNER_TEMP}"/golib-source-install.*) ;;
  *) echo 'unexpected source bootstrap workspace' >&2; exit 1 ;;
esac
[[ ! -L "${root}" ]] || { echo 'symbolic source bootstrap workspace' >&2; exit 1; }
printf 'TMPDIR=%s\n' "${RUNNER_TEMP}" >>"${GITHUB_ENV}"
if [[ -e "${root}" ]]; then
  [[ -d "${root}" ]] || { echo 'non-directory source bootstrap workspace' >&2; exit 1; }
  # The pinned installer creates private directories and regular files only.
  # Reject unexpected links/devices rather than touching another owner's inode.
  unsafe="$(find "${root}" -mindepth 1 ! -type d ! -type f -print -quit)"
  linked="$(find "${root}" -type f -links +1 -print -quit)"
  [[ -z "${unsafe}" && -z "${linked}" ]] || { echo 'unsafe source bootstrap contents' >&2; exit 1; }
  chmod -R u+w "${root}"
  find "${root}" -depth -delete
fi
test ! -e "${root}"
