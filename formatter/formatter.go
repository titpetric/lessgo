package formatter

import (
	"bytes"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
	"github.com/sourcegraph/lessgo/renderer"
)

// Formatter formats LESS code with proper indentation and semicolons
type Formatter struct {
	indentSize int
	output     bytes.Buffer
	indent     int
}

// New creates a new formatter with the specified indentation size
func New(indentSize int) *Formatter {
	return &Formatter{
		indentSize: indentSize,
	}
}

// Format formats a stylesheet with proper indentation and semicolons
func (f *Formatter) Format(stylesheet *ast.Stylesheet) string {
	f.output.Reset()
	f.indent = 0

	for i, rule := range stylesheet.Rules {
		f.formatStatement(rule)
		if i < len(stylesheet.Rules)-1 {
			f.output.WriteString("\n")
		}
	}

	return f.output.String()
}

// formatStatement formats a statement
func (f *Formatter) formatStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.Rule:
		f.formatRule(s)
	case *ast.VariableDeclaration:
		f.formatVariableDeclaration(s)
	case *ast.AtRule:
		f.formatAtRule(s)
	case *ast.MixinCall:
		f.formatMixinCall(s)
	}
}

// formatRule formats a CSS rule
func (f *Formatter) formatRule(rule *ast.Rule) {
	// Write selector
	f.writeIndent()
	selector := strings.Join(rule.Selector.Parts, ", ")
	f.output.WriteString(selector)
	f.output.WriteString(" {\n")

	f.indent++

	// Write declarations
	for _, decl := range rule.Declarations {
		f.writeIndent()
		f.output.WriteString(decl.Property)
		f.output.WriteString(": ")
		f.output.WriteString(renderValue(decl.Value))
		f.output.WriteString(";\n")
	}

	// Write nested rules
	for i, nestedStmt := range rule.Rules {
		if _, isMixin := nestedStmt.(*ast.MixinCall); isMixin {
			// Format mixin calls as inline statements
			f.writeIndent()
			f.formatMixinCall(nestedStmt.(*ast.MixinCall))
			f.output.WriteString(";\n")
		} else {
			if i > 0 || len(rule.Declarations) > 0 {
				f.output.WriteString("\n")
			}
			f.formatStatement(nestedStmt)
		}
	}

	f.indent--

	f.writeIndent()
	f.output.WriteString("}\n")
}

// formatVariableDeclaration formats a variable declaration
func (f *Formatter) formatVariableDeclaration(decl *ast.VariableDeclaration) {
	f.writeIndent()
	f.output.WriteString("@")
	f.output.WriteString(decl.Name)
	f.output.WriteString(": ")
	f.output.WriteString(renderValue(decl.Value))
	f.output.WriteString(";\n")
}

// formatAtRule formats an at-rule
func (f *Formatter) formatAtRule(rule *ast.AtRule) {
	f.writeIndent()
	f.output.WriteString("@")
	f.output.WriteString(rule.Name)
	if rule.Parameters != "" {
		f.output.WriteString(" ")
		f.output.WriteString(rule.Parameters)
	}
	f.output.WriteString(" {\n")

	f.indent++

	if stmts, ok := rule.Block.([]ast.Statement); ok {
		for _, stmt := range stmts {
			f.formatStatement(stmt)
		}
	}

	f.indent--

	f.writeIndent()
	f.output.WriteString("}\n")
}

// formatMixinCall formats a mixin call
func (f *Formatter) formatMixinCall(call *ast.MixinCall) {
	// Build the mixin call string
	parts := []string{}
	for _, p := range call.Path {
		parts = append(parts, "."+p)
	}
	f.output.WriteString(strings.Join(parts, " > "))
	f.output.WriteString("()")
}

// renderValue renders a value (reuse renderer logic)
func renderValue(value ast.Value) string {
	r := renderer.NewRenderer()
	return r.RenderValuePublic(value)
}

// writeIndent writes the current indentation level
func (f *Formatter) writeIndent() {
	for i := 0; i < f.indent*f.indentSize; i++ {
		f.output.WriteString(" ")
	}
}
