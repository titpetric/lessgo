package functions

import (
	"regexp"
	"strings"
)

// Replace replaces a pattern in a string with a replacement string
// replace(string, pattern, replacement, flags?)
func Replace(str, pattern, replacement string, flags ...string) string {
	str = strings.TrimSpace(str)
	pattern = strings.TrimSpace(pattern)
	replacement = strings.TrimSpace(replacement)

	// Remove quotes from string if present
	if len(str) >= 2 && ((str[0] == '"' && str[len(str)-1] == '"') ||
		(str[0] == '\'' && str[len(str)-1] == '\'')) {
		str = str[1 : len(str)-1]
	}

	// Remove quotes from pattern if present
	if len(pattern) >= 2 && ((pattern[0] == '"' && pattern[len(pattern)-1] == '"') ||
		(pattern[0] == '\'' && pattern[len(pattern)-1] == '\'')) {
		pattern = pattern[1 : len(pattern)-1]
	}

	// Remove quotes from replacement if present
	if len(replacement) >= 2 && ((replacement[0] == '"' && replacement[len(replacement)-1] == '"') ||
		(replacement[0] == '\'' && replacement[len(replacement)-1] == '\'')) {
		replacement = replacement[1 : len(replacement)-1]
	}

	// Check for global flag (default is to replace all)
	global := true
	if len(flags) > 0 {
		f := strings.TrimSpace(flags[0])
		// Remove quotes from flag
		if len(f) >= 2 && ((f[0] == '"' && f[len(f)-1] == '"') ||
			(f[0] == '\'' && f[len(f)-1] == '\'')) {
			f = f[1 : len(f)-1]
		}
		global = !strings.Contains(f, "g")
	}

	// Use regex replace if pattern looks like regex (contains regex metacharacters)
	// Otherwise do simple string replacement
	if hasRegexMetacharacters(pattern) {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			// If regex is invalid, fall back to string replacement
			return stringReplace(str, pattern, replacement, global)
		}
		if global {
			return regex.ReplaceAllString(str, replacement)
		} else {
			return regexReplaceFirst(regex, str, replacement)
		}
	} else {
		return stringReplace(str, pattern, replacement, global)
	}
}

// hasRegexMetacharacters checks if a string contains regex metacharacters
func hasRegexMetacharacters(s string) bool {
	regexMetachars := []string{".", "*", "+", "?", "^", "$", "|", "(", ")", "[", "]", "{", "}"}
	for _, ch := range regexMetachars {
		if strings.Contains(s, ch) {
			return true
		}
	}
	return false
}

// stringReplace performs simple string replacement
func stringReplace(str, pattern, replacement string, global bool) string {
	if global {
		return strings.ReplaceAll(str, pattern, replacement)
	} else {
		return strings.Replace(str, pattern, replacement, 1)
	}
}

// regexReplaceFirst replaces the first occurrence using regex
func regexReplaceFirst(re *regexp.Regexp, str string, repl string) string {
	loc := re.FindStringIndex(str)
	if loc == nil {
		return str
	}
	return str[:loc[0]] + repl + str[loc[1]:]
}
