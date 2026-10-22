package fraglet

import (
	"fmt"
	"strings"
)

// shebangValueFlags are fragletc's flags a shebang may carry, each taking a
// value (as -flag value or -flag=value). The kernel appends the script's
// path after them, so nothing else may follow.
var shebangValueFlags = map[string]bool{
	"vein": true, "v": true, "image": true, "i": true, "mode": true, "m": true,
	"network": true, "fraglet-path": true, "output-dir": true, "e": true,
	"param": true, "p": true, "output": true, "receipt": true, "stdin": true,
	"retry": true, "timeout": true,
}

// lintShebang checks a fragletc shebang as fragletc will read it: every
// flag one it has, each value flag given a value (not the next flag), no
// word left over that fragletc would take for the script, arguments passed
// through env -S (Linux hands a shebang's arguments over as one word), and
// an image pinned by digest. A first line that is not a fragletc shebang
// is not this lint's business.
func lintShebang(code string) []Finding {
	first, _, _ := strings.Cut(code, "\n")
	if !strings.HasPrefix(first, "#!") || !strings.Contains(first, "fragletc") {
		return nil
	}
	fields := strings.Fields(strings.TrimPrefix(first, "#!"))
	i := 0
	for i < len(fields) && !strings.HasSuffix(fields[i], "fragletc") {
		i++
	}
	if i == len(fields) {
		return nil
	}
	args := fields[i+1:]
	viaEnvS := i >= 2 && strings.HasSuffix(fields[0], "env") && fields[1] == "-S"
	if len(args) > 0 && !viaEnvS {
		return []Finding{{1, "shebang-env-split", Error,
			"a shebang's arguments reach fragletc as one word on Linux: start it #!/usr/bin/env -S fragletc …"}}
	}
	var out []Finding
	for j := 0; j < len(args); j++ {
		arg := args[j]
		if !strings.HasPrefix(arg, "-") {
			out = append(out, Finding{1, "shebang-positional", Error,
				fmt.Sprintf("%q after fragletc's flags would be read as the script (the script's own path comes last, from the kernel); drop it, or give it as a flag's value", arg)})
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if !shebangValueFlags[name] {
			out = append(out, Finding{1, "shebang-flag", Error,
				fmt.Sprintf("fragletc has no flag %s a shebang can carry (fragletc --help lists them)", arg)})
			if !hasValue && j+1 < len(args) && !strings.HasPrefix(args[j+1], "-") {
				j++ // its value, said once
			}
			continue
		}
		if !hasValue {
			if j+1 == len(args) || strings.HasPrefix(args[j+1], "-") {
				next := "nothing"
				if j+1 < len(args) {
					next = args[j+1]
				}
				out = append(out, Finding{1, "shebang-flag-value", Error,
					fmt.Sprintf("%s takes a value, but is followed by %s (e.g. -v python, or --image name@sha256:…)", arg, next)})
				continue
			}
			j++
			value = args[j]
		}
		if (name == "image" || name == "i") && !strings.Contains(value, "@sha256:") {
			out = append(out, Finding{1, "shebang-image-unpinned", Warning,
				fmt.Sprintf("image %s is not pinned: a tag can move, so the same script may run differently later; pin it as name@sha256:…", value)})
		}
	}
	return out
}
