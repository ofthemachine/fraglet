package fraglet

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"
)

// RenderTemplate renders a Go template string with provided arguments
// This is for future invocation-level templating, not used in FragletProc definition
func RenderTemplate(tmplStr string, args map[string]string) (string, error) {
	if len(args) == 0 {
		return tmplStr, nil
	}

	t, err := template.New("fraglet").Funcs(template.FuncMap{
		"arg": func(key string) string {
			return args[key]
		},
		"quote": func(s string) string {
			return strconv.Quote(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		// Add other template functions as needed
	}).Parse(tmplStr)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}
