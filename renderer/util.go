package renderer

import "github.com/titpetric/lessgo/internal/strings"

// isValueChar checks if a character can be part of a value
func isValueChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '%' || r == '#'
}

// isLikeList checks if a string looks like a list
func isLikeList(s string) bool {
	return strings.Contains(s, ",") && (strings.Contains(s, "\"") || strings.Contains(s, "'"))
}

// isVarChar checks if a character is valid in a variable name
func isVarChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-'
}

// normalizeSelectorSpacing normalizes spaces around CSS combinators (+, >, ~)
// Ensures consistent spacing: " + " regardless of input spacing
func normalizeSelectorSpacing(sel string) string {
	// First, remove all spaces around combinators
	sel = strings.ReplaceAll(sel, " + ", "+")
	sel = strings.ReplaceAll(sel, " > ", ">")
	sel = strings.ReplaceAll(sel, " ~ ", "~")

	// Then add consistent spacing around combinators
	sel = strings.ReplaceAll(sel, "+", " + ")
	sel = strings.ReplaceAll(sel, ">", " > ")
	sel = strings.ReplaceAll(sel, "~", " ~ ")

	// Clean up any multiple spaces that may have been created
	for strings.Contains(sel, "  ") {
		sel = strings.ReplaceAll(sel, "  ", " ")
	}

	return sel
}

// selector will combine a parent and child selector.
func selector(parent, child string) string {
	child = normalizeSelectorSpacing(child)
	if parent == "" {
		return child
	}
	if strings.HasPrefix(child, "&") {
		return normalizeSelectorSpacing(parent + child[1:])
	}
	return parent + " " + child
}
