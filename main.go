package main

import (
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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

	const addr = ":8080"
	log.Printf("MCP HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, middleware.HTTPLoggingMiddleware(handler)); err != nil {
		log.Fatal(err)
	}
}
