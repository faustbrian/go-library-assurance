#!/usr/bin/env bash
# Quality owns classification. A false classification expects a skipped scenario
# job, not a manufactured success; selected, missing or cancelled work stays fatal.
set -euo pipefail
test "${QUALITY_RESULT}" = success
test "${SECRETS_RESULT}" = success
test "${SOURCE_DOCS_RESULT}" = success
case "${SCENARIOS_REQUIRED}" in
  true) test "${SCENARIOS_RESULT}" = success ;;
  false) test "${SCENARIOS_RESULT}" = skipped ;;
  *) exit 1 ;;
esac
case "${MINIMUM_COMPILER_REQUIRED}" in
  true)
    test "${EVENT_NAME}" = workflow_dispatch
    test "${MINIMUM_COMPILER_RESULT}" = success
    ;;
  false) test "${MINIMUM_COMPILER_RESULT}" = skipped ;;
  *) exit 1 ;;
esac
case "${RELEASE_INTEGRITY_REQUIRED}" in
  true)
    test "${EVENT_NAME}" = workflow_dispatch
    test "${RELEASE_INTEGRITY_RESULT}" = success
    ;;
  false) test "${RELEASE_INTEGRITY_RESULT}" = skipped ;;
  *) exit 1 ;;
esac
case "${EVENT_NAME}" in
  pull_request) test "${DEPENDENCY_REVIEW_RESULT}" = success ;;
  push|workflow_dispatch) test "${DEPENDENCY_REVIEW_RESULT}" = skipped ;;
  *) exit 1 ;;
esac
