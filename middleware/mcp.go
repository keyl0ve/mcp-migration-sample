package middleware

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func MCPLoggingMiddleware() func(next mcp.MethodHandler) mcp.MethodHandler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return func(next mcp.MethodHandler) mcp.MethodHandler {
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
	}
}
