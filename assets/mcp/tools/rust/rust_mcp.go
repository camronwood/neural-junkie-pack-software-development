package rust

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	host "github.com/camronwood/neural-junkie-pack-software-development/mcp/host"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/shared"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RustMCP provides MCP tools for Rust development.
type RustMCP struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	config     *host.MCPServerConfig
}

// NewRustMCP creates a new Rust MCP server.
func NewRustMCP() (*RustMCP, error) {
	config := host.GetMCPServerConfig("rust")
	mcpServer, httpServer, err := host.NewMCPServer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}
	r := &RustMCP{mcpServer: mcpServer, httpServer: httpServer, config: config}
	r.registerTools()
	return r, nil
}

func (r *RustMCP) Start() error {
	return host.StartMCPServer(r.httpServer, r.config.Port)
}

func (r *RustMCP) GetMCPServer() *server.MCPServer {
	return r.mcpServer
}

func (r *RustMCP) registerTools() {
	r.mcpServer.AddTool(host.CreateTool(
		"cargo_clippy",
		"Run cargo clippy lints on a Rust crate",
		host.CreateStringInputSchema("crate_path", "Path to Cargo.toml directory"),
		nil,
	), r.handleCargoClippy)

	r.mcpServer.AddTool(host.CreateTool(
		"cargo_test",
		"Run cargo test in a Rust crate",
		host.CreateStringInputSchema("crate_path", "Path to Cargo.toml directory"),
		nil,
	), r.handleCargoTest)

	r.mcpServer.AddTool(host.CreateTool(
		"cargo_audit",
		"Run cargo audit for crate vulnerability advisories",
		host.CreateStringInputSchema("crate_path", "Path to Cargo.toml directory"),
		nil,
	), r.handleCargoAudit)

	r.mcpServer.AddTool(host.CreateTool(
		"check_cargo_toml",
		"Parse Cargo.toml and summarize package metadata and dependencies",
		host.CreateStringInputSchema("crate_path", "Path to Cargo.toml directory"),
		nil,
	), r.handleCheckCargoToml)

	log.Printf("Registered %d Rust MCP tools", len(r.mcpServer.ListTools()))
}

func (r *RustMCP) crateRoot(cratePath string) string {
	return shared.FindProjectRoot(cratePath, "Cargo.toml")
}

func (r *RustMCP) handleCargoClippy(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := r.crateRoot(request.GetString("crate_path", "."))
	out, err := shared.RunCommand(ctx, root, "cargo", "clippy", "--message-format=short")
	return host.HandleToolSuccess(shared.FormatCommandResult("cargo clippy:", out, err)), nil
}

func (r *RustMCP) handleCargoTest(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := r.crateRoot(request.GetString("crate_path", "."))
	out, err := shared.RunCommand(ctx, root, "cargo", "test")
	return host.HandleToolSuccess(shared.FormatCommandResult("cargo test:", out, err)), nil
}

func (r *RustMCP) handleCargoAudit(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := r.crateRoot(request.GetString("crate_path", "."))
	out, err := shared.RunCommand(ctx, root, "cargo", "audit")
	if err != nil && strings.Contains(out, "no such subcommand") {
		return host.HandleToolSuccess(host.MissingBinaryMessage("cargo-audit", "Install: cargo install cargo-audit")), nil
	}
	return host.HandleToolSuccess(shared.FormatCommandResult("cargo audit:", out, err)), nil
}

func (r *RustMCP) handleCheckCargoToml(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := r.crateRoot(request.GetString("crate_path", "."))
	path := filepath.Join(root, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return host.HandleToolError(fmt.Errorf("Cargo.toml not found: %w", err), "check_cargo_toml"), nil
	}
	var manifest struct {
		Package struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Edition string `json:"edition"`
		} `json:"package"`
		Dependencies map[string]any `json:"dependencies"`
		Features     map[string]any `json:"features"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return host.HandleToolSuccess(fmt.Sprintf("Cargo.toml (%d bytes):\n%s", len(data), string(data))), nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Crate: %s v%s (edition %s)\n", manifest.Package.Name, manifest.Package.Version, manifest.Package.Edition)
	fmt.Fprintf(&b, "Dependencies: %d\nFeatures: %d\n", len(manifest.Dependencies), len(manifest.Features))
	return host.HandleToolSuccess(b.String()), nil
}
