#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}/assets/mcp"
mkdir -p bin
go build -o bin/sd-mcp-server ./cmd/sd-mcp-server
# Platform-specific name for hub resolver
cp -f bin/sd-mcp-server "bin/sd-mcp-server-$(go env GOOS)-$(go env GOARCH)"
echo "Built assets/mcp/bin/sd-mcp-server"
