package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ofthemachine/fraglet/mcp/tools"
)

func main() {
	tools.Server.Run(context.Background(), &mcp.StdioTransport{})
}
