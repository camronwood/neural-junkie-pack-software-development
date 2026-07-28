#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ROOT}/dist"
mkdir -p "$OUT"
id="$(grep '^id:' "${ROOT}/pack.yaml" | head -1 | awk '{print $2}')"
ver="$(grep '^version:' "${ROOT}/pack.yaml" | head -1 | awk -F'"' '{print $2}')"
artifact="${OUT}/${id}-${ver}.zip"

# Sidecar binary is required for catalog installs; build if missing.
BIN="${ROOT}/assets/mcp/bin/sd-mcp-server"
if [[ ! -f "${BIN}" ]] || file "${BIN}" | grep -q 'shell script\|ASCII text'; then
  echo "Building sd-mcp-server before packaging..."
  "${ROOT}/scripts/build-mcp-sidecar.sh"
fi
if [[ ! -f "${BIN}" ]]; then
  echo "missing ${BIN} after build-mcp" >&2
  exit 1
fi

rm -f "${artifact}"
(cd "${ROOT}" && zip -r "${artifact}" pack.yaml -x '*.DS_Store')
# Include sd-mcp-server; exclude source trees and platform-suffixed copies (resolver uses the generic name).
[[ -d "${ROOT}/assets" ]] && (cd "${ROOT}" && zip -ur "${artifact}" assets -x '*.DS_Store' -x 'assets/mcp/tools/*' -x 'assets/mcp/cmd/*' -x 'assets/mcp/host/*' -x 'assets/mcp/shared/*' -x 'assets/mcp/go.*' -x 'assets/mcp/bin/*-*')
[[ -d "${ROOT}/scenarios" ]] && (cd "${ROOT}" && zip -ur "${artifact}" scenarios -x '*.DS_Store')
if ! unzip -l "${artifact}" | grep -q 'assets/mcp/bin/sd-mcp-server$'; then
  echo "pack zip missing assets/mcp/bin/sd-mcp-server" >&2
  exit 1
fi
echo "Wrote ${artifact} ($(du -h "${artifact}" | awk '{print $1}'))"
