package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/engine"
)

// receiptDestination is --receipt's path when given; otherwise, when
// FRAGLETC_RECEIPT_DIR is set, a file in it named by start time, script,
// and memo key (or "unkeyed" for a run that has none). Unset, no receipt
// is written -- receipts are computed on every run but saved only where
// the caller said to.
func receiptDestination(explicit, dir, scriptFile string, report engine.RunReport) string {
	if explicit != "" {
		return explicit
	}
	if dir == "" {
		return ""
	}
	key := "unkeyed"
	if k, ok := report.MemoKey(); ok {
		key = strings.TrimPrefix(k, "sha256:")[:12]
	}
	name := strings.TrimSuffix(filepath.Base(scriptFile), filepath.Ext(scriptFile))
	if name == "" || name == "." {
		name = "inline"
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%s-%s.json", report.Started.Format("20060102T150405Z"), name, key))
}
