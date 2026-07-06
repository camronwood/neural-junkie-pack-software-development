package mobile

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	host "github.com/camronwood/neural-junkie-pack-software-development/mcp/host"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/shared"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MobileMCP struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	config     *host.MCPServerConfig
}

func NewMobileMCP() (*MobileMCP, error) {
	config := host.GetMCPServerConfig("mobile")
	mcpServer, httpServer, err := host.NewMCPServer(config)
	if err != nil {
		return nil, err
	}
	m := &MobileMCP{mcpServer: mcpServer, httpServer: httpServer, config: config}
	m.registerTools()
	return m, nil
}

func (m *MobileMCP) Start() error  { return host.StartMCPServer(m.httpServer, m.config.Port) }
func (m *MobileMCP) GetMCPServer() *server.MCPServer { return m.mcpServer }

func (m *MobileMCP) registerTools() {
	m.mcpServer.AddTool(host.CreateTool("run_react_native_doctor", "Run npx react-native doctor in project root", host.CreateStringInputSchema("project_path", "React Native project directory"), nil), m.handleRNDoctor)
	m.mcpServer.AddTool(host.CreateTool("check_ios_build", "Validate iOS workspace / Podfile presence", host.CreateStringInputSchema("project_path", "iOS or RN project directory"), nil), m.handleIOSBuild)
	m.mcpServer.AddTool(host.CreateTool("analyze_android_gradle", "Run ./gradlew tasks --all and summarize Android module layout", host.CreateStringInputSchema("android_path", "Path to android/ directory"), nil), m.handleAndroidGradle)
}

func (m *MobileMCP) handleRNDoctor(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := req.GetString("project_path", ".")
	out, err := shared.RunCommandMaybeRemote(ctx, root, "npx", "react-native", "doctor")
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		return host.HandleToolSuccess(host.MissingBinaryMessage("npx", "Install Node.js to run React Native doctor.")), nil
	}
	return host.HandleToolSuccess(shared.FormatCommandResult("react-native doctor:", out, err)), nil
}

func (m *MobileMCP) handleIOSBuild(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := req.GetString("project_path", ".")
	notes := []string{}
	for _, name := range []string{"ios/Podfile", "Podfile", "ios/*.xcworkspace"} {
		matches, _ := filepath.Glob(filepath.Join(root, name))
		if len(matches) > 0 {
			notes = append(notes, "Found: "+matches[0])
		}
	}
	if _, err := os.Stat(filepath.Join(root, "ios")); err == nil {
		notes = append(notes, "ios/ directory present")
	}
	if len(notes) == 0 {
		return host.HandleToolSuccess("No iOS project artifacts found under " + root), nil
	}
	return host.HandleToolSuccess(strings.Join(notes, "\n")), nil
}

func (m *MobileMCP) handleAndroidGradle(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	android := req.GetString("android_path", "android")
	out, err := shared.RunCommandMaybeRemote(ctx, android, "./gradlew", "tasks", "--all")
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		return host.HandleToolSuccess(host.MissingBinaryMessage("gradlew", "Open android/ folder with Gradle wrapper.")), nil
	}
	return host.HandleToolSuccess(shared.FormatCommandResult("gradle tasks:", out, err)), nil
}
