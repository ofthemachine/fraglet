package fraglet

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Kind is what a parameter's value is, declared in the header with the kind=
// modifier (#: param=hex:kind=color), so a caller can offer the right input
// (a color picker, a slider, a choice) and check a value before running. A
// value is always a string on the wire; the kind says which strings are
// allowed. Undeclared, a parameter is a string (a file with :file).
type Kind struct {
	Name    string   // one of KindNames
	Options []string // enum: the allowed values, in declared order
	Min     float64  // range
	Max     float64  // range
	Step    float64  // range; 0 means any
}

// KindNames is the closed vocabulary of kinds.
var KindNames = []string{"string", "text", "number", "integer", "range", "enum", "bool", "color", "date", "datetime", "file", "secret"}

// ParseKind parses a kind= value: a name, enum(a|b|c), or range(min..max[,step]).
func ParseKind(s string) (Kind, error) {
	s = strings.TrimSpace(s)
	name, args, hasArgs := strings.Cut(s, "(")
	if hasArgs {
		if !strings.HasSuffix(args, ")") {
			return Kind{}, fmt.Errorf("kind %q: missing )", s)
		}
		args = strings.TrimSuffix(args, ")")
	}
	switch name {
	case "enum":
		var opts []string
		for _, o := range strings.Split(args, "|") {
			if o = strings.TrimSpace(o); o != "" {
				opts = append(opts, o)
			}
		}
		if len(opts) < 2 {
			return Kind{}, fmt.Errorf("kind %q: an enum lists at least two options, enum(a|b)", s)
		}
		return Kind{Name: name, Options: opts}, nil
	case "range":
		span, step, _ := strings.Cut(args, ",")
		lo, hi, ok := strings.Cut(span, "..")
		k := Kind{Name: name}
		var err1, err2, err3 error
		k.Min, err1 = strconv.ParseFloat(strings.TrimSpace(lo), 64)
		k.Max, err2 = strconv.ParseFloat(strings.TrimSpace(hi), 64)
		if step != "" {
			k.Step, err3 = strconv.ParseFloat(strings.TrimSpace(step), 64)
		}
		if !ok || err1 != nil || err2 != nil || err3 != nil || k.Min >= k.Max || k.Step < 0 {
			return Kind{}, fmt.Errorf("kind %q: want range(min..max) or range(min..max,step), min < max", s)
		}
		return k, nil
	}
	if hasArgs {
		return Kind{}, fmt.Errorf("kind %q: only enum and range take arguments", s)
	}
	for _, n := range KindNames {
		if n == name && n != "enum" && n != "range" {
			return Kind{Name: name}, nil
		}
	}
	return Kind{}, fmt.Errorf("kind %q: want one of %s", s, strings.Join(KindNames, ", "))
}

// String renders k as it is written in a header.
func (k Kind) String() string {
	switch k.Name {
	case "enum":
		return "enum(" + strings.Join(k.Options, "|") + ")"
	case "range":
		s := "range(" + fmtNum(k.Min) + ".." + fmtNum(k.Max)
		if k.Step > 0 {
			s += "," + fmtNum(k.Step)
		}
		return s + ")"
	}
	return k.Name
}

func fmtNum(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

// Kind returns the parameter's declared kind: kind=, else file for :file,
// else string. A kind= that does not parse reads as string (lint reports it).
func (d ParamDecl) Kind() Kind {
	if s, ok := d.Modifiers["kind"]; ok {
		if k, err := ParseKind(s); err == nil {
			return k
		}
	}
	if d.Shape == "file" {
		return Kind{Name: "file"}
	}
	return Kind{Name: "string"}
}

// Normalize checks v against k and returns it in canonical form (a color as
// #rrggbb, a bool as true/false, an enum option as declared). An empty value
// is the caller's business (required or not), so it is returned as is.
func (k Kind) Normalize(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return v, nil
	}
	switch k.Name {
	case "number":
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return "", fmt.Errorf("%q is not a number", v)
		}
	case "integer":
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			return "", fmt.Errorf("%q is not a whole number", v)
		}
	case "range":
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < k.Min || f > k.Max {
			return "", fmt.Errorf("%q is not a number from %s to %s", v, fmtNum(k.Min), fmtNum(k.Max))
		}
	case "enum":
		for _, o := range k.Options {
			if strings.EqualFold(o, v) {
				return o, nil
			}
		}
		return "", fmt.Errorf("%q is not one of %s", v, strings.Join(k.Options, ", "))
	case "bool":
		switch strings.ToLower(v) {
		case "true", "yes", "on", "1":
			return "true", nil
		case "false", "no", "off", "0":
			return "false", nil
		}
		return "", fmt.Errorf("%q is not true or false", v)
	case "color":
		return normalizeColor(v)
	case "date":
		if _, err := time.Parse("2006-01-02", v); err != nil {
			return "", fmt.Errorf("%q is not a date (YYYY-MM-DD)", v)
		}
	case "datetime":
		if _, err := time.Parse(time.RFC3339, v); err != nil {
			return "", fmt.Errorf("%q is not a date and time (RFC 3339)", v)
		}
	}
	return v, nil
}

// normalizeColor accepts #rgb, #rrggbb (with or without #) and a CSS color
// name, and returns #rrggbb in lower case.
func normalizeColor(v string) (string, error) {
	if hex, ok := colorNames[strings.ToLower(strings.ReplaceAll(v, " ", ""))]; ok {
		return hex, nil
	}
	h := strings.ToLower(strings.TrimPrefix(v, "#"))
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 {
		return "", fmt.Errorf("%q is not a color (#rrggbb, #rgb or a name like orange)", v)
	}
	if _, err := strconv.ParseUint(h, 16, 32); err != nil {
		return "", fmt.Errorf("%q is not a color (#rrggbb, #rgb or a name like orange)", v)
	}
	return "#" + h, nil
}

// colorNames are the CSS basic and common extended color names, plus the
// whites people ask lights for.
var colorNames = map[string]string{
	"black": "#000000", "white": "#ffffff", "red": "#ff0000", "green": "#008000",
	"blue": "#0000ff", "yellow": "#ffff00", "orange": "#ffa500", "purple": "#800080",
	"pink": "#ffc0cb", "cyan": "#00ffff", "magenta": "#ff00ff", "lime": "#00ff00",
	"teal": "#008080", "navy": "#000080", "maroon": "#800000", "olive": "#808000",
	"silver": "#c0c0c0", "gray": "#808080", "grey": "#808080", "gold": "#ffd700",
	"coral": "#ff7f50", "salmon": "#fa8072", "violet": "#ee82ee", "indigo": "#4b0082",
	"turquoise": "#40e0d0", "lavender": "#e6e6fa", "crimson": "#dc143c", "amber": "#ffbf00",
	"warmwhite": "#ffd6aa", "coolwhite": "#f4fbff", "daylight": "#fffaf4",
}
