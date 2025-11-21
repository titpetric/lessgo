package renderer

import (
	"bytes"
	"strings"

	"github.com/titpetric/lessgo/parser"
)

// Renderer renders DST nodes to CSS
type Renderer struct {
	output bytes.Buffer
	vars   map[string]string // Simple string-based variables
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		vars: make(map[string]string),
	}
}

// RenderNode renders a DST node to CSS
func (r *Renderer) RenderNode(root *parser.Node) string {
	if root == nil {
		return ""
	}

	r.output.Reset()
	r.vars = make(map[string]string)

	// First pass: collect variables
	r.collectVariables(root)

	// Second pass: render
	r.renderNode(root, "")

	return r.output.String()
}

// collectVariables collects all variable definitions
func (r *Renderer) collectVariables(node *parser.Node) {
	if node == nil {
		return
	}

	for _, child := range node.Children {
		if child.Type == "variable" {
			r.vars[child.Name] = child.Value
		}
		r.collectVariables(child)
	}
}

// renderNode renders a node
func (r *Renderer) renderNode(node *parser.Node, parentSelector string) {
	if node == nil {
		return
	}

	switch node.Type {
	case "stylesheet":
		for _, child := range node.Children {
			r.renderNode(child, "")
		}

	case "rule":
		selector := r.buildSelector(node.Name, parentSelector)
		if len(node.Children) == 0 {
			return
		}

		r.output.WriteString(selector)
		r.output.WriteString(" {\n")

		for _, child := range node.Children {
			if child.Type == "declaration" {
				value := r.evaluateValue(child.Value)
				r.output.WriteString("  ")
				r.output.WriteString(child.Name)
				r.output.WriteString(": ")
				r.output.WriteString(value)
				r.output.WriteString(";\n")
			} else if child.Type == "rule" {
				r.renderNode(child, selector)
			} else if child.Type == "variable" {
				r.vars[child.Name] = child.Value
			}
		}

		r.output.WriteString("}\n")

	case "variable":
		r.vars[node.Name] = node.Value

	case "atrule":
		r.output.WriteString("@")
		r.output.WriteString(node.Name)
		if node.Value != "" {
			r.output.WriteString(" ")
			r.output.WriteString(node.Value)
		}
		r.output.WriteString(" {\n")

		for _, child := range node.Children {
			if child.Type == "rule" {
				r.output.WriteString("  ")
				r.output.WriteString(child.Name)
				r.output.WriteString(" {\n")
				for _, decl := range child.Children {
					if decl.Type == "declaration" {
						value := r.evaluateValue(decl.Value)
						r.output.WriteString("    ")
						r.output.WriteString(decl.Name)
						r.output.WriteString(": ")
						r.output.WriteString(value)
						r.output.WriteString(";\n")
					}
				}
				r.output.WriteString("  }\n")
			}
		}

		r.output.WriteString("}\n")
	}
}

// buildSelector builds selector with nesting support
func (r *Renderer) buildSelector(selector, parentSelector string) string {
	if parentSelector == "" {
		return selector
	}

	// Handle & parent reference
	if strings.Contains(selector, "&") {
		return strings.ReplaceAll(selector, "&", parentSelector)
	}

	// Default: child selector
	return parentSelector + " " + selector
}

// evaluateValue evaluates a value (variable substitution, functions, etc)
func (r *Renderer) evaluateValue(value string) string {
	// Replace variables
	for varName, varValue := range r.vars {
		value = strings.ReplaceAll(value, "@"+varName, varValue)
	}

	// Handle functions like ceil(), floor(), etc
	value = r.evaluateFunctions(value)

	return value
}

// evaluateFunctions evaluates built-in functions
func (r *Renderer) evaluateFunctions(value string) string {
	// Simple function evaluation
	// Pattern: functionName(arg1, arg2, ...)

	// ceil(value) -> round up
	if strings.Contains(value, "ceil(") {
		value = r.evalFunc(value, "ceil", Ceil)
	}
	// floor(value) -> round down
	if strings.Contains(value, "floor(") {
		value = r.evalFunc(value, "floor", Floor)
	}
	// round(value) -> round
	if strings.Contains(value, "round(") {
		value = r.evalFunc(value, "round", Round)
	}
	// abs(value) -> absolute value
	if strings.Contains(value, "abs(") {
		value = r.evalFunc(value, "abs", Abs)
	}

	return value
}

// evalFunc evaluates a function call
func (r *Renderer) evalFunc(expr, funcName string, fn func(string) string) string {
	pattern := funcName + "("
	for {
		idx := strings.Index(expr, pattern)
		if idx == -1 {
			break
		}

		// Find matching closing paren
		start := idx + len(pattern)
		parenCount := 1
		end := start
		for end < len(expr) && parenCount > 0 {
			if expr[end] == '(' {
				parenCount++
			} else if expr[end] == ')' {
				parenCount--
			}
			end++
		}

		if parenCount != 0 {
			break // Unmatched parens
		}

		// Extract argument
		arg := expr[start : end-1]
		// Recursively evaluate in case argument is another function
		arg = r.evaluateValue(arg)

		// Call function
		result := fn(arg)

		// Replace in expression
		expr = expr[:idx] + result + expr[end:]
	}

	return expr
}
