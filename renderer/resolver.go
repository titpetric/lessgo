package renderer

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/evaluator"
	"github.com/titpetric/lessgo/expression"
)

var (
	// Cache compiled regex for variable interpolation
	varInterpolateRegex = regexp.MustCompile(`@\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
)

// Resolver resolves variables and expressions in declarations for rendering
type Resolver struct {
	file *dst.File
}

// NewResolver creates a new resolver from a file's variable stack
func NewResolver(file *dst.File) *Resolver {
	return &Resolver{
		file: file,
	}
}

// ResolveValue resolves a value string by substituting variables and evaluating expressions
func (r *Resolver) ResolveValue(stack *Stack, value string) (string, error) {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "#") {
		return value, nil
	}

	// Strip outer parentheses if present (they're just for grouping)
	if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
		// Check that the closing paren matches the opening one
		depth := 0
		allWrapped := true
		for i, ch := range value {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
				// If we hit zero before the last character, it's not fully wrapped
				if depth == 0 && i < len(value)-1 {
					allWrapped = false
					break
				}
			}
		}
		if allWrapped && depth == 0 {
			value = value[1 : len(value)-1]
			value = strings.TrimSpace(value)
		}
	}

	// First, substitute variables
	value = r.substituteVariables(stack, value)

	// Skip evaluation if it contains CSS-only functions (these should pass through)
	if isCSSOnlyFunction(value) {
		return value, nil
	}

	vars := stack.All()

	eval, err := expression.NewEvaluator(vars)
	if err != nil {
		return "", err
	}

	// Check if this is a function call (possibly with complex arguments like lists or nested calls)
	// If it is, evaluate it directly with the expression evaluator
	if expression.IsFunctionCall(value) {
		v, err := eval.Eval(value)
		if err == nil {
			return v.String(), nil
		}
		// If evaluation fails, fall through to tokenization
	}

	tokens, err := evaluator.Tokenize(value)
	if err != nil {
		v, err := eval.Eval(value)
		log.Printf("error tokenizing %s, eval %s", value, err)
		return fmt.Sprint(v), err
	}

	isExpression := evaluator.IsExpression(tokens)

	parts := []string{}
	for _, tok := range tokens {
		if expression.IsFunctionCall(tok.Text) {
			v, err := eval.Eval(tok.Text)
			if err == nil {
				tok.Text = fmt.Sprint(v)
			}
		}
		parts = append(parts, tok.Text)
	}

	if isExpression {
		result := strings.Join(parts, " ")
		v, err := eval.Eval(result)
		if err != nil {
			return value, err
		}
		return v.String(), nil
	}

	// For non-expressions, evaluate embedded functions and detect delimiter
	// Some values like "Arial, sans-serif" use commas, others like "1px solid" use spaces
	// Only check for top-level commas (not inside parentheses like rgb(...))
	delimiter := " "
	depth := 0
	for i := 0; i < len(value); i++ {
		if value[i] == '(' {
			depth++
		} else if value[i] == ')' {
			depth--
		} else if value[i] == ',' && depth == 0 {
			delimiter = ", "
			break
		}
	}

	// Join the parts first
	result := strings.Join(parts, delimiter)

	// Now evaluate any embedded functions in the result
	result = r.evaluateEmbeddedFunctions(eval, result)

	return result, nil
}

// InterpolateVariables replaces @{varname} patterns with their values from the stack
// This handles LESS variable interpolation syntax like .@{prefix} and @{prop}: value
func (r *Resolver) InterpolateVariables(stack *Stack, text string) string {
	// Use cached regex to avoid recompiling on every call
	return varInterpolateRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name from @{name}
		varName := match[2 : len(match)-1] // Remove @{ and }
		if val, ok := stack.Get(varName); ok {
			return val
		}
		return match // If variable not found, return original
	})
}

// isCSSOnlyFunction checks if the value contains CSS-only functions that we shouldn't evaluate
// Note: rgb, rgba, hsl, hsla are now handled by the evaluator, so we don't skip them
func isCSSOnlyFunction(value string) bool {
	cssOnlyFuncs := []string{"hwb(", "url("}
	trimmedValue := strings.TrimSpace(value)
	for _, fn := range cssOnlyFuncs {
		if strings.HasPrefix(trimmedValue, fn) {
			return true
		}
	}
	return false
}

// substituteVariables replaces @variable with their values
func (r *Resolver) substituteVariables(stack *Stack, value string) string {
	// Simple variable substitution
	for {
		idx := strings.Index(value, "@")
		if idx == -1 {
			break
		}

		// Find the end of the variable name
		i := idx + 1
		for i < len(value) && (isVarChar(rune(value[i]))) {
			i++
		}

		if i == idx+1 {
			// No valid variable name found
			break
		}

		varName := value[idx+1 : i]
		if val, ok := stack.Get(varName); ok {
			// Recursively resolve in case variable references another variable
			resolved, _ := r.ResolveValue(stack, val)
			value = value[:idx] + resolved + value[i:]
		} else {
			break
		}
	}

	return value
}

// containsOperator checks if a value contains arithmetic operators
func (r *Resolver) containsOperator(value string) bool {
	// Skip comma-separated lists (these shouldn't be evaluated as expressions)
	// But only if commas are at top level, not inside parentheses (like rgb() args)
	parenDepth := 0
	hasTopLevelComma := false
	for i := 0; i < len(value); i++ {
		if value[i] == '(' {
			parenDepth++
		} else if value[i] == ')' {
			parenDepth--
		} else if value[i] == ',' && parenDepth == 0 {
			hasTopLevelComma = true
			break
		}
	}
	if hasTopLevelComma {
		return false
	}

	parenDepth = 0
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch == '(' {
			parenDepth++
		} else if ch == ')' {
			parenDepth--
		} else if parenDepth == 0 {
			if ch == '+' || ch == '*' || ch == '/' {
				return true
			} else if ch == '-' {
				// Only treat - as operator if surrounded by values/numbers
				if i > 0 && i < len(value)-1 {
					// Skip back over spaces to find previous token
					j := i - 1
					for j >= 0 && value[j] == ' ' {
						j--
					}
					// Skip forward over spaces to find next token
					k := i + 1
					for k < len(value) && value[k] == ' ' {
						k++
					}

					if j >= 0 && k < len(value) {
						prev := value[j]
						next := value[k]
						// It's an operator if surrounded by value chars or )
						if (isValueChar(rune(prev)) || prev == ')') && (isValueChar(rune(next)) || next == '(') {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// evaluateEmbeddedFunctions evaluates all function calls embedded in a string
func (r *Resolver) evaluateEmbeddedFunctions(eval *expression.Evaluator, value string) string {
	functions := r.extractFunctionsFromValue(value)
	result := value

	for _, funcCall := range functions {
		v, err := eval.Eval(funcCall)
		if err == nil {
			// Replace all occurrences of this function call with its result
			result = strings.ReplaceAll(result, funcCall, v.String())
		}
	}

	return result
}

var (
	// Cached list of function names to evaluate in property values
	// Populated once at init time
	evaluableEmbeddedFuncNames []string
)

func initEvaluableFuncNames() {
	if evaluableEmbeddedFuncNames != nil {
		return
	}

	// Exclude type-checking (is*) and special syntax (format, replace) functions
	// Note: e, escape, boolean, if are included because they may appear in embedded contexts
	excludeFuncs := map[string]bool{
		// Type checking functions that return booleans
		"isnumber": true, "isstring": true, "iscolor": true, "iskeyword": true,
		"isurl": true, "ispixel": true, "isem": true, "ispercentage": true,
		"isunit": true, "isruleset": true, "islist": true, "isdefined": true,
		"isnumberfunction": true, "isstringfunction": true, "iscolorfunction": true,
		"iskeywordfunction": true, "isurlfunc": true, "ispixelfunction": true,
		"isemfunction": true, "ispercentagefunction": true, "isunitfunction": true,
		"isrulesetfunction": true, "islistfunction": true,
		// Arithmetic functions that are operators, not embedded functions
		"add": true, "subtract": true, "divide": true,
	}

	allFuncNames := expression.GetRegisteredFunctionNames()
	evaluableEmbeddedFuncNames = make([]string, 0, len(allFuncNames))
	for _, name := range allFuncNames {
		if !excludeFuncs[name] {
			evaluableEmbeddedFuncNames = append(evaluableEmbeddedFuncNames, name)
		}
	}
}

// extractFunctionsFromValue extracts all function calls from a string value
// Uses registered LESS evaluation functions from the expression package
func (r *Resolver) extractFunctionsFromValue(value string) []string {
	initEvaluableFuncNames()

	var functions []string
	funcNames := evaluableEmbeddedFuncNames

	for _, fnName := range funcNames {
		searchStr := fnName + "("
		pos := 0

		for {
			idx := strings.Index(value[pos:], searchStr)
			if idx == -1 {
				break
			}
			idx += pos

			// Skip if this function call is inside quotes
			// Count quotes before this position to determine if we're inside a quoted string
			quoteCount := 0
			for i := 0; i < idx; i++ {
				if (value[i] == '"' || value[i] == '\'') && (i == 0 || value[i-1] != '\\') {
					quoteCount++
				}
			}
			if quoteCount%2 != 0 {
				// We're inside a quoted string, skip this match
				pos = idx + len(searchStr)
				continue
			}

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
