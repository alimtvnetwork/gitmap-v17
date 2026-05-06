package cmd

import (
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/render"
)

// Flag tokens for pretty-mode parsing. Two synonym pairs are accepted so
// the flag is comfortable to type either way: `--pretty[=bool]` follows
// the standard Go flag style, while `--no-pretty` matches the convention
// many CLIs use for explicit negation.
const (
	flagPrettyPositive = "--pretty"
	flagPrettyNegative = "--no-pretty"
)

// ParsePrettyFlag pulls --pretty / --no-pretty out of args and returns
// the cleaned slice + the resolved render.PrettyMode. Accepted forms:
//
//	--pretty            → PrettyOn
//	--pretty=true       → PrettyOn
//	--pretty=on         → PrettyOn
//	--pretty=1          → PrettyOn
//	--pretty=false      → PrettyOff
//	--pretty=off        → PrettyOff
//	--pretty=0          → PrettyOff
//	--pretty=auto       → PrettyAuto (explicit reset)
//	--no-pretty         → PrettyOff
//
// When the same flag is repeated, the **last** occurrence wins (matches
// stdlib flag.Parse semantics). When neither appears, the returned mode
// is PrettyAuto so callers can rely on Decide()'s default ladder.
//
// Unrecognized values fall through to PrettyAuto and the token is left
// in place so the downstream parser can produce a meaningful error.
func ParsePrettyFlag(args []string) ([]string, render.PrettyMode) {
	mode := render.PrettyAuto
	out := make([]string, 0, len(args))
	for _, a := range args {
		token, value, hasValue := splitPrettyToken(a)
		switch token {
		case flagPrettyPositive:
			mode = resolvePositivePretty(value, hasValue, mode, &out, a)
		case flagPrettyNegative:
			mode = render.PrettyOff
		default:
			out = append(out, a)
		}
	}

	return out, mode
}

// splitPrettyToken splits "--pretty=value" into ("--pretty", "value", true)
// and "--pretty" into ("--pretty", "", false). Anything else returns the
// original token in slot 0 with hasValue=false so the caller can passthrough.
func splitPrettyToken(arg string) (token, value string, hasValue bool) {
	if !strings.HasPrefix(arg, "--pretty") && !strings.HasPrefix(arg, "--no-pretty") {
		return arg, "", false
	}
	if eq := strings.IndexByte(arg, '='); eq >= 0 {
		return arg[:eq], arg[eq+1:], true
	}

	return arg, "", false
}

// resolvePositivePretty maps a "--pretty[=value]" occurrence to a
// PrettyMode. Falls back to keeping the original token in `out` when
// the value is unrecognized so flag.Parse downstream can report it.
func resolvePositivePretty(value string, hasValue bool, current render.PrettyMode, out *[]string, original string) render.PrettyMode {
	if !hasValue {
		return render.PrettyOn
	}
	switch strings.ToLower(value) {
	case "1", "t", "true", "on", "yes", "y":
		return render.PrettyOn
	case "0", "f", "false", "off", "no", "n":
		return render.PrettyOff
	case "auto", "":
		return render.PrettyAuto
	}
	*out = append(*out, original)

	return current
}
