package tools

import (
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var Server *mcp.Server

var (
	runSavePath   string
	runSavePathMu sync.RWMutex
)

// SetRunSavePath sets the directory path for persisting successfully run fraglets.
// When non-empty, the MCP run tool will save artifacts (content-addressed) after successful runs.
// Must be called before Server.Run if persistence is desired (e.g. from fragletc mcp --save=/path).
func SetRunSavePath(path string) {
	runSavePathMu.Lock()
	defer runSavePathMu.Unlock()
	runSavePath = path
}

func getRunSavePath() string {
	runSavePathMu.RLock()
	defer runSavePathMu.RUnlock()
	return runSavePath
}

func init() {
	Server = mcp.NewServer(
		&mcp.Implementation{
			Name:    "fraglet",
			Title:   "fraglet - Fraglet Runner",
			Version: "v0.0.1",
		}, nil)

	mcp.AddTool(Server, RunTool, Run)
	mcp.AddTool(Server, LanguageHelpTool, LanguageHelp)
}
