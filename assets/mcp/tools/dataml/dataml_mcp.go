package dataml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	host "github.com/camronwood/neural-junkie-pack-software-development/mcp/host"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type DataMLMCP struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	config     *host.MCPServerConfig
}

func NewDataMLMCP() (*DataMLMCP, error) {
	config := host.GetMCPServerConfig("data-ml")
	mcpServer, httpServer, err := host.NewMCPServer(config)
	if err != nil {
		return nil, err
	}
	d := &DataMLMCP{mcpServer: mcpServer, httpServer: httpServer, config: config}
	d.registerTools()
	return d, nil
}

func (d *DataMLMCP) Start() error { return host.StartMCPServer(d.httpServer, d.config.Port) }
func (d *DataMLMCP) GetMCPServer() *server.MCPServer { return d.mcpServer }

func (d *DataMLMCP) registerTools() {
	d.mcpServer.AddTool(host.CreateTool("profile_dataset", "Summarize a CSV dataset (row count, columns)", host.CreateStringInputSchema("csv_path", "Path to CSV file"), nil), d.handleProfileDataset)
	d.mcpServer.AddTool(host.CreateTool("check_notebook", "Lint a Jupyter notebook for missing cells and kernel metadata", host.CreateStringInputSchema("notebook_path", "Path to .ipynb file"), nil), d.handleCheckNotebook)
	d.mcpServer.AddTool(host.CreateTool("validate_ml_pipeline", "Check pipeline folder for train/eval artifacts", host.CreateStringInputSchema("pipeline_path", "ML pipeline directory"), nil), d.handleValidatePipeline)
}

func (d *DataMLMCP) handleProfileDataset(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	path := req.GetString("csv_path", "")
	if path == "" {
		return host.HandleToolError(fmt.Errorf("csv_path required"), "profile_dataset"), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return host.HandleToolError(err, "profile_dataset"), nil
	}
	lines := strings.Split(string(data), "\n")
	cols := 0
	if len(lines) > 0 {
		cols = len(strings.Split(lines[0], ","))
	}
	return host.HandleToolSuccess(fmt.Sprintf("Dataset %s: ~%d rows, ~%d columns (comma-separated estimate)", filepath.Base(path), len(lines), cols)), nil
}

func (d *DataMLMCP) handleCheckNotebook(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	path := req.GetString("notebook_path", "")
	if path == "" {
		return host.HandleToolError(fmt.Errorf("notebook_path required"), "check_notebook"), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return host.HandleToolError(err, "check_notebook"), nil
	}
	raw := string(data)
	notes := []string{fmt.Sprintf("Notebook size: %d bytes", len(data))}
	if strings.Contains(raw, `"cells"`) {
		notes = append(notes, "cells array present")
	}
	if strings.Contains(raw, `"kernelspec"`) {
		notes = append(notes, "kernelspec metadata present")
	} else {
		notes = append(notes, "missing kernelspec metadata")
	}
	return host.HandleToolSuccess(strings.Join(notes, "\n")), nil
}

func (d *DataMLMCP) handleValidatePipeline(_ context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	root := req.GetString("pipeline_path", "")
	if root == "" {
		return host.HandleToolError(fmt.Errorf("pipeline_path required"), "validate_ml_pipeline"), nil
	}
	found := []string{}
	for _, name := range []string{"train.py", "pipeline.py", "mlflow.yaml", "requirements.txt", "Dockerfile"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			found = append(found, name)
		}
	}
	if len(found) == 0 {
		return host.HandleToolSuccess("No standard ML pipeline files found in " + root), nil
	}
	return host.HandleToolSuccess("Pipeline artifacts: " + strings.Join(found, ", ")), nil
}
