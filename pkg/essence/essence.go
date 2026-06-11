package essence

import (
	"context"
	"fmt"

	"github.com/ofthemachine/fraglet/pkg/runner"
	"github.com/ofthemachine/fraglet/pkg/vein"
)

// Run executes the "essence" command in the container for the given vein and optional mode,
// or uses container image directly when image is non-empty (veinName must then be empty).
// Uses the same vein loading and runner path as the CLI; respects FRAGLET_VEINS_PATH for vein lookup.
// When mode is non-empty, sets FRAGLET_MODE for the container entrypoint.
func Run(ctx context.Context, registry *vein.VeinRegistry, veinName, mode, image string) (runner.RunResult, error) {
	if image != "" && veinName != "" {
		return runner.RunResult{}, fmt.Errorf("cannot specify vein name together with --image")
	}
	if image == "" && veinName == "" {
		return runner.RunResult{}, fmt.Errorf("specify vein name or --image")
	}

	var img string
	if image != "" {
		img = image
	} else {
		if registry == nil {
			return runner.RunResult{}, fmt.Errorf("vein registry required for vein lookup")
		}
		v, ok := registry.Get(veinName)
		if !ok {
			return runner.RunResult{}, fmt.Errorf("vein not found: %s", veinName)
		}
		img = v.ContainerImage()
	}
	var envVars []string
	if mode != "" {
		envVars = append(envVars, fmt.Sprintf("FRAGLET_MODE=%s", mode))
	}
	r := runner.NewRunner(img, "")
	spec := runner.RunSpec{
		Container: img,
		Env:       envVars,
		Args:      []string{"essence"},
	}
	return r.Run(ctx, spec)
}
