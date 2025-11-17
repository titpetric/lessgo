package evaluator

import (
	"fmt"
	"strconv"
	"strings"

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
func (e *Evaluator) Eval(expression string) (interface{}, error) {
	program, err := expr.Compile(expression, expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	result, err := expr.Run(program, e.variables)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
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
