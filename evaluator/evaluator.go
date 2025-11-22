package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/expr-lang/expr"
)

// Evaluator evaluates LESS expressions with variable context
type Evaluator struct {
	variables map[string]interface{}
}

// NewEvaluator creates a new expression evaluator with variable context
func NewEvaluator(vars map[string]string) *Evaluator {
	// Convert string variables to numeric or boolean values where appropriate
	evalVars := make(map[string]interface{})
	for k, v := range vars {
		// Try to parse as number (remove units for comparison)
		numVal := extractNumber(v)
		if numVal != nil {
			evalVars[k] = numVal
		} else if v == "true" {
			evalVars[k] = true
		} else if v == "false" {
			evalVars[k] = false
		} else {
			evalVars[k] = v
		}
	}
	return &Evaluator{variables: evalVars}
}

// extractNumber extracts numeric value from LESS value (e.g., "5px" -> 5)
func extractNumber(value string) interface{} {
	value = strings.TrimSpace(value)

	// Remove common CSS units
	units := []string{"px", "em", "rem", "%", "pt", "cm", "mm", "in", "pc", "ex", "ch", "vw", "vh", "vmin", "vmax"}
	for _, unit := range units {
		if strings.HasSuffix(value, unit) {
			numStr := strings.TrimSuffix(value, unit)
			if num, err := strconv.ParseFloat(numStr, 64); err == nil {
				return num
			}
		}
	}

	// Try direct float parse
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num
	}

	return nil
}

// Eval evaluates an expression string with the evaluator's variable context
// Preprocesses the expression to extract numbers from LESS values with units
func (e *Evaluator) Eval(expression string) (interface{}, error) {
	// Preprocess expression to handle LESS values with units
	processedExpr := preprocessExpression(expression)

	// fmt.Fprintf(os.Stderr, "[expr] Evaluating: %q (processed: %q)\n", expression, processedExpr)

	program, err := expr.Compile(processedExpr, expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	result, err := expr.Run(program, e.variables)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
}

// preprocessExpression handles LESS literals with units in expressions
// e.g., "14px > 12px" becomes "14 > 12"
func preprocessExpression(expr string) string {
	// Replace numeric values with units with just the numeric part
	// Note: % is a valid value which is later sanitized to float.
	units := []string{"px", "em", "rem", "pt", "cm", "mm", "in", "pc", "ex", "ch", "vw", "vh", "vmin", "vmax"}

	result := expr
	for _, unit := range units {
		// Find patterns like "123px" and replace with "123"
		// We need to be careful not to replace unit names in other contexts
		i := 0
		for i < len(result) {
			idx := strings.Index(result[i:], unit)
			if idx == -1 {
				break
			}
			idx += i
			// Check if there's a digit before the unit
			if idx > 0 && isDigit(result[idx-1]) {
				// Find the start of the number
				numStart := idx - 1
				for numStart > 0 && (isDigit(result[numStart-1]) || result[numStart-1] == '.') {
					numStart--
				}
				// Remove the unit
				result = result[:idx] + result[idx+len(unit):]
				i = idx
			} else {
				i = idx + len(unit)
			}
		}
	}

	parts := strings.Split(result, " ")
	for k, v := range parts {
		if strings.HasSuffix(v, "%") {
			if num, err := strconv.ParseFloat(v[:len(v)-1], 64); err == nil {
				num = num / 100.0
				parts[k] = fmt.Sprint(num)
			}
		}
	}

	spew.Dump(parts)

	return strings.Join(parts, " ")
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// EvalBool evaluates an expression and returns a boolean result
func (e *Evaluator) EvalBool(expression string) (bool, error) {
	result, err := e.Eval(expression)
	if err != nil {
		return false, err
	}

	switch v := result.(type) {
	case bool:
		return v, nil
	case float64:
		return v != 0, nil
	case int:
		return v != 0, nil
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		return v == "true", nil
	default:
		return false, nil
	}
}
