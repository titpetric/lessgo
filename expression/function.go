package expression

import (
	"fmt"
	"regexp"

	"github.com/titpetric/lessgo/internal/strings"
)

var (
	// Cache compiled regex for parsing function calls
	funcCallRegex = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*\((.*)\)$`)
)

// ParseFunctionCall parses a string like "ceil(10px)" into function name and args
func ParseFunctionCall(expr string) (string, []*Value, error) {
	expr = strings.TrimSpace(expr)

	// Handle LESS format function syntax: %(...) is shorthand for format(...)
	if strings.HasPrefix(expr, "%") {
		expr = "format" + expr[1:]
	}

	// Match function pattern: name(args)
	// Allow hyphens in function names (e.g., get-unit, is-defined)
	matches := funcCallRegex.FindStringSubmatch(expr)

	if len(matches) != 3 {
		return "", nil, fmt.Errorf("not a function call: %s", expr)
	}

	funcName := matches[1]
	argsStr := matches[2]

	// Parse arguments (comma-separated)
	var args []*Value
	if argsStr != "" {
		argParts := splitArgs(argsStr)
		for _, argStr := range argParts {
			argStr = strings.TrimSpace(argStr)
			if argStr == "" {
				continue
			}
			v, err := Parse(argStr)
			if err != nil {
				// If it fails to parse as a value, it might be a list or complex expression
				// Store it as a raw value
				v = &Value{Raw: argStr}
			}
			args = append(args, v)
		}
	}

	return funcName, args, nil
}

// splitArgs splits function arguments by comma, respecting nesting and quotes
func splitArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(argsStr); i++ {
		ch := argsStr[i]

		// Handle quotes
		if (ch == '"' || ch == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = ch
			current.WriteByte(ch)
		} else if ch == quoteChar && inQuotes {
			inQuotes = false
			current.WriteByte(ch)
		} else if inQuotes {
			// Inside quotes, keep everything
			current.WriteByte(ch)
		} else {
			// Not in quotes
			switch ch {
			case '(':
				depth++
				current.WriteByte(ch)
			case ')':
				depth--
				current.WriteByte(ch)
			case ',':
				if depth == 0 {
					args = append(args, current.String())
					current.Reset()
				} else {
					current.WriteByte(ch)
				}
			default:
				current.WriteByte(ch)
			}
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// IsFunctionCall checks if a string is a function call
func IsFunctionCall(expr string) bool {
	idx := strings.Index(expr, "(")
	if idx == -1 {
		return false
	}

	name := expr[:idx]

	// LESS format function syntax: %(...) is shorthand for format(...)
	if strings.TrimSpace(name) == "%" {
		return true
	}

	return IsRegisteredFunction(name)
}

// IsRegisteredFunction checks if a function name is registered
func IsRegisteredFunction(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	_, ok := funcMap[name]
	return ok
}
