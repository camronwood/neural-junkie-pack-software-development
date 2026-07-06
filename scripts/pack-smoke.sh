#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NJ_ROOT="${NEURAL_JUNKIE_ROOT:-$(cd "${ROOT}/../neural-junkie" 2>/dev/null && pwd || true)}"
if [[ -z "${NJ_ROOT}" || ! -d "${NJ_ROOT}/scripts" ]]; then
  echo "pack-smoke: set NEURAL_JUNKIE_ROOT to neural-junkie checkout" >&2
  exit 1
fi
export NEURAL_JUNKIE_SCENARIO_REPO="${NJ_ROOT}"
echo "pack-smoke: implement go-build-error-fix"
python3 "${NJ_ROOT}/scripts/implement-scenarios.py" --pack-dir "${ROOT}" --scenario go-build-error-fix || exit 1
echo "pack-smoke: collab sre-alert-triage"
python3 "${NJ_ROOT}/scripts/collab-scenarios.py" --pack-dir "${ROOT}" --scenario sre-alert-triage || exit 1
echo "pack-smoke OK"
