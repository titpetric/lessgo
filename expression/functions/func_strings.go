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

	// Track the original quote character
	quoteChar := "\""
	if len(str) >= 2 && str[0] == '\'' && str[len(str)-1] == '\'' {
		quoteChar = "'"
	}

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

	// Check for global and case-insensitive flags
	global := true // default: replace all occurrences
	caseInsensitive := false
	if len(flags) > 0 {
		f := strings.TrimSpace(flags[0])
		// Remove quotes from flag
		if len(f) >= 2 && ((f[0] == '"' && f[len(f)-1] == '"') ||
			(f[0] == '\'' && f[len(f)-1] == '\'')) {
			f = f[1 : len(f)-1]
		}
		// Check flags: 'g' = global (on by default), 'i' = case-insensitive
		if strings.Contains(f, "i") {
			caseInsensitive = true
		}
		// 'g' flag means global, no 'g' means just first match
		if !strings.Contains(f, "g") {
			global = false
		}
	}

	var result string
	// Use regex replace if pattern looks like regex (contains regex metacharacters)
	// Otherwise do simple string replacement
	if hasRegexMetacharacters(pattern) || caseInsensitive {
		patternStr := pattern
		if caseInsensitive {
			patternStr = "(?i)" + patternStr
		}
		regex, err := regexp.Compile(patternStr)
		if err != nil {
			// If regex is invalid, fall back to string replacement
			result = stringReplace(str, pattern, replacement, global)
		} else if global {
			result = regex.ReplaceAllString(str, replacement)
		} else {
			result = regexReplaceFirst(regex, str, replacement)
		}
	} else {
		result = stringReplace(str, pattern, replacement, global)
	}

	// Return with quotes preserved
	return quoteChar + result + quoteChar
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
