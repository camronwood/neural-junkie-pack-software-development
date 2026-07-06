package codereview

import (
	"context"
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

// CodeReviewMCP provides read-only diagnostic MCP tools for code review.
type CodeReviewMCP struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	config     *host.MCPServerConfig
}

// NewCodeReviewMCP creates a new Code Review MCP server.
func NewCodeReviewMCP() (*CodeReviewMCP, error) {
	config := host.GetMCPServerConfig("code-review")
	mcpServer, httpServer, err := host.NewMCPServer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}
	c := &CodeReviewMCP{mcpServer: mcpServer, httpServer: httpServer, config: config}
	c.registerTools()
	return c, nil
}

func (c *CodeReviewMCP) Start() error {
	return host.StartMCPServer(c.httpServer, c.config.Port)
}

func (c *CodeReviewMCP) GetMCPServer() *server.MCPServer {
	return c.mcpServer
}

func (c *CodeReviewMCP) registerTools() {
	c.mcpServer.AddTool(host.CreateTool(
		"analyze_go_code",
		"Analyze Go code using go vet (read-only review)",
		host.CreateStringInputSchema("file_path", "Path to Go file or directory"),
		nil,
	), c.handleAnalyzeGoCode)

	c.mcpServer.AddTool(host.CreateTool(
		"run_go_tests",
		"Run Go tests for review (read-only)",
		host.CreateStringInputSchema("package_path", "Go package path to test"),
		nil,
	), c.handleRunGoTests)

	c.mcpServer.AddTool(host.CreateTool(
		"run_eslint",
		"Run ESLint for frontend code review",
		host.CreateStringInputSchema("target_path", "Path to lint"),
		nil,
	), c.handleRunESLint)

	c.mcpServer.AddTool(host.CreateTool(
		"run_typescript_check",
		"Run TypeScript check for review",
		host.CreateStringInputSchema("project_path", "Project directory with tsconfig.json"),
		nil,
	), c.handleRunTypescriptCheck)

	log.Printf("Registered %d Code Review MCP tools", len(c.mcpServer.ListTools()))
}

func (c *CodeReviewMCP) workingDir(path string) string {
	if filepath.IsAbs(path) {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return filepath.Dir(path)
		}
		return path
	}
	return shared.FindProjectRoot(path, "go.mod", "package.json")
}

func (c *CodeReviewMCP) handleAnalyzeGoCode(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" || !shared.PathExists(filePath) {
		return host.HandleToolError(fmt.Errorf("file_path not found"), "analyze_go_code"), nil
	}
	out, err := shared.RunCommand(ctx, c.workingDir(filePath), "go", "vet", filePath)
	return host.HandleToolSuccess(shared.FormatCommandResult("go vet:", out, err)), nil
}

func (c *CodeReviewMCP) handleRunGoTests(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	pkg := request.GetString("package_path", ".")
	if !c.isValidGoPath(pkg) {
		return host.HandleToolError(fmt.Errorf("package_path not found: %s", pkg), "run_go_tests"), nil
	}
	out, err := shared.RunCommand(ctx, c.workingDir(pkg), "go", "test", "-timeout", "30s", pkg)
	return host.HandleToolSuccess(shared.FormatCommandResult("go test:", out, err)), nil
}

func (c *CodeReviewMCP) isValidGoPath(path string) bool {
	if path == "" {
		return false
	}
	if path == "." {
		return true
	}
	if strings.Contains(path, "..") {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func (c *CodeReviewMCP) handleRunESLint(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	target := request.GetString("target_path", "")
	root := shared.FindProjectRoot(target, "package.json")
	if !shared.ProjectHasESLint(root) {
		return host.HandleToolSuccess(shared.ESLintNotConfiguredMessage(root)), nil
	}
	out, err := shared.RunCommand(ctx, root, "npx", "--yes", "eslint", target)
	return host.HandleToolSuccess(shared.FormatCommandResult("eslint:", out, err)), nil
}

func (c *CodeReviewMCP) handleRunTypescriptCheck(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := shared.FindProjectRoot(request.GetString("project_path", "."), "tsconfig.json", "package.json")
	if !shared.PathExists(filepath.Join(root, "tsconfig.json")) {
		return host.HandleToolError(fmt.Errorf("tsconfig.json not found in %s", root), "run_typescript_check"), nil
	}
	if !shared.ProjectHasTypeScript(root) {
		return host.HandleToolSuccess(shared.TypeScriptNotConfiguredMessage(root)), nil
	}
	out, err := shared.RunTypeScriptCheck(ctx, root)
	return host.HandleToolSuccess(shared.FormatCommandResult("TypeScript check:", out, err)), nil
}
