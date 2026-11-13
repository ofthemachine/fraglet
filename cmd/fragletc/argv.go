package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

const fragletHelpArg = "--fraglet-help"

// argParser implements a simple stateful parser for argument preprocessing
type argParser struct {
	args        []string
	pos         int
	passthrough bool
	preprocessed
}

// preprocessed is what preprocessFragletArgv lifts out of argv before flag
// parsing: everything that must work in any position, because a shebang
// invocation puts the script name first and the caller's arguments after.
type preprocessed struct {
	filtered []string // argv with the lifted arguments removed
	wantHelp bool
	params   []string
	// paramsFile is --params: a JSON file of name to value, read into
	// params after the loop (a value on the command line wins).
	paramsFile string
	outputs    []string
	receipt    string
	stdin      string // --stdin=<mode>
}

func (p *argParser) peek() (string, bool) {
	if p.pos >= len(p.args) {
		return "", false
	}
	return p.args[p.pos], true
}

func (p *argParser) consume() (string, bool) {
	arg, ok := p.peek()
	if ok {
		p.pos++
	}
	return arg, ok
}

func (p *argParser) setStdin(mode string) error {
	switch mode {
	case fraglet.StdinNone, fraglet.StdinBuffer, fraglet.StdinStream:
		p.stdin = mode
		return nil
	}
	return fmt.Errorf("--stdin=%q is not none, buffer, or stream", mode)
}

func preprocessFragletArgv(args []string) (preprocessed, error) {
	p := &argParser{args: args}

	for {
		arg, ok := p.consume()
		if !ok {
			break
		}

		if p.passthrough {
			p.filtered = append(p.filtered, arg)
			continue
		}

		switch {
		case arg == "--":
			p.passthrough = true
			p.filtered = append(p.filtered, arg)

		case arg == fragletHelpArg:
			p.wantHelp = true

		case strings.HasPrefix(arg, "--param="):
			p.params = append(p.params, arg[len("--param="):])

		case arg == "--param":
			val, ok := p.consume()
			if !ok {
				return preprocessed{}, errors.New("--param requires a value")
			}
			p.params = append(p.params, val)

		case strings.HasPrefix(arg, "-p="):
			p.params = append(p.params, arg[len("-p="):])

		case arg == "-p":
			val, ok := p.consume()
			if !ok {
				return preprocessed{}, errors.New("-p requires a value")
			}
			p.params = append(p.params, val)

		case strings.HasPrefix(arg, "-p") && len(arg) > 2 && arg[2] != '=' && strings.Contains(arg[2:], "="):
			// -pKEY=value
			p.params = append(p.params, arg[2:])

		case strings.HasPrefix(arg, "--params="):
			p.paramsFile = arg[len("--params="):]

		case arg == "--params":
			val, ok := p.consume()
			if !ok {
				return preprocessed{}, errors.New("--params requires a file")
			}
			p.paramsFile = val

		case strings.HasPrefix(arg, "--output="):
			p.outputs = append(p.outputs, arg[len("--output="):])

		case arg == "--output":
			val, ok := p.consume()
			if !ok {
				return preprocessed{}, errors.New("--output requires a value")
			}
			p.outputs = append(p.outputs, val)

		case strings.HasPrefix(arg, "--receipt="):
			p.receipt = arg[len("--receipt="):]

		case arg == "--receipt":
			val, ok := p.consume()
			if !ok {
				return preprocessed{}, errors.New("--receipt requires a value")
			}
			p.receipt = val

		case strings.HasPrefix(arg, "--stdin="):
			if err := p.setStdin(arg[len("--stdin="):]); err != nil {
				return preprocessed{}, err
			}

		case arg == "--stdin":
			val, ok := p.consume()
			if !ok {
				return preprocessed{}, errors.New("--stdin requires a value (none, buffer, or stream)")
			}
			if err := p.setStdin(val); err != nil {
				return preprocessed{}, err
			}

		default:
			p.filtered = append(p.filtered, arg)
		}
	}

	if p.paramsFile != "" {
		params, err := paramsFromFile(p.paramsFile, p.params)
		if err != nil {
			return preprocessed{}, err
		}
		p.params = params
	}
	return p.preprocessed, nil
}

// paramsFromFile is given with the params of path, a JSON object of name to
// value (a string, number or boolean; a file param's value is a path), put
// before them: a name given on the command line is taken from there, so a
// run's recorded params rerun as they were, and one can still be changed.
func paramsFromFile(path string, given []string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("--params: %w", err)
	}
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("--params %s: a JSON object of name to value: %w", path, err)
	}
	set := map[string]bool{}
	for _, p := range given {
		name, _, _ := strings.Cut(p, "=")
		set[name] = true
	}
	var out []string
	for _, name := range slices.Sorted(maps.Keys(values)) {
		if set[name] {
			continue
		}
		var v string
		switch x := values[name].(type) {
		case string:
			v = x
		case float64, bool:
			v = fmt.Sprint(x)
		default:
			return nil, fmt.Errorf("--params %s: %s is not a string, number or boolean", path, name)
		}
		out = append(out, name+"="+v)
	}
	return append(out, given...), nil
}
