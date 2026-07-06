package observability

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	host "github.com/camronwood/neural-junkie-pack-software-development/mcp/host"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/shared"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ObservabilityMCP struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	config     *host.MCPServerConfig
}

func NewObservabilityMCP() (*ObservabilityMCP, error) {
	config := host.GetMCPServerConfig("sre")
	mcpServer, httpServer, err := host.NewMCPServer(config)
	if err != nil {
		return nil, err
	}
	o := &ObservabilityMCP{mcpServer: mcpServer, httpServer: httpServer, config: config}
	o.registerTools()
	return o, nil
}

func (o *ObservabilityMCP) Start() error {
	return host.StartMCPServer(o.httpServer, o.config.Port)
}

func (o *ObservabilityMCP) GetMCPServer() *server.MCPServer {
	return o.mcpServer
}

func (o *ObservabilityMCP) registerTools() {
	o.mcpServer.AddTool(host.CreateTool("query_prometheus", "Run an instant PromQL query against Prometheus", host.CreateStringInputSchema("query", "PromQL expression"), nil), o.handleQueryPrometheus)
	o.mcpServer.AddTool(host.CreateTool("check_pod_logs", "Fetch recent logs for a Kubernetes pod", host.CreateMultiStringInputSchema(map[string]string{
		"pod_name":  "Pod name",
		"namespace": "Kubernetes namespace (default: default)",
	}), nil), o.handleCheckPodLogs)
	o.mcpServer.AddTool(host.CreateTool("analyze_trace", "Summarize an OpenTelemetry or Jaeger trace JSON snippet", host.CreateStringInputSchema("trace_json", "Trace JSON or span export"), nil), o.handleAnalyzeTrace)
	o.mcpServer.AddTool(host.CreateTool("check_alert_rules", "Lint Prometheus alert rule YAML for common issues", host.CreateStringInputSchema("rules_file", "Path to alert rules YAML"), nil), o.handleCheckAlertRules)
}

func (o *ObservabilityMCP) handleQueryPrometheus(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	q := req.GetString("query", "")
	if q == "" {
		return host.HandleToolError(fmt.Errorf("query required"), "query_prometheus"), nil
	}
	url := strings.TrimSpace(os.Getenv("PROMETHEUS_URL"))
	if url == "" {
		url = "http://127.0.0.1:9090"
	}
	out, err := exec.CommandContext(ctx, "curl", "-sf", url+"/api/v1/query", "--data-urlencode", "query="+q).CombinedOutput()
	return host.HandleToolSuccess(shared.FormatCommandResult("prometheus query:", string(out), err)), nil
}

func (o *ObservabilityMCP) handleCheckPodLogs(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	pod := req.GetString("pod_name", "")
	ns := req.GetString("namespace", "default")
	if pod == "" {
		return host.HandleToolError(fmt.Errorf("pod_name required"), "check_pod_logs"), nil
	}
	out, err := shared.RunCommandMaybeRemote(ctx, "", "kubectl", "logs", pod, "-n", ns, "--tail=200")
	return host.HandleToolSuccess(shared.FormatCommandResult("kubectl logs:", out, err)), nil
}

func (o *ObservabilityMCP) handleAnalyzeTrace(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	raw := req.GetString("trace_json", "")
	if raw == "" {
		return host.HandleToolError(fmt.Errorf("trace_json required"), "analyze_trace"), nil
	}
	notes := []string{"Trace payload received (" + fmt.Sprintf("%d", len(raw)) + " bytes)."}
	if strings.Contains(raw, "error") || strings.Contains(raw, "Error") {
		notes = append(notes, "Contains error markers — inspect failing spans first.")
	}
	if strings.Contains(raw, "duration") || strings.Contains(raw, "durationMs") {
		notes = append(notes, "Duration fields present — sort spans by latency.")
	}
	return host.HandleToolSuccess(strings.Join(notes, "\n")), nil
}

func (o *ObservabilityMCP) handleCheckAlertRules(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	path := req.GetString("rules_file", "")
	if path == "" {
		return host.HandleToolError(fmt.Errorf("rules_file required"), "check_alert_rules"), nil
	}
	out, err := shared.RunCommand(context.Background(), "", "promtool", "check", "rules", path)
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		return host.HandleToolSuccess(host.MissingBinaryMessage("promtool", "Install Prometheus toolkit for rule linting.")), nil
	}
	return host.HandleToolSuccess(shared.FormatCommandResult("promtool check rules:", out, err)), nil
}
