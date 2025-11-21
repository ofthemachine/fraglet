package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var Server *mcp.Server

func init() {
	Server = mcp.NewServer(
		&mcp.Implementation{
			Name:    "fraglet",
			Title:   "fraglet - Code Fragment Runner",
			Version: "v0.0.1",
		}, nil)

	mcp.AddTool(Server, RunTool, Run)
	mcp.AddTool(Server, LanguageHelpTool, LanguageHelp)
}
