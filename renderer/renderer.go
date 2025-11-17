package renderer

import (
	"bytes"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
)

// Renderer converts an AST to CSS
type Renderer struct {
	output bytes.Buffer
	indent int
	vars   map[string]ast.Value
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		vars: make(map[string]ast.Value),
	}
}

// Render renders the stylesheet to CSS
func (r *Renderer) Render(stylesheet *ast.Stylesheet) string {
	for _, rule := range stylesheet.Rules {
		r.renderStatement(rule, "")
	}
	return r.output.String()
}

// renderStatement renders a statement
func (r *Renderer) renderStatement(stmt ast.Statement, parentSelector string) {
	switch s := stmt.(type) {
	case *ast.Rule:
		r.renderRule(s, parentSelector)
	case *ast.VariableDeclaration:
		r.renderVariableDeclaration(s)
	case *ast.AtRule:
		r.renderAtRule(s)
	}
}

// renderRule renders a CSS rule
func (r *Renderer) renderRule(rule *ast.Rule, parentSelector string) {
	selector := r.buildSelector(rule.Selector, parentSelector)

	// Render declarations
	if len(rule.Declarations) > 0 {
		r.output.WriteString(selector)
		r.output.WriteString(" {\n")

		for _, decl := range rule.Declarations {
			r.output.WriteString("  ")
			r.output.WriteString(decl.Property)
			r.output.WriteString(": ")
			r.output.WriteString(r.renderValue(decl.Value))
			r.output.WriteString(";\n")
		}

		r.output.WriteString("}\n")
	}

	// Render nested rules
	for _, nestedStmt := range rule.Rules {
		r.renderStatement(nestedStmt, selector)
	}
}

// buildSelector builds the full selector from nesting
func (r *Renderer) buildSelector(selector ast.Selector, parentSelector string) string {
	if len(selector.Parts) == 0 {
		return parentSelector
	}

	parts := []string{}
	for _, part := range selector.Parts {
		if strings.Contains(part, "&") {
			// Replace & with parent selector
			result := strings.ReplaceAll(part, "&", parentSelector)
			parts = append(parts, result)
		} else if parentSelector != "" {
			// Append to parent selector
			parts = append(parts, parentSelector+" "+part)
		} else {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, ", ")
}

// renderValue renders a value to CSS
func (r *Renderer) renderValue(value ast.Value) string {
	switch v := value.(type) {
	case *ast.Literal:
		return v.Value
	case *ast.Variable:
		// Look up variable
		if val, ok := r.vars[v.Name]; ok {
			return r.renderValue(val)
		}
		return "@" + v.Name // Fallback
	case *ast.FunctionCall:
		return r.renderFunctionCall(v)
	case *ast.List:
		parts := []string{}
		for _, val := range v.Values {
			parts = append(parts, r.renderValue(val))
		}
		sep := v.Separator
		if sep == "" {
			sep = " "
		}
		return strings.Join(parts, sep)
	case *ast.BinaryOp:
		return r.renderBinaryOp(v)
	default:
		return ""
	}
}

// renderFunctionCall renders a function call
func (r *Renderer) renderFunctionCall(fn *ast.FunctionCall) string {
	args := []string{}
	for _, arg := range fn.Arguments {
		args = append(args, r.renderValue(arg))
	}
	return fn.Name + "(" + strings.Join(args, ", ") + ")"
}

// renderBinaryOp renders a binary operation
func (r *Renderer) renderBinaryOp(op *ast.BinaryOp) string {
	return r.renderValue(op.Left) + " " + op.Operator + " " + r.renderValue(op.Right)
}

// renderVariableDeclaration renders a variable declaration (stores it)
func (r *Renderer) renderVariableDeclaration(decl *ast.VariableDeclaration) {
	r.vars[decl.Name] = decl.Value
}

// renderAtRule renders an at-rule
func (r *Renderer) renderAtRule(rule *ast.AtRule) {
	r.output.WriteString("@")
	r.output.WriteString(rule.Name)
	if rule.Parameters != "" {
		r.output.WriteString(" ")
		r.output.WriteString(rule.Parameters)
	}
	r.output.WriteString(" {\n")

	if stmts, ok := rule.Block.([]ast.Statement); ok {
		for _, stmt := range stmts {
			r.renderStatement(stmt, "")
		}
	}

	r.output.WriteString("}\n")
}
