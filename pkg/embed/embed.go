package embed

import (
	"embed"
	"gopkg.in/yaml.v3"

	"github.com/ofthemachine/fraglet/pkg/vein"
)

//go:embed veins.yml
var veinsFS embed.FS

// LoadEmbeddedVeins loads veins from the embedded veins.yml file
func LoadEmbeddedVeins() (*vein.VeinRegistry, error) {
	data, err := veinsFS.ReadFile("veins.yml")
	if err != nil {
		return nil, err
	}

	// Parse the YAML
	var config struct {
		Veins []*vein.Vein `yaml:"veins"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	registry := vein.NewVeinRegistry()
	for _, v := range config.Veins {
		if err := registry.Add(v); err != nil {
			return nil, err
		}
	}

	return registry, nil
}
