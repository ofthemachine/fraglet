package language

// LanguageConfig defines the configuration for a language container
type LanguageConfig struct {
	Name         string // Language identifier (e.g., "python")
	Container    string // Docker image name (e.g., "100hellos/python:local")
	FragmentPath string // Path where code fragments should be mounted (e.g., "/code-fragments/MAIN")
	Entrypoint   string // Optional entrypoint override (empty means use container default)
}

// Languages is the registry of supported languages
// To help the agent use the language, the container should respond to "guide"
var Languages = map[string]LanguageConfig{
	"python": {
		Name:         "python",
		Container:    "100hellos/python:local",
		FragmentPath: "/code-fragments/MAIN",
	},
	"prolog": {
		Name:         "prolog",
		Container:    "100hellos/prolog:local",
		FragmentPath: "/code-fragments/MAIN",
	},
	"lisp": {
		Name:         "lisp",
		Container:    "100hellos/lisp:local",
		FragmentPath: "/code-fragments/MAIN",
	},
	"lua": {
		Name:         "lua",
		Container:    "100hellos/lua:local",
		FragmentPath: "/code-fragments/MAIN",
	},
	"ruby": {
		Name:         "ruby",
		Container:    "100hellos/ruby:local",
		FragmentPath: "/code-fragments/MAIN",
	},
	"r": {
		Name:         "r",
		Container:    "100hellos/r-project:local",
		FragmentPath: "/code-fragments/MAIN",
	},
}

// GetLanguage retrieves a language configuration by name
func GetLanguage(name string) (LanguageConfig, bool) {
	config, ok := Languages[name]
	return config, ok
}

// ListLanguages returns a slice of all available language names
func ListLanguages() []string {
	names := make([]string, 0, len(Languages))
	for name := range Languages {
		names = append(names, name)
	}
	return names
}
