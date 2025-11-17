package functions

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
