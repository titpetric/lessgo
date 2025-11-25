package expression

import (
	"fmt"
	"regexp"

	"github.com/titpetric/lessgo/internal/strings"
)

var (
	// Cache compiled regex for variable substitution
	varSubstituteRegex = regexp.MustCompile(`@([a-zA-Z_][a-zA-Z0-9_-]*)`)
)

// Evaluator evaluates LESS expressions
type Evaluator struct {
	variables map[string]*Value
}

// NewEvaluator creates a new evaluator
func NewEvaluator(vars map[string]string) (*Evaluator, error) {
	result := &Evaluator{
		variables: make(map[string]*Value),
	}
	for k, v := range vars {
		r, err := Parse(v)
		if err != nil {
			return nil, err
		}
		result.variables[k] = r
	}
	return result, nil
}

// SetVariable sets a variable value
func (e *Evaluator) SetVariable(name string, v *Value) {
	e.variables[name] = v
}

// Eval evaluates an expression string
// Examples: "10px * 2", "@base + 5px", "50% + 10px"
func (e *Evaluator) Eval(expr string) (*Value, error) {
	expr = strings.TrimSpace(expr)

	// Don't substitute variables here - let evalExpression handle them
	// This allows functions to receive variable references directly

	// Evaluate the expression
	return e.evalExpression(expr)
}

// substituteVariables replaces @variable with their values
func (e *Evaluator) substituteVariables(expr string) string {
	return varSubstituteRegex.ReplaceAllStringFunc(expr, func(match string) string {
		varName := match[1:] // remove the @
		if v, ok := e.variables[varName]; ok {
			// Prefer Raw field for lists and other non-numeric values
			if v.Raw != "" {
				return v.Raw
			}
			return v.String()
		}
		return match
	})
}

// evalExpression evaluates a mathematical expression
func (e *Evaluator) evalExpression(expr string) (*Value, error) {
	expr = strings.TrimSpace(expr)

	// Try to parse as simple value
	if !containsOperator(expr) {
		// Check for variable reference
		if strings.HasPrefix(expr, "@") && !strings.ContainsAny(expr, " ()+-*/") {
			varName := strings.TrimPrefix(expr, "@")
			if v, ok := e.variables[varName]; ok {
				return v, nil
			}
		}

		// Check if it's a function call
		if IsFunctionCall(expr) {
			return e.evalFunctionCall(expr)
		}

		// Parse as a value
		v, err := Parse(expr)
		if err != nil {
			return nil, err
		}

		// If the parsed value is a raw string that contains functions,
		// try to evaluate those functions
		if v.Raw != "" && v.Color == nil && (strings.Contains(v.Raw, "(") && strings.Contains(v.Raw, ")")) {
			result := e.evaluateEmbeddedFunctions(v.Raw)
			if result != v.Raw {
				// Successfully evaluated functions, return new value
				return Parse(result)
			}
		}

		return v, nil
	}

	// Parse and evaluate operators left-to-right (same precedence)
	return e.parseAddSub(expr)
}

// evalFunctionCall evaluates a function call expression
func (e *Evaluator) evalFunctionCall(expr string) (*Value, error) {
	funcName, args, err := ParseFunctionCall(expr)
	if err != nil {
		return nil, err
	}

	// isdefined() is not evaluated by lessc - pass it through with substituted variables
	if strings.ToLower(funcName) == "isdefined" {
		// Reconstruct the function call with substituted variables
		var argStrs []string
		for _, v := range args {
			argStr := v.String()
			argStr = e.substituteVariables(argStr)
			argStrs = append(argStrs, argStr)
		}
		result := funcName + "(" + strings.Join(argStrs, ", ") + ")"
		return Parse(result) // Return as raw string
	}

	funcArgs := make([]any, 0, len(args))
	for _, v := range args {
		argStr := v.String()
		// Substitute variables in function arguments
		argStr = e.substituteVariables(argStr)
		funcArgs = append(funcArgs, argStr)
	}

	if !IsRegisteredFunction(funcName) {
		return nil, fmt.Errorf("Unknown function: %s", funcName)
	}

	res, err := Call(funcName, funcArgs...)
	if err != nil {
		return nil, err
	}

	return Parse(fmt.Sprint(res))
}

// parseAddSub handles + and - operators
func (e *Evaluator) parseAddSub(expr string) (*Value, error) {
	// Substitute variables in arithmetic expressions
	expr = e.substituteVariables(expr)

	// Find operators not inside parentheses
	parts := splitByOperator(expr, []string{"+", "-"})

	if len(parts) == 1 {
		return e.parseMulDiv(parts[0].value)
	}

	left, err := e.parseMulDiv(parts[0].value)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(parts); i++ {
		right, err := e.parseMulDiv(parts[i].value)
		if err != nil {
			return nil, err
		}

		switch parts[i].op {
		case "+":
			left, err = left.Add(right)
		case "-":
			left, err = left.Subtract(right)
		}
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

// parseMulDiv handles * and / operators
func (e *Evaluator) parseMulDiv(expr string) (*Value, error) {
	expr = strings.TrimSpace(expr)

	// Find operators not inside parentheses
	parts := splitByOperator(expr, []string{"*", "/"})

	if len(parts) == 0 {
		return Parse(expr)
	}

	if len(parts) == 1 {
		valStr := strings.TrimSpace(parts[0].value)
		// Check if it's a function call
		if IsFunctionCall(valStr) {
			return e.evalFunctionCall(valStr)
		}
		return Parse(valStr)
	}

	left, err := e.parseValue(parts[0].value)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(parts); i++ {
		right, err := e.parseValue(parts[i].value)
		if err != nil {
			return nil, err
		}

		switch parts[i].op {
		case "*":
			left, err = left.Multiply(right)
		case "/":
			left, err = left.Divide(right)
		}
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

// parseValue parses either a function call, a variable reference, or a simple value
func (e *Evaluator) parseValue(expr string) (*Value, error) {
	expr = strings.TrimSpace(expr)

	// Check for variable reference
	if strings.HasPrefix(expr, "@") && !strings.ContainsAny(expr, " ()+-*/") {
		varName := strings.TrimPrefix(expr, "@")
		if v, ok := e.variables[varName]; ok {
			return v, nil
		}
	}

	if IsFunctionCall(expr) {
		return e.evalFunctionCall(expr)
	}
	return Parse(expr)
}

// opPart represents an operand with its preceding operator
type opPart struct {
	op    string
	value string
}

// splitByOperator splits an expression by operators, respecting parentheses
func splitByOperator(expr string, ops []string) []opPart {
	var parts []opPart
	var current strings.Builder
	parenDepth := 0
	var lastOp string

	for i := 0; i < len(expr); i++ {
		ch := expr[i]

		if ch == '(' {
			parenDepth++
			current.WriteByte(ch)
		} else if ch == ')' {
			parenDepth--
			current.WriteByte(ch)
		} else if parenDepth == 0 && isOperator(string(ch), ops) {
			// Found an operator at depth 0
			parts = append(parts, opPart{
				op:    lastOp,
				value: strings.TrimSpace(current.String()),
			})
			current.Reset()
			lastOp = string(ch)
		} else {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, opPart{
			op:    lastOp,
			value: strings.TrimSpace(current.String()),
		})
	}

	return parts
}

// isOperator checks if a string is in the operator list
func isOperator(s string, ops []string) bool {
	for _, op := range ops {
		if s == op {
			return true
		}
	}
	return false
}

// evaluateEmbeddedFunctions evaluates all function calls embedded in a string
func (e *Evaluator) evaluateEmbeddedFunctions(value string) string {
	functions := e.extractFunctions(value)
	result := value

	for _, funcCall := range functions {
		v, err := e.evalFunctionCall(funcCall)
		if err == nil {
			// Replace all occurrences of this function call with its result
			result = strings.ReplaceAll(result, funcCall, v.String())
		}
	}

	return result
}

// extractFunctions extracts all function calls from a string
func (e *Evaluator) extractFunctions(value string) []string {
	var functions []string
	lessFunctions := make(map[string]bool)
	for k := range funcMap {
		lessFunctions[k] = true
	}

	for fnName := range lessFunctions {
		searchStr := fnName + "("
		pos := 0

		for {
			idx := strings.Index(value[pos:], searchStr)
			if idx == -1 {
				break
			}
			idx += pos

			// Find the closing paren
			depth := 0
			endIdx := -1
			for i := idx + len(searchStr); i < len(value); i++ {
				if value[i] == '(' {
					depth++
				} else if value[i] == ')' {
					if depth == 0 {
						endIdx = i
						break
					}
					depth--
				}
			}

			if endIdx == -1 {
				break
			}

			// Extract the function
			funcCall := value[idx : endIdx+1]
			functions = append(functions, funcCall)

			// Move past this function
			pos = endIdx + 1
		}
	}

	return functions
}

// containsOperator checks if an expression contains any operator
func containsOperator(expr string) bool {
	parenDepth := 0
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if ch == '(' {
			parenDepth++
		} else if ch == ')' {
			parenDepth--
		} else if parenDepth == 0 && (ch == '+' || ch == '-' || ch == '*' || ch == '/') {
			// For hyphens, check if it's part of a function name (e.g., get-unit)
			// A hyphen is an operator only if it's not surrounded by alphanumeric characters
			if ch == '-' {
				// Check if previous and next characters are alphanumeric (making it part of a name)
				if i > 0 && i < len(expr)-1 {
					prevChar := expr[i-1]
					nextChar := expr[i+1]
					if isIdentifierChar(rune(prevChar)) && isIdentifierChar(rune(nextChar)) {
						// This hyphen is part of a function name, not an operator
						continue
					}
				}
			}
			return true
		}
	}
	return false
}

func isIdentifierChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}
