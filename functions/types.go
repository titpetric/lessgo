package functions

import (
	"regexp"
	"strconv"
	"strings"
)

// IsNumber checks if a value is a number (with optional unit)
func IsNumber(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	// Try to parse the numeric part
	i := 0
	if i < len(value) && (value[i] == '-' || value[i] == '+') {
		i++
	}

	hasDigit := false
	for i < len(value) && (value[i] >= '0' && value[i] <= '9' || value[i] == '.') {
		if value[i] >= '0' && value[i] <= '9' {
			hasDigit = true
		}
		i++
	}

	// If we found digits, it's a number (units can follow)
	return hasDigit
}

// IsString checks if a value is a string literal
func IsString(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return false
	}

	// Check for quotes
	return (value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'')
}

// IsColor checks if a value is a color
func IsColor(value string) bool {
	value = strings.TrimSpace(value)

	// Check for hex colors
	if strings.HasPrefix(value, "#") {
		hex := strings.TrimPrefix(value, "#")
		if len(hex) == 3 || len(hex) == 4 || len(hex) == 6 || len(hex) == 8 {
			// Check if all characters are valid hex digits
			for _, ch := range hex {
				if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
					return false
				}
			}
			return true
		}
	}

	// Check for rgb/rgba
	if strings.HasPrefix(value, "rgb") {
		return strings.HasPrefix(value, "rgb(") || strings.HasPrefix(value, "rgba(")
	}

	// Check for hsl/hsla
	if strings.HasPrefix(value, "hsl") {
		return strings.HasPrefix(value, "hsl(") || strings.HasPrefix(value, "hsla(")
	}

	// Check for named colors (CSS color keywords)
	namedColors := map[string]bool{
		"red": true, "green": true, "blue": true, "yellow": true, "orange": true,
		"purple": true, "pink": true, "cyan": true, "magenta": true, "white": true,
		"black": true, "gray": true, "grey": true, "silver": true, "gold": true,
		"maroon": true, "navy": true, "teal": true, "olive": true, "lime": true,
		"aqua": true, "fuchsia": true, "indigo": true, "turquoise": true, "khaki": true,
		"tomato": true, "coral": true, "salmon": true, "chocolate": true, "peru": true,
		"wheat": true, "tan": true, "beige": true, "ivory": true, "bisque": true,
		"aliceblue": true, "antiquewhite": true, "aquamarine": true, "azure": true,
		"brown": true, "burlywood": true, "cadetblue": true, "chartreuse": true,
		"darkblue": true, "darkcyan": true, "darkgray": true, "darkgreen": true,
		"darkkhaki": true, "darkmagenta": true, "darkolivegreen": true, "darkorange": true,
		"darkorchid": true, "darkred": true, "darksalmon": true, "darkseagreen": true,
		"darkslateblue": true, "darkslategray": true, "darkturquoise": true, "darkviolet": true,
		"deeppink": true, "deepskyblue": true, "dimgray": true, "dodgerblue": true,
		"firebrick": true, "floralwhite": true, "forestgreen": true, "gainsboro": true,
		"ghostwhite": true, "goldenrod": true, "honeydew": true, "hotpink": true,
		"indianred": true, "lavender": true, "lavenderblush": true, "lawngreen": true,
		"lemonchiffon": true, "lightblue": true, "lightcoral": true, "lightcyan": true,
		"lightgoldenrodyellow": true, "lightgray": true, "lightgreen": true, "lightpink": true,
		"lightsalmon": true, "lightseagreen": true, "lightskyblue": true, "lightslategray": true,
		"lightsteelblue": true, "lightyellow": true, "mediumaquamarine": true, "mediumblue": true,
		"mediumorchid": true, "mediumpurple": true, "mediumseagreen": true, "mediumslateblue": true,
		"mediumspringgreen": true, "mediumturquoise": true, "mediumvioletred": true, "midnightblue": true,
		"mintcream": true, "mistyrose": true, "moccasin": true, "navajowhite": true, "oldlace": true,
		"olivedrab": true, "orangered": true, "orchid": true, "palegoldenrod": true, "palegreen": true,
		"paleturquoise": true, "palevioletred": true, "papayawhip": true, "peachpuff": true,
		"plum": true, "powderblue": true, "rosybrown": true, "royalblue": true, "saddlebrown": true,
		"sandybrown": true, "seagreen": true, "seashell": true, "sienna": true, "skyblue": true,
		"slateblue": true, "slategray": true, "snow": true, "springgreen": true, "steelblue": true,
		"thistle": true, "violet": true, "whitesmoke": true, "yellowgreen": true,
	}

	return namedColors[value]
}

// IsKeyword checks if a value is a CSS keyword
func IsKeyword(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	// Common CSS keywords
	keywords := map[string]bool{
		"inherit":      true,
		"initial":      true,
		"unset":        true,
		"revert":       true,
		"auto":         true,
		"none":         true,
		"transparent":  true,
		"solid":        true,
		"dashed":       true,
		"dotted":       true,
		"double":       true,
		"groove":       true,
		"ridge":        true,
		"inset":        true,
		"outset":       true,
		"left":         true,
		"right":        true,
		"center":       true,
		"top":          true,
		"bottom":       true,
		"middle":       true,
		"absolute":     true,
		"relative":     true,
		"fixed":        true,
		"static":       true,
		"block":        true,
		"inline":       true,
		"inline-block": true,
		"flex":         true,
		"grid":         true,
		"bold":         true,
		"italic":       true,
		"normal":       true,
	}

	return keywords[value]
}

// IsURL checks if a value is a URL
func IsURL(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasPrefix(value, "url(") && strings.HasSuffix(value, ")")
}

// IsPixel checks if a value is in pixels
func IsPixel(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasSuffix(value, "px") && IsNumber(strings.TrimSuffix(value, "px"))
}

// IsEm checks if a value is in em units
func IsEm(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasSuffix(value, "em") && IsNumber(strings.TrimSuffix(value, "em"))
}

// IsPercentage checks if a value is a percentage
func IsPercentage(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasSuffix(value, "%") && IsNumber(strings.TrimSuffix(value, "%"))
}

// IsUnit checks if a value has the specified unit
func IsUnit(value string, unit string) bool {
	value = strings.TrimSpace(value)
	unit = strings.TrimSpace(unit)

	// If unit is a string literal with quotes, remove them
	if len(unit) >= 2 && ((unit[0] == '"' && unit[len(unit)-1] == '"') ||
		(unit[0] == '\'' && unit[len(unit)-1] == '\'')) {
		unit = unit[1 : len(unit)-1]
	}

	// Extract unit from value
	actualUnit := extractUnit(value)
	return actualUnit == unit
}

// IsRuleset checks if a value is a ruleset/object (stored in a variable)
// This would need context from the renderer to properly determine
func IsRuleset(value string) bool {
	// Check if it looks like a stored ruleset reference
	value = strings.TrimSpace(value)

	// Rulesets are stored in variables, so check if it's a variable reference
	if strings.HasPrefix(value, "@") && !strings.Contains(value, "{") {
		return false // It's a variable reference, not a ruleset itself
	}

	// In LESS, rulesets are enclosed in braces
	return strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")
}

// IsDefined checks if a variable is defined (this requires context)
// For now, return false as this needs renderer context
func IsDefined(varName string) bool {
	// This function needs access to the variable scope
	// which is not available in this package
	return false
}

// IsNumberFunction is the exposed function for isnumber()
func IsNumberFunction(value string) string {
	if IsNumber(value) {
		return "true"
	}
	return "false"
}

// IsStringFunction is the exposed function for isstring()
func IsStringFunction(value string) string {
	if IsString(value) {
		return "true"
	}
	return "false"
}

// IsColorFunction is the exposed function for iscolor()
func IsColorFunction(value string) string {
	if IsColor(value) {
		return "true"
	}
	return "false"
}

// IsKeywordFunction is the exposed function for iskeyword()
func IsKeywordFunction(value string) string {
	if IsKeyword(value) {
		return "true"
	}
	return "false"
}

// IsURLFunction is the exposed function for isurl()
func IsURLFunction(value string) string {
	if IsURL(value) {
		return "true"
	}
	return "false"
}

// IsPixelFunction is the exposed function for ispixel()
func IsPixelFunction(value string) string {
	if IsPixel(value) {
		return "true"
	}
	return "false"
}

// IsEmFunction is the exposed function for isem()
func IsEmFunction(value string) string {
	if IsEm(value) {
		return "true"
	}
	return "false"
}

// IsPercentageFunction is the exposed function for ispercentage()
func IsPercentageFunction(value string) string {
	if IsPercentage(value) {
		return "true"
	}
	return "false"
}

// IsUnitFunction is the exposed function for isunit()
func IsUnitFunction(value string, unit string) string {
	if IsUnit(value, unit) {
		return "true"
	}
	return "false"
}

// IsRulesetFunction is the exposed function for isruleset()
func IsRulesetFunction(value string) string {
	if IsRuleset(value) {
		return "true"
	}
	return "false"
}

// IsList checks if a value is a list (comma or space-separated values)
func IsList(value string) bool {
	value = strings.TrimSpace(value)

	// Comma-separated list
	if strings.Contains(value, ",") {
		return true
	}

	// Space-separated list (more than one space-separated value)
	parts := strings.Fields(value)
	return len(parts) > 1
}

// IsListFunction is the exposed function for islist()
func IsListFunction(value string) string {
	if IsList(value) {
		return "true"
	}
	return "false"
}

// Boolean converts a value to a boolean (only true for the literal keyword 'true')
// In LESS, boolean() returns true ONLY for the keyword 'true', and false for everything else
func Boolean(value string) string {
	value = strings.TrimSpace(value)

	// Only true if the value is literally the keyword 'true'
	if value == "true" {
		return "true"
	}

	// Everything else is false (including numbers, other keywords, etc.)
	return "false"
}

// Length returns the length of a list/string value
func Length(value string) string {
	value = strings.TrimSpace(value)

	// If it's a string, return the character count
	if IsString(value) {
		// Remove quotes
		if len(value) >= 2 {
			value = value[1 : len(value)-1]
		}
		return strconv.Itoa(len(value))
	}

	// For space or comma-separated lists, count items
	// Split by space first
	items := strings.Fields(value)
	if len(items) > 1 {
		return strconv.Itoa(len(items))
	}

	// Split by comma
	items = strings.Split(value, ",")
	if len(items) > 1 {
		return strconv.Itoa(len(items))
	}

	// Single item
	return "1"
}

// Extract gets an item from a list by index (1-based)
func Extract(list string, index string) string {
	list = strings.TrimSpace(list)
	idx, err := strconv.Atoi(strings.TrimSpace(index))
	if err != nil || idx < 1 {
		return ""
	}

	// Try space-separated first
	items := strings.Fields(list)
	if len(items) >= idx {
		return items[idx-1]
	}

	// Try comma-separated
	items = strings.Split(list, ",")
	if len(items) >= idx {
		return strings.TrimSpace(items[idx-1])
	}

	return ""
}

// Range generates a comma-separated list of numbers from start to end
func Range(start string, end string, step ...string) string {
	s, _ := strconv.ParseFloat(strings.TrimSpace(start), 64)
	e, _ := strconv.ParseFloat(strings.TrimSpace(end), 64)

	stepVal := 1.0
	if len(step) > 0 && step[0] != "" {
		stepVal, _ = strconv.ParseFloat(strings.TrimSpace(step[0]), 64)
	}

	if stepVal == 0 {
		stepVal = 1
	}

	var result []string
	if s <= e {
		for i := s; i <= e; i += stepVal {
			if i == float64(int64(i)) {
				result = append(result, strconv.FormatInt(int64(i), 10))
			} else {
				result = append(result, strconv.FormatFloat(i, 'f', -1, 64))
			}
		}
	} else {
		for i := s; i >= e; i -= stepVal {
			if i == float64(int64(i)) {
				result = append(result, strconv.FormatInt(int64(i), 10))
			} else {
				result = append(result, strconv.FormatFloat(i, 'f', -1, 64))
			}
		}
	}

	return strings.Join(result, ", ")
}

// Escape URL-encodes a string (using strict LESS escaping rules)
// LESS escape() does NOT escape all special characters - only specific ones
func Escape(str string) string {
	// Remove quotes if present
	str = strings.TrimSpace(str)
	if len(str) >= 2 && ((str[0] == '"' && str[len(str)-1] == '"') ||
		(str[0] == '\'' && str[len(str)-1] == '\'')) {
		str = str[1 : len(str)-1]
	}

	// URL encode only specific characters (matching LESS behavior)
	replacer := strings.NewReplacer(
		" ", "%20",
		"\"", "%22",
		"#", "%23",
		"$", "%24",
		"%", "%25",
		"&", "%26",
		"'", "%27",
		"(", "%28",
		")", "%29",
		"*", "%2A",
		"+", "%2B",
		",", "%2C",
		"/", "%2F",
		":", "%3A",
		";", "%3B",
		"<", "%3C",
		"=", "%3D",
		">", "%3E",
		"?", "%3F",
		"@", "%40",
		"[", "%5B",
		"\\", "%5C",
		"]", "%5D",
		"^", "%5E",
		"`", "%60",
		"{", "%7B",
		"|", "%7C",
		"}", "%7D",
		"~", "%7E",
	)

	return replacer.Replace(str)
}

// E returns the escaped string (similar to escape but used in LESS for removing quotes)
func E(str string) string {
	str = strings.TrimSpace(str)

	// Remove quotes if present
	if len(str) >= 2 && ((str[0] == '"' && str[len(str)-1] == '"') ||
		(str[0] == '\'' && str[len(str)-1] == '\'')) {
		str = str[1 : len(str)-1]
	}

	return str
}

// Format string - simple % formatting similar to LESS
func Format(format string, args ...string) string {
	format = strings.TrimSpace(format)

	// Remove quotes if present
	if len(format) >= 2 && ((format[0] == '"' && format[len(format)-1] == '"') ||
		(format[0] == '\'' && format[len(format)-1] == '\'')) {
		format = format[1 : len(format)-1]
	}

	// Replace %s with arguments in order
	argIdx := 0
	result := regexp.MustCompile(`%[sd]`).ReplaceAllStringFunc(format, func(match string) string {
		if argIdx < len(args) {
			arg := args[argIdx]
			argIdx++
			return strings.TrimSpace(arg)
		}
		return match
	})

	return result
}
