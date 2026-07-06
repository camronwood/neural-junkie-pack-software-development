package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/architecture"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/backend"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/codereview"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/database"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/dataml"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/devops"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/frontend"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/mobile"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/observability"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/rust"
	"github.com/camronwood/neural-junkie-pack-software-development/mcp/tools/security"
)

type mcpStarter interface {
	Start() error
}

func main() {
	healthPort := flag.Int("health-port", 8765, "HTTP port for /health")
	flag.Parse()

	agents := defaultAgents()
	if raw := strings.TrimSpace(os.Getenv("NJ_MCP_AGENTS_JSON")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &agents); err != nil {
			log.Fatalf("NJ_MCP_AGENTS_JSON: %v", err)
		}
	}

	enabled := make(map[string]struct{}, len(agents))
	for _, a := range agents {
		enabled[strings.ToLower(strings.TrimSpace(a))] = struct{}{}
	}

	start := func(agent string, fn func() (mcpStarter, error)) {
		if _, ok := enabled[agent]; !ok {
			return
		}
		srv, err := fn()
		if err != nil {
			log.Fatalf("start %s MCP: %v", agent, err)
		}
		if err := srv.Start(); err != nil {
			log.Fatalf("start %s MCP server: %v", agent, err)
		}
		log.Printf("Started MCP for agent type %q", agent)
	}

	start("backend", func() (mcpStarter, error) { return backend.NewBackendMCP() })
	start("frontend", func() (mcpStarter, error) { return frontend.NewFrontendMCP() })
	start("devops", func() (mcpStarter, error) { return devops.NewDevOpsMCP() })
	start("database", func() (mcpStarter, error) { return database.NewDatabaseMCP() })
	start("security", func() (mcpStarter, error) { return security.NewSecurityMCP() })
	start("code-review", func() (mcpStarter, error) { return codereview.NewCodeReviewMCP() })
	start("architecture", func() (mcpStarter, error) { return architecture.NewArchitectureMCP() })
	start("rust", func() (mcpStarter, error) { return rust.NewRustMCP() })
	start("sre", func() (mcpStarter, error) { return observability.NewObservabilityMCP() })
	start("mobile", func() (mcpStarter, error) { return mobile.NewMobileMCP() })
	start("data-ml", func() (mcpStarter, error) { return dataml.NewDataMLMCP() })

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"pack_id":"software-development"}`))
	})
	addr := fmt.Sprintf("127.0.0.1:%d", *healthPort)
	log.Printf("sd-mcp-server health listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func defaultAgents() []string {
	return []string{
		"backend", "frontend", "devops", "database", "security", "code-review", "architecture",
		"rust", "sre", "mobile", "data-ml",
	}
}
