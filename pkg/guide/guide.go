package guide

import (
	"context"
	"fmt"

	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

// Run executes the "guide" command in the container for the given vein and optional mode.
// Uses the same vein loading and runner path as the CLI; respects FRAGLET_VEINS_PATH.
// When mode is non-empty, sets both FRAGLET_MODE and FRAGLET_CONFIG for compatibility
// with existing container entrypoints that read FRAGLET_CONFIG.
func Run(ctx context.Context, registry *vein.VeinRegistry, veinName, mode string) (runner.RunResult, error) {
	v, ok := registry.Get(veinName)
	if !ok {
		return runner.RunResult{}, fmt.Errorf("vein not found: %s", veinName)
	}
	img := v.ContainerImage()
	var envVars []string
	if mode != "" {
		envVars = append(envVars, fmt.Sprintf("FRAGLET_MODE=%s", mode))
	}
	r := runner.NewRunner(img, "")
	spec := runner.RunSpec{
		Container: img,
		Env:       envVars,
		Args:      []string{"guide"},
	}
	return r.Run(ctx, spec)
}
