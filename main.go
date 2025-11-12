package main

import (
	"log"
	"net/http"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Keyl0ve/mcp-migration-sample/authz"
	"github.com/Keyl0ve/mcp-migration-sample/middleware"
	"github.com/Keyl0ve/mcp-migration-sample/tool"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, tool.GreetTool(), tool.SayHi)
	server.AddReceivingMiddleware(middleware.MCPLoggingMiddleware())

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	jwtAuth := mcpauth.RequireBearerToken(authz.VerifyJWT, &mcpauth.RequireBearerTokenOptions{
		Scopes: []string{"read"},
	})
	apiKeyAuth := mcpauth.RequireBearerToken(authz.VerifyAPIKey, &mcpauth.RequireBearerTokenOptions{
		Scopes: []string{"read"},
	})

	mux := http.NewServeMux()
	mux.Handle("/mcp/jwt", jwtAuth(handler))
	mux.Handle("/mcp/apikey", apiKeyAuth(handler))
	mux.HandleFunc("/generate-token", authz.GenerateTokenHandler)
	mux.HandleFunc("/generate-api-key", authz.GenerateAPIKeyHandler)
	mux.HandleFunc("/health", authz.HealthHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	const addr = ":8080"
	log.Printf("MCP HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, middleware.HTTPLoggingMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
