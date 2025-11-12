package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type Output struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (
	*mcp.CallToolResult,
	Output,
	error,
) {
	return nil, Output{Greeting: "Hi " + input.Name}, nil
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s from %s -> %d (%s)", r.Method, r.URL.Path, r.RemoteAddr, lrw.status, duration)
	})
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// MCPサーバーを作成
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)

	// MCPミドルウェアでリクエスト／レスポンスをログ出力
	server.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			sessionID := ""
			if session := req.GetSession(); session != nil {
				sessionID = session.ID()
			}

			logger.Info("MCP method started",
				"method", method,
				"session_id", sessionID,
				"has_params", req.GetParams() != nil,
			)

			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				logger.Info("Calling tool",
					"name", ctr.Params.Name,
					"args", ctr.Params.Arguments,
				)
			}

			start := time.Now()
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			if err != nil {
				logger.Error("MCP method failed",
					"method", method,
					"session_id", sessionID,
					"duration_ms", duration.Milliseconds(),
					"err", err,
				)
			} else {
				logger.Info("MCP method completed",
					"method", method,
					"session_id", sessionID,
					"duration_ms", duration.Milliseconds(),
					"has_result", result != nil,
				)
				if ctr, ok := result.(*mcp.CallToolResult); ok {
					logger.Info("tool result",
						"isError", ctr.IsError,
						"structuredContent", ctr.StructuredContent,
					)
				}
			}

			return result, err
		}
	})

	// HTTPハンドラーを介してMCPリクエストを処理
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	loggedHandler := loggingMiddleware(handler)

	const addr = ":8080"
	log.Printf("MCP HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, loggedHandler); err != nil {
		log.Fatal(err)
	}
}
