#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"
fail() { echo "verify-pack: $*" >&2; exit 1; }
[[ -f pack.yaml ]] || fail "missing pack.yaml"
id="$(grep '^id:' pack.yaml | head -1 | awk '{print $2}')"
ver="$(grep '^version:' pack.yaml | head -1 | awk -F'"' '{print $2}')"
[[ -n "${id}" ]] || fail "pack.yaml missing id"
[[ -n "${ver}" ]] || fail "pack.yaml missing version"
[[ -f assets/WORKSPACE.md ]] || fail "missing assets/WORKSPACE.md"
[[ -f assets/mcp/cmd/sd-mcp-server/main.go ]] || fail "missing assets/mcp/cmd/sd-mcp-server"
for rb in security-review migration-planning incident-handoff; do
  [[ -f "assets/runbooks/${rb}.md" ]] || fail "missing assets/runbooks/${rb}.md"
done
if [[ ! -f assets/mcp/bin/sd-mcp-server ]]; then
  "${ROOT}/scripts/build-mcp-sidecar.sh"
fi
health_port="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')"
NJ_MCP_AGENTS_JSON='["backend"]' assets/mcp/bin/sd-mcp-server --health-port="${health_port}" &
pid=$!
trap 'kill ${pid} 2>/dev/null || true' EXIT
for _ in $(seq 1 30); do
  if curl -sf "http://127.0.0.1:${health_port}/health" >/dev/null; then
    break
  fi
  sleep 0.2
done
curl -sf "http://127.0.0.1:${health_port}/health" >/dev/null || fail "mcp sidecar health check failed"
kill "${pid}" 2>/dev/null || true
trap - EXIT
"${ROOT}/scripts/build-pack-zip.sh" >/dev/null
echo "OK pack ${id} ${ver}"
