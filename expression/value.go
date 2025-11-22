package expression

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Cache compiled regex for parsing numeric values with units
	numericUnitRegex = regexp.MustCompile(`^(-?[\d.]+)([\w%]*)$`)
)

// Value represents a computed value in LESS (number with unit, color, etc)
type Value struct {
	Number       float64 // numeric value (for numeric values)
	Unit         string  // unit (px, %, em, etc) or empty
	OriginalUnit string  // original unit before conversions (e.g., % before decimal conversion)
	Color        *Color  // color value (for color values)
	Raw          string  // original raw string (for debugging)
}

// NewValue creates a value from a number and unit
func NewValue(num float64, unit string) *Value {
	return &Value{
		Number: num,
		Unit:   unit,
		Raw:    fmt.Sprintf("%g%s", num, unit),
	}
}

// NewColorValue creates a value from a Color
func NewColorValue(c *Color) *Value {
	return &Value{
		Color: c,
		Raw:   c.String(),
	}
}

// Parse parses a string into a Value
// Examples: "10px", "50%", "1.5em", "3", "#3498db", "rgb(52, 152, 219)"
func Parse(s string) (*Value, error) {
	s = strings.TrimSpace(s)

	// Handle quoted strings
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return &Value{
			Raw: s,
		}, nil
	}

	// Try to parse as color first (before bare keyword check)
	if isColorLike(s) {
		c, err := ParseColor(s)
		if err == nil {
			return NewColorValue(c), nil
		}
		// If color parsing fails, fall through to other parsing
	}

	// Handle bare keywords (like colors, booleans, keywords)
	// If it doesn't start with a number or -, treat it as a keyword value
	if !strings.HasPrefix(s, "-") && (len(s) == 0 || !isDigit(rune(s[0]))) {
		// It's a keyword/bare word, store as raw value
		return &Value{
			Raw: s,
		}, nil
	}

	// Match numeric part and unit part
	// Unit part can contain letters, %, or other valid unit chars, but NOT spaces or function calls
	// This prevents "1px solid rgb(...)" from being parsed as a single value
	matches := numericUnitRegex.FindStringSubmatch(s)

	if len(matches) != 3 {
		// If it doesn't match simple number+unit pattern, treat as raw value (keyword)
		// This handles compound values like "1px solid rgb(...)"
		return &Value{
			Raw: s,
		}, nil
	}

	numStr := matches[1]
	unit := strings.TrimSpace(matches[2])

	// Parse the number
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number: %s", numStr)
	}

	// Store original unit before conversion
	originalUnit := unit

	// Handle percentage -> decimal conversion
	if unit == "%" {
		num = num / 100.0
		unit = ""
	}

	return &Value{
		Number:       num,
		Unit:         unit,
		OriginalUnit: originalUnit,
		Raw:          s,
	}, nil
}

// isDigit checks if a rune is a digit
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// isColorLike checks if a string looks like a color
func isColorLike(s string) bool {
	return strings.HasPrefix(s, "#") ||
		strings.HasPrefix(s, "rgb(") ||
		strings.HasPrefix(s, "rgba(") ||
		strings.HasPrefix(s, "hsl(") ||
		strings.HasPrefix(s, "hsla(")
}

// String returns the string representation
func (v *Value) String() string {
	if v == nil {
		return ""
	}

	if v.Color != nil {
		return v.Color.String()
	}

	// If Raw is set and Number is 0 (for lists and non-numeric values)
	if v.Raw != "" && v.Number == 0 && v.Unit == "" {
		return v.Raw
	}

	// Use OriginalUnit if present (for percentages converted to decimal)
	unit := v.Unit
	if unit == "" && v.OriginalUnit != "" {
		unit = v.OriginalUnit
		// For percentage, reconstruct the percentage value from the decimal
		if unit == "%" {
			// Keep 8 decimal places for percentage values (matching lessc precision)
			percentVal := v.Number * 100
			formatted := fmt.Sprintf("%.8f", percentVal)
			// Remove trailing zeros but keep integer format if possible
			formatted = strings.TrimRight(formatted, "0")
			formatted = strings.TrimRight(formatted, ".")
			return formatted + "%"
		}
	}

	if unit == "" {
		return fmt.Sprintf("%g", trimFloat(v.Number))
	}
	return fmt.Sprintf("%g%s", trimFloat(v.Number), unit)
}

// trimFloat limits floating point precision to 9 significant figures (matching lessc)
func trimFloat(num float64) float64 {
	// Use strconv to format and parse back to limit precision
	str := fmt.Sprintf("%.9g", num)
	parsed, _ := strconv.ParseFloat(str, 64)
	return parsed
}

// Add adds two values, preserving unit from the left operand
func (v *Value) Add(other *Value) (*Value, error) {
	if v.Unit != other.Unit {
		if v.Unit != "" && other.Unit != "" {
			return nil, fmt.Errorf("cannot add %s and %s", v.Unit, other.Unit)
		}
	}

	unit := v.Unit
	if unit == "" {
		unit = other.Unit
	}

	return NewValue(v.Number+other.Number, unit), nil
}

// Subtract subtracts other from v
func (v *Value) Subtract(other *Value) (*Value, error) {
	if v.Unit != other.Unit {
		if v.Unit != "" && other.Unit != "" {
			return nil, fmt.Errorf("cannot subtract %s from %s", other.Unit, v.Unit)
		}
	}

	unit := v.Unit
	if unit == "" {
		unit = other.Unit
	}

	return NewValue(v.Number-other.Number, unit), nil
}

// Multiply multiplies two values
func (v *Value) Multiply(other *Value) (*Value, error) {
	// When multiplying: (5px) * (10) = 50px
	// When multiplying: (5) * (10px) = 50px
	// When multiplying: (5px) * (10px) = error (no unit)

	if v.Unit != "" && other.Unit != "" {
		return nil, fmt.Errorf("cannot multiply %s by %s", v.Unit, other.Unit)
	}

	unit := v.Unit
	if unit == "" {
		unit = other.Unit
	}

	return NewValue(v.Number*other.Number, unit), nil
}

// Divide divides v by other
func (v *Value) Divide(other *Value) (*Value, error) {
	if other.Number == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	// When dividing: (50px) / (5) = 10px
	// When dividing: (50px) / (10px) = 5 (no unit)

	var unit string
	if other.Unit == "" {
		unit = v.Unit
	} else if v.Unit == other.Unit {
		unit = ""
	} else {
		return nil, fmt.Errorf("cannot divide %s by %s", v.Unit, other.Unit)
	}

	return NewValue(v.Number/other.Number, unit), nil
}
