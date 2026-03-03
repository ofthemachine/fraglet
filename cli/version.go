package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed build-info.json
var buildInfoRaw []byte

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
	Dirty     bool   `json:"dirty"`
}

func loadBuildInfo() BuildInfo {
	var info BuildInfo
	if err := json.Unmarshal(buildInfoRaw, &info); err != nil {
		return BuildInfo{Version: "dev", Commit: "unknown", BuildTime: "unknown", Dirty: true}
	}
	return info
}

func handleVersion() {
	info := loadBuildInfo()

	shortCommit := info.Commit
	if len(shortCommit) > 12 {
		shortCommit = shortCommit[:12]
	}

	fmt.Printf("fragletc %s\n", info.Version)
	fmt.Printf("  commit: %s\n", shortCommit)
	fmt.Printf("  built:  %s\n", info.BuildTime)
	if info.Dirty {
		fmt.Printf("  dirty:  true\n")
	}
	os.Exit(0)
}
