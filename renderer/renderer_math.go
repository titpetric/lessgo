package renderer

import (
	"math"
	"strconv"
	"strings"
)

// Ceil returns the smallest integer >= x
func Ceil(value string) string {
	num := parseNumber(value)
	result := math.Ceil(num)
	return formatNumber(result, value)
}

// Floor returns the largest integer <= x
func Floor(value string) string {
	num := parseNumber(value)
	result := math.Floor(num)
	return formatNumber(result, value)
}

// Round returns the nearest integer
func Round(value string) string {
	num := parseNumber(value)
	result := math.Round(num)
	return formatNumber(result, value)
}

// Abs returns the absolute value
func Abs(value string) string {
	num := parseNumber(value)
	result := math.Abs(num)
	return formatNumber(result, value)
}

// Sqrt returns the square root
func Sqrt(value string) string {
	num := parseNumber(value)
	result := math.Sqrt(num)
	return formatNumber(result, value)
}

// Pow returns base to the power of exponent
func Pow(base, exponent string) string {
	b := parseNumber(base)
	e := parseNumber(exponent)
	result := math.Pow(b, e)
	return formatNumber(result, base)
}

// Min returns the minimum of the provided values
func Min(values ...string) string {
	if len(values) == 0 {
		return "0"
	}

	min := math.MaxFloat64
	minUnit := ""

	for _, val := range values {
		num := parseNumber(val)
		if num < min {
			min = num
			minUnit = extractUnit(val)
		}
	}

	return formatNumberWithUnit(min, minUnit)
}

// Max returns the maximum of the provided values
func Max(values ...string) string {
	if len(values) == 0 {
		return "0"
	}

	max := -math.MaxFloat64
	maxUnit := ""

	for _, val := range values {
		num := parseNumber(val)
		if num > max {
			max = num
			maxUnit = extractUnit(val)
		}
	}

	return formatNumberWithUnit(max, maxUnit)
}

// parseNumber extracts the numeric part from a value string
func parseNumber(value string) float64 {
	value = strings.TrimSpace(value)
	// Remove unit suffix
	i := 0
	for i < len(value) && (value[i] == '-' || value[i] == '+' || (value[i] >= '0' && value[i] <= '9') || value[i] == '.') {
		i++
	}
	numStr := value[:i]
	num, _ := strconv.ParseFloat(numStr, 64)
	return num
}

// extractUnit extracts the unit part from a value string
func extractUnit(value string) string {
	value = strings.TrimSpace(value)
	// Skip the number part
	i := 0
	for i < len(value) && (value[i] == '-' || value[i] == '+' || (value[i] >= '0' && value[i] <= '9') || value[i] == '.') {
		i++
	}
	return strings.TrimSpace(value[i:])
}

// formatNumber formats a number, preserving the unit from the original value
func formatNumber(result float64, original string) string {
	unit := extractUnit(original)
	return formatNumberWithUnit(result, unit)
}

// formatNumberWithUnit formats a number with a unit
func formatNumberWithUnit(num float64, unit string) string {
	// Handle integer representation if the result is a whole number
	if num == math.Floor(num) && num >= -1e15 && num <= 1e15 {
		if unit == "" {
			return strconv.FormatInt(int64(num), 10)
		}
		return strconv.FormatInt(int64(num), 10) + unit
	}

	// Otherwise use the decimal representation
	result := strconv.FormatFloat(num, 'f', -1, 64)
	// Remove trailing zeros after decimal point
	if strings.Contains(result, ".") {
		result = strings.TrimRight(result, "0")
		result = strings.TrimRight(result, ".")
	}
	return result + unit
}

// Mod returns the remainder of a / b
func Mod(a, b string) string {
	aNum := parseNumber(a)
	bNum := parseNumber(b)

	if bNum == 0 {
		return "0" // Avoid division by zero
	}

	result := math.Mod(aNum, bNum)
	return formatNumber(result, a)
}

// Sin returns the sine of a number (in radians)
func Sin(value string) string {
	num := parseNumber(value)
	result := math.Sin(num)
	return formatNumber(result, value)
}

// Cos returns the cosine of a number (in radians)
func Cos(value string) string {
	num := parseNumber(value)
	result := math.Cos(num)
	return formatNumber(result, value)
}

// Tan returns the tangent of a number (in radians)
func Tan(value string) string {
	num := parseNumber(value)
	result := math.Tan(num)
	return formatNumber(result, value)
}

// Asin returns the arcsine of a number (in radians)
func Asin(value string) string {
	num := parseNumber(value)
	result := math.Asin(num)
	return formatNumber(result, value)
}

// Acos returns the arccosine of a number (in radians)
func Acos(value string) string {
	num := parseNumber(value)
	result := math.Acos(num)
	return formatNumber(result, value)
}

// Atan returns the arctangent of a number (in radians)
func Atan(value string) string {
	num := parseNumber(value)
	result := math.Atan(num)
	return formatNumber(result, value)
}

// Pi returns the value of pi
func Pi() string {
	// LESS limits pi() to 8 decimal places
	rounded := math.Round(math.Pi*100000000) / 100000000
	result := strconv.FormatFloat(rounded, 'f', -1, 64)
	// Remove trailing zeros after decimal point
	if strings.Contains(result, ".") {
		result = strings.TrimRight(result, "0")
		result = strings.TrimRight(result, ".")
	}
	return result
}

// Percentage converts a decimal number to a percentage
func Percentage(value string) string {
	num := parseNumber(value)
	result := num * 100
	return formatNumberWithUnit(result, "%")
}

// EvaluateExpression evaluates a mathematical expression with units
// e.g., "10px * 2" -> "20px", "20px - 5px" -> "15px"
func EvaluateExpression(expr string) string {
	expr = strings.TrimSpace(expr)

	// Simple parser for left operand, operator, right operand
	// Handle operators: +, -, *, /

	// Find the operator (rightmost operator with lowest precedence first)
	// Look for + or - first (lower precedence)
	for i := len(expr) - 1; i > 0; i-- {
		if (expr[i] == '+' || expr[i] == '-') && !isPartOfNumber(expr, i) {
			left := strings.TrimSpace(expr[:i])
			op := string(expr[i])
			right := strings.TrimSpace(expr[i+1:])
			return evaluateBinaryOp(left, op, right)
		}
	}

	// Look for * or / (higher precedence)
	for i := len(expr) - 1; i > 0; i-- {
		if (expr[i] == '*' || expr[i] == '/') && !isPartOfNumber(expr, i) {
			left := strings.TrimSpace(expr[:i])
			op := string(expr[i])
			right := strings.TrimSpace(expr[i+1:])
			return evaluateBinaryOp(left, op, right)
		}
	}

	return ""
}

// isPartOfNumber checks if the character at position i is part of a number (not an operator)
func isPartOfNumber(expr string, i int) bool {
	if i == 0 {
		return false
	}
	// Check if previous char is a digit or decimal point
	prevChar := expr[i-1]
	return (prevChar >= '0' && prevChar <= '9') || prevChar == '.'
}

// evaluateBinaryOp evaluates a binary operation (left op right)
func evaluateBinaryOp(left, op, right string) string {
	leftNum := parseNumber(left)
	rightNum := parseNumber(right)

	var result float64
	var unit string

	switch op {
	case "+":
		result = leftNum + rightNum
		unit = extractUnit(left)
		if unit == "" {
			unit = extractUnit(right)
		}
	case "-":
		result = leftNum - rightNum
		unit = extractUnit(left)
		if unit == "" {
			unit = extractUnit(right)
		}
	case "*":
		result = leftNum * rightNum
		// When multiplying, prefer the unit from the value that has it
		unit = extractUnit(left)
		if unit == "" {
			unit = extractUnit(right)
		}
	case "/":
		if rightNum == 0 {
			return ""
		}
		result = leftNum / rightNum
		unit = extractUnit(left)
	default:
		return ""
	}

	return formatNumberWithUnit(result, unit)
}
