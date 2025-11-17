package renderer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
	"github.com/sourcegraph/lessgo/functions"
)

// Renderer converts an AST to CSS
type Renderer struct {
	output bytes.Buffer
	indent int
	vars   map[string]ast.Value
	mixins map[string]*ast.Rule // Store mixin definitions by name
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		vars:   make(map[string]ast.Value),
		mixins: make(map[string]*ast.Rule),
	}
}

// Render renders the stylesheet to CSS
func (r *Renderer) Render(stylesheet *ast.Stylesheet) string {
	// First pass: collect mixin definitions
	r.collectMixins(stylesheet)

	// Second pass: render all statements (mixins are just normal rules)
	for _, rule := range stylesheet.Rules {
		r.renderStatement(rule, "")
	}
	return r.output.String()
}

// collectMixins finds all mixin definitions in the stylesheet
func (r *Renderer) collectMixins(stylesheet *ast.Stylesheet) {
	for _, stmt := range stylesheet.Rules {
		if rule, ok := stmt.(*ast.Rule); ok {
			// Check if this is a mixin definition (class or id selector with no body that could be used as mixin)
			if len(rule.Selector.Parts) == 1 {
				selector := rule.Selector.Parts[0]
				// Extract mixin name from .classname or #id format
				if (strings.HasPrefix(selector, ".") || strings.HasPrefix(selector, "#")) && !strings.Contains(selector, " ") {
					name := selector[1:] // Remove . or #
					r.mixins[name] = rule
				}
			}
		}
	}
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
	case *ast.MixinCall:
		r.renderMixinCall(s)
	}
}

// renderRule renders a CSS rule
func (r *Renderer) renderRule(rule *ast.Rule, parentSelector string) {
	selector := r.buildSelector(rule.Selector, parentSelector)

	// Build list of all declarations, with mixin declarations inserted at mixin call positions
	// Note: rule.Declarations and rule.Rules are separate lists, but the parser preserves
	// the order in which declarations and nested rules appear in the source.
	// However, to get the correct order, we need to use both lists together.

	// For simplicity, we'll collect all declarations and mixin-provided declarations
	// The mixin calls should appear in the same relative position
	allDeclarations := []ast.Declaration{}

	// Add rule's own declarations
	allDeclarations = append(allDeclarations, rule.Declarations...)

	// Also add declarations from mixin calls (they'll appear first in the output)
	for _, nestedStmt := range rule.Rules {
		if mixinCall, ok := nestedStmt.(*ast.MixinCall); ok {
			if len(mixinCall.Path) > 0 {
				mixinName := mixinCall.Path[len(mixinCall.Path)-1]
				if mixin, found := r.mixins[mixinName]; found {
					// Insert mixin declarations at the beginning
					allDeclarations = append(mixin.Declarations, allDeclarations...)
				}
			}
		}
	}

	// Render declarations
	if len(allDeclarations) > 0 {
		r.output.WriteString(selector)
		r.output.WriteString(" {\n")

		for _, decl := range allDeclarations {
			r.output.WriteString("  ")
			r.output.WriteString(decl.Property)
			r.output.WriteString(": ")
			r.output.WriteString(r.renderValue(decl.Value))
			r.output.WriteString(";\n")
		}

		r.output.WriteString("}\n")
	}

	// Render nested rules (excluding mixin calls)
	for _, nestedStmt := range rule.Rules {
		// Skip mixin calls - they're already handled above
		if _, ok := nestedStmt.(*ast.MixinCall); !ok {
			r.renderStatement(nestedStmt, selector)
		}
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

// RenderValuePublic renders a value to CSS (public for external use)
func (r *Renderer) RenderValuePublic(value ast.Value) string {
	return r.renderValue(value)
}

// FormatValue renders a value for formatting (preserves variables without evaluation)
func (r *Renderer) FormatValue(value ast.Value) string {
	return r.formatValue(value)
}

// formatValue renders a value for formatting without evaluating variables
func (r *Renderer) formatValue(value ast.Value) string {
	switch v := value.(type) {
	case *ast.Literal:
		return v.Value
	case *ast.Variable:
		// Keep variable as-is without evaluating
		return "@" + v.Name
	case *ast.FunctionCall:
		// Format function call without evaluation
		args := []string{}
		for _, arg := range v.Arguments {
			args = append(args, r.formatValue(arg))
		}
		return v.Name + "(" + strings.Join(args, ", ") + ")"
	case *ast.List:
		parts := []string{}
		for _, val := range v.Values {
			parts = append(parts, r.formatValue(val))
		}
		sep := v.Separator
		if sep == "" {
			sep = " "
		}
		return strings.Join(parts, sep)
	case *ast.BinaryOp:
		// Format binary op without evaluation
		return r.formatValue(v.Left) + " " + v.Operator + " " + r.formatValue(v.Right)
	default:
		return ""
	}
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

	// Try to evaluate color functions
	if result := r.evaluateColorFunction(fn.Name, args); result != "" {
		return result
	}

	return fn.Name + "(" + strings.Join(args, ", ") + ")"
}

// evaluateColorFunction evaluates color and math functions
func (r *Renderer) evaluateColorFunction(name string, args []string) string {
	switch name {
	case "rgb":
		if len(args) != 3 {
			return ""
		}
		return r.evalRGB(args)
	case "rgba":
		if len(args) != 4 {
			return ""
		}
		return r.evalRGBA(args)
	case "lighten":
		if len(args) != 2 {
			return ""
		}
		return r.evalLighten(args[0], args[1])
	case "darken":
		if len(args) != 2 {
			return ""
		}
		return r.evalDarken(args[0], args[1])
	case "saturate":
		if len(args) != 2 {
			return ""
		}
		return r.evalSaturate(args[0], args[1])
	case "desaturate":
		if len(args) != 2 {
			return ""
		}
		return r.evalDesaturate(args[0], args[1])
	case "spin":
		if len(args) != 2 {
			return ""
		}
		return r.evalSpin(args[0], args[1])
	case "greyscale":
		if len(args) != 1 {
			return ""
		}
		return r.evalGreyscale(args[0])
	case "ceil":
		if len(args) != 1 {
			return ""
		}
		return functions.Ceil(args[0])
	case "floor":
		if len(args) != 1 {
			return ""
		}
		return functions.Floor(args[0])
	case "round":
		if len(args) != 1 {
			return ""
		}
		return functions.Round(args[0])
	case "abs":
		if len(args) != 1 {
			return ""
		}
		return functions.Abs(args[0])
	case "sqrt":
		if len(args) != 1 {
			return ""
		}
		return functions.Sqrt(args[0])
	case "pow":
		if len(args) != 2 {
			return ""
		}
		return functions.Pow(args[0], args[1])
	case "min":
		if len(args) < 1 {
			return ""
		}
		return functions.Min(args...)
	case "max":
		if len(args) < 1 {
			return ""
		}
		return functions.Max(args...)
	}
	return ""
}

func (r *Renderer) evalRGB(args []string) string {
	if len(args) != 3 {
		return ""
	}

	r1, _ := strconv.ParseFloat(strings.TrimSpace(args[0]), 64)
	g1, _ := strconv.ParseFloat(strings.TrimSpace(args[1]), 64)
	b1, _ := strconv.ParseFloat(strings.TrimSpace(args[2]), 64)

	c := &functions.Color{R: r1, G: g1, B: b1, A: 1.0}
	return c.ToHex()
}

func (r *Renderer) evalRGBA(args []string) string {
	if len(args) != 4 {
		return ""
	}

	r1, _ := strconv.ParseFloat(strings.TrimSpace(args[0]), 64)
	g1, _ := strconv.ParseFloat(strings.TrimSpace(args[1]), 64)
	b1, _ := strconv.ParseFloat(strings.TrimSpace(args[2]), 64)
	a1, _ := strconv.ParseFloat(strings.TrimSpace(args[3]), 64)

	c := &functions.Color{R: r1, G: g1, B: b1, A: a1}
	return c.ToRGB()
}

func (r *Renderer) evalLighten(colorStr, amountStr string) string {
	c, err := functions.ParseColor(colorStr)
	if err != nil {
		return ""
	}

	amount := parsePercentage(amountStr)
	result := c.Lighten(amount)
	return result.ToHex()
}

func (r *Renderer) evalDarken(colorStr, amountStr string) string {
	c, err := functions.ParseColor(colorStr)
	if err != nil {
		return ""
	}

	amount := parsePercentage(amountStr)
	result := c.Darken(amount)
	return result.ToHex()
}

func (r *Renderer) evalSaturate(colorStr, amountStr string) string {
	c, err := functions.ParseColor(colorStr)
	if err != nil {
		return ""
	}

	amount := parsePercentage(amountStr)
	result := c.Saturate(amount)
	return result.ToHex()
}

func (r *Renderer) evalDesaturate(colorStr, amountStr string) string {
	c, err := functions.ParseColor(colorStr)
	if err != nil {
		return ""
	}

	amount := parsePercentage(amountStr)
	result := c.Desaturate(amount)
	return result.ToHex()
}

func (r *Renderer) evalSpin(colorStr, degreesStr string) string {
	c, err := functions.ParseColor(colorStr)
	if err != nil {
		return ""
	}

	degrees, _ := strconv.ParseFloat(strings.TrimSuffix(degreesStr, "deg"), 64)
	result := c.Spin(degrees)
	return result.ToHex()
}

func (r *Renderer) evalGreyscale(colorStr string) string {
	c, err := functions.ParseColor(colorStr)
	if err != nil {
		return ""
	}

	result := c.Greyscale()
	return result.ToHex()
}

// renderBinaryOp evaluates and renders a binary operation
func (r *Renderer) renderBinaryOp(op *ast.BinaryOp) string {
	// Try to evaluate the operation
	result := r.evaluateBinaryOp(op)
	if result != "" {
		return result
	}
	// Fallback to rendering as-is if we can't evaluate
	return r.renderValue(op.Left) + " " + op.Operator + " " + r.renderValue(op.Right)
}

// evaluateBinaryOp evaluates a binary operation and returns the result, or empty string if not evaluable
func (r *Renderer) evaluateBinaryOp(op *ast.BinaryOp) string {
	leftStr := r.renderValue(op.Left)
	rightStr := r.renderValue(op.Right)

	// Try to parse as numbers with units
	leftNum, leftUnit := parseNumberWithUnit(leftStr)
	rightNum, rightUnit := parseNumberWithUnit(rightStr)

	if leftNum == nil || rightNum == nil {
		return "" // Can't evaluate
	}

	var result float64
	switch op.Operator {
	case "+":
		// For addition, ensure units are compatible or use left unit
		if leftUnit != rightUnit && rightUnit != "" && leftUnit != "" {
			return "" // Can't add incompatible units
		}
		result = *leftNum + *rightNum
	case "-":
		// For subtraction, ensure units are compatible
		if leftUnit != rightUnit && rightUnit != "" && leftUnit != "" {
			return "" // Can't subtract incompatible units
		}
		result = *leftNum - *rightNum
	case "*":
		result = *leftNum * *rightNum
		// Multiplication: units multiply
		if leftUnit != "" && rightUnit != "" {
			return "" // Can't multiply two numbers with units in standard CSS
		}
	case "/":
		if *rightNum == 0 {
			return "" // Division by zero
		}
		result = *leftNum / *rightNum
		// Division: right unit must be empty or match left
		if rightUnit != "" && rightUnit != leftUnit {
			return "" // Can't divide by a unit
		}
	default:
		return ""
	}

	// Format the result
	unit := leftUnit
	if unit == "" {
		unit = rightUnit
	}

	// Format as integer if whole number, otherwise with decimals
	var resultStr string
	if result == float64(int64(result)) {
		resultStr = fmt.Sprintf("%d", int64(result))
	} else {
		resultStr = fmt.Sprintf("%g", result)
	}

	return resultStr + unit
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

// parseNumberWithUnit parses a number with optional unit (e.g., "10px", "5", "1.5em")
// Returns (number, unit) or (nil, "") if not a valid number
func parseNumberWithUnit(s string) (*float64, string) {
	if s == "" {
		return nil, ""
	}

	// Regular expression to match optional sign, digits, optional decimal, and optional unit
	re := regexp.MustCompile(`^(-?\d+(?:\.\d+)?)(.*?)$`)
	matches := re.FindStringSubmatch(s)

	if matches == nil || len(matches) < 2 {
		return nil, ""
	}

	numStr := matches[1]
	unit := strings.TrimSpace(matches[2])

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, ""
	}

	return &num, unit
}

// parsePercentage parses a percentage value and returns it as a decimal (0-1)
func parsePercentage(s string) float64 {
	s = strings.TrimSpace(s)
	num, unit := parseNumberWithUnit(s)
	if num == nil {
		return 0
	}

	if unit == "%" {
		return *num / 100
	}

	return *num / 100 // assume percentage if no unit
}

// renderMixinCall renders a mixin call by applying the mixin's declarations
func (r *Renderer) renderMixinCall(call *ast.MixinCall) {
	// Get the mixin name (last element in path)
	if len(call.Path) == 0 {
		return
	}

	mixinName := call.Path[len(call.Path)-1]

	// Look up the mixin definition
	_, ok := r.mixins[mixinName]
	if !ok {
		// Mixin not found, skip it
		return
	}

	// TODO: Handle parametric mixins (mixin arguments)
	// For now, just copy the declarations

	// Note: We don't output anything for mixin calls directly.
	// The declarations from the mixin are applied by the parent rule rendering.
	// This is handled in renderRule where we process nested rules/mixins.
}
