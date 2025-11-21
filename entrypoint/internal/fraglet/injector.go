package fraglet

import (
	"os"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

// Injector performs fraglet injection into a single target file.
type Injector struct{}

func NewInjector() *Injector {
	return &Injector{}
}

// Inject performs fraglet injection using the file injector.
// The injection config already contains the CodePath, so we just pass it through.
func (i *Injector) Inject(fragletPath string, injection fraglet.InjectionConfig) error {
	if err := fraglet.InjectFile(fragletPath, &injection); err != nil {
		return err
	}
	// Remove temp fraglet file after injection
	_ = os.Remove(fragletPath)
	return nil
}
