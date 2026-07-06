.PHONY: verify pack-zip pack-smoke build-mcp clean

verify:
	./scripts/verify-pack.sh

pack-zip:
	./scripts/build-pack-zip.sh

build-mcp:
	./scripts/build-mcp-sidecar.sh

pack-smoke: verify
	./scripts/pack-smoke.sh

clean:
	rm -rf dist assets/mcp/bin
