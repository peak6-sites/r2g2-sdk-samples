package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	mcpServer := server.NewMCPServer("stock-demo", "v1.0.0")
	tool := mcp.NewTool("get_stock_price",
		mcp.WithDescription("Get the price of a stock"),
		mcp.WithString("symbol",
			mcp.Description("The stock to get the price of"),
			mcp.Required(),
		),
	)
	mcpServer.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		symbol := req.Params.Arguments["symbol"]
		results := map[string]any{
			"symbol": symbol,
			"price":  100.0,
		}
		js, _ := json.Marshal(results)
		return mcp.NewToolResultText(string(js)), nil
	})
	stdioServer := server.NewStdioServer(mcpServer)
	err := stdioServer.Listen(ctx, os.Stdin, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
