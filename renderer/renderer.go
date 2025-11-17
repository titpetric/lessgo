package renderer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
	"github.com/sourcegraph/lessgo/functions"
	"github.com/sourcegraph/lessgo/parser"
)

// Renderer converts an AST to CSS
type Renderer struct {
	output bytes.Buffer
	indent int
	vars   *parser.Stack          // Stack-based variable scoping
	mixins map[string][]*ast.Rule // Store mixin definitions by name (can have multiple variants with guards)
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		vars:   parser.NewStack(make(map[string]ast.Value)),
		mixins: make(map[string][]*ast.Rule),
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
			// Check if this is a mixin definition (class or id selector that could be used as mixin)
			if len(rule.Selector.Parts) == 1 {
				selector := rule.Selector.Parts[0]
				// Extract mixin name from .classname or #id format
				if (strings.HasPrefix(selector, ".") || strings.HasPrefix(selector, "#")) && !strings.Contains(selector, " ") {
					name := selector[1:] // Remove . or #
					r.mixins[name] = append(r.mixins[name], rule)
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
		r.renderAtRuleWithContext(s, parentSelector)
	case *ast.MixinCall:
		r.renderMixinCall(s)
	case *ast.DeclarationStmt:
		r.renderDeclarationStmt(s, parentSelector)
	}
}

// renderRule renders a CSS rule
func (r *Renderer) renderRule(rule *ast.Rule, parentSelector string) {
	// Skip parametric mixin definitions - they're not output to CSS
	if len(rule.Parameters) > 0 {
		return
	}

	// Skip mixin definitions with guards - they're only used when called
	if rule.Guard != nil {
		return
	}

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
				if mixins, found := r.mixins[mixinName]; found {
					// Find the first matching mixin variant (check guards in order)
					for _, mixin := range mixins {
						// Check if mixin's guard condition is satisfied
						if r.evaluateGuard(mixin.Guard) {
							// Bind arguments to parameters if this is a parametric mixin
							mixinDecls := r.bindMixinArguments(mixin, mixinCall.Arguments)
							// Insert mixin declarations at the beginning
							allDeclarations = append(mixinDecls, allDeclarations...)
							break // Apply only the first matching mixin variant
						}
					}
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
			// Resolve interpolation in property names
			property := r.resolveInterpolation(decl.Property)
			r.output.WriteString(property)
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
		// Resolve interpolation in selectors
		part = r.resolveInterpolation(part)

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

// resolveInterpolation replaces @{varname} with variable values in a string
func (r *Renderer) resolveInterpolation(input string) string {
	// Find and replace all @{...} patterns
	re := regexp.MustCompile(`@\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name from @{varname}
		varName := match[2 : len(match)-1] // Remove @{ and }

		// Look up variable
		if val, ok := r.vars.Lookup(varName); ok {
			return r.renderValue(val)
		}
		// If not found, return the original (though this is probably an error)
		return match
	})
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
		if val, ok := r.vars.Lookup(v.Name); ok {
			return r.renderValue(val)
		}
		return "@" + v.Name // Fallback
	case *ast.Interpolation:
		return r.renderValue(v.Expression)
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
	// For type checking functions, evaluate directly on AST values
	if isTypeCheckingFunction(fn.Name) {
		return r.evaluateTypeCheckingFunction(fn)
	}

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

// resolveVariableValue resolves a variable to its value, if it's a variable
func (r *Renderer) resolveVariableValue(v ast.Value) ast.Value {
	if varRef, ok := v.(*ast.Variable); ok {
		if val, ok := r.vars.Lookup(varRef.Name); ok {
			return r.resolveVariableValue(val) // Recurse in case of nested variables
		}
	}
	return v
}

// isTypeCheckingFunction returns true if the function is a type checking function
func isTypeCheckingFunction(name string) bool {
	switch name {
	case "isnumber", "isstring", "iscolor", "iskeyword", "isurl", "ispixel",
		"isem", "ispercentage", "isunit", "isruleset", "islist", "boolean",
		"length", "extract", "range", "escape", "e":
		return true
	}
	return false
}

// evaluateTypeCheckingFunction evaluates type checking functions on AST values
func (r *Renderer) evaluateTypeCheckingFunction(fn *ast.FunctionCall) string {
	name := fn.Name

	// Get the rendered arguments and resolve variables
	args := []string{}
	astArgs := []ast.Value{}
	expandedArgs := []string{} // For variable expansion check
	var expandedFromVar bool      // Did we expand any arguments from variables?
	for _, arg := range fn.Arguments {
		// Check if this argument is a variable reference
		_, isVarRef := arg.(*ast.Variable)
		
		// Resolve variables to their values
		resolvedArg := r.resolveVariableValue(arg)
		astArgs = append(astArgs, resolvedArg)
		
		// If the argument WAS a variable and resolves to a List, expand it
		// This is used to detect if the function should be evaluated or output literally
		if isVarRef {
			if list, ok := resolvedArg.(*ast.List); ok {
				expandedFromVar = true
				for _, item := range list.Values {
					expandedArgs = append(expandedArgs, r.renderValue(item))
				}
			} else {
				expandedArgs = append(expandedArgs, r.renderValue(resolvedArg))
			}
		} else {
			expandedArgs = append(expandedArgs, r.renderValue(resolvedArg))
		}
		
		args = append(args, r.renderValue(resolvedArg))
	}
	
	// Special handling: if a variable argument expanded to a list, output function call literally
	// This happens when you pass a list variable to a function expecting a single argument
	// Example: islist(@list) where @list: 1, 2, 3 should output islist(1, 2, 3) literally
	if expandedFromVar && len(expandedArgs) != len(astArgs) {
		// Variable expanded to multiple arguments - output function call literally
		return fn.Name + "(" + strings.Join(expandedArgs, ", ") + ")"
	}

	switch name {
	case "isnumber":
		if len(astArgs) != 1 {
			return ""
		}
		if r.isNumberAST(astArgs[0]) {
			return "true"
		}
		return "false"
	case "isstring":
		if len(astArgs) != 1 {
			return ""
		}
		if r.isStringAST(astArgs[0]) {
			return "true"
		}
		return "false"
	case "iscolor":
		if len(astArgs) != 1 {
			return ""
		}
		if r.isColorAST(astArgs[0]) {
			return "true"
		}
		return "false"
	case "iskeyword":
		if len(astArgs) != 1 {
			return ""
		}
		// In LESS, any unquoted identifier/literal is a keyword
		if r.isKeywordAST(astArgs[0]) {
			return "true"
		}
		return "false"
	case "isurl":
		if len(args) != 1 {
			return ""
		}
		return functions.IsURLFunction(args[0])
	case "ispixel":
		if len(args) != 1 {
			return ""
		}
		return functions.IsPixelFunction(args[0])
	case "isem":
		if len(args) != 1 {
			return ""
		}
		return functions.IsEmFunction(args[0])
	case "ispercentage":
		if len(args) != 1 {
			return ""
		}
		return functions.IsPercentageFunction(args[0])
	case "isunit":
		if len(args) != 2 {
			return ""
		}
		return functions.IsUnitFunction(args[0], args[1])
	case "isruleset":
		if len(args) != 1 {
			return ""
		}
		return functions.IsRulesetFunction(args[0])
	case "islist":
		if len(astArgs) != 1 {
			return ""
		}
		if r.isListAST(astArgs[0]) {
			return "true"
		}
		return "false"
	case "boolean":
		if len(args) != 1 {
			return ""
		}
		return functions.Boolean(args[0])
	case "length":
		if len(args) != 1 {
			return ""
		}
		return r.lengthAST(astArgs[0])
	case "extract":
		if len(args) != 2 {
			return ""
		}
		return functions.Extract(args[0], args[1])
	case "range":
		if len(args) < 2 {
			return ""
		}
		if len(args) == 2 {
			return functions.Range(args[0], args[1])
		}
		return functions.Range(args[0], args[1], args[2])
	case "escape":
		if len(args) != 1 {
			return ""
		}
		return functions.Escape(args[0])
	case "e":
		if len(args) != 1 {
			return ""
		}
		return functions.E(args[0])
	}
	return ""
}

// isNumberAST checks if an AST value is a number
func (r *Renderer) isNumberAST(v ast.Value) bool {
	switch val := v.(type) {
	case *ast.Literal:
		return val.Type == ast.UnitLiteral || val.Type == ast.NumberLiteral
	case *ast.BinaryOp:
		return true // Binary operations always return numbers
	default:
		return false
	}
}

// isStringAST checks if an AST value is a string
func (r *Renderer) isStringAST(v ast.Value) bool {
	switch val := v.(type) {
	case *ast.Literal:
		// Only quoted strings are strings in LESS type system
		// Unquoted keywords/identifiers are keywords, not strings
		return val.Type == ast.StringLiteral
	default:
		return false
	}
}

// isColorAST checks if an AST value is a color
func (r *Renderer) isColorAST(v ast.Value) bool {
	switch val := v.(type) {
	case *ast.Literal:
		if val.Type == ast.ColorLiteral {
			return true
		}
		// Check if it's a named color keyword
		if val.Type == ast.KeywordLiteral {
			return functions.IsColor(val.Value)
		}
		return false
	default:
		return false
	}
}

// isKeywordAST checks if an AST value is a keyword (any unquoted identifier/literal)
func (r *Renderer) isKeywordAST(v ast.Value) bool {
	switch val := v.(type) {
	case *ast.Literal:
		// Keywords are unquoted literals: numbers, keywords, colors, etc.
		// NOT strings (quoted literals)
		return val.Type != ast.StringLiteral
	default:
		return false
	}
}

// isListAST checks if an AST value is a list
func (r *Renderer) isListAST(v ast.Value) bool {
	switch v.(type) {
	case *ast.List:
		return true
	default:
		return false
	}
}

// lengthAST returns the length of an AST value
// In LESS, length() returns:
// - 1 for quoted strings (they're single values)
// - The number of items in a list
// - 1 for single values
func (r *Renderer) lengthAST(v ast.Value) string {
	switch val := v.(type) {
	case *ast.Literal:
		// Quoted strings are single values, so length is 1
		// (not the character count)
		return "1"
	case *ast.List:
		return strconv.Itoa(len(val.Values))
	default:
		return "1"
	}
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
	case "isnumber":
		if len(args) != 1 {
			return ""
		}
		return functions.IsNumberFunction(args[0])
	case "isstring":
		if len(args) != 1 {
			return ""
		}
		return functions.IsStringFunction(args[0])
	case "iscolor":
		if len(args) != 1 {
			return ""
		}
		return functions.IsColorFunction(args[0])
	case "iskeyword":
		if len(args) != 1 {
			return ""
		}
		return functions.IsKeywordFunction(args[0])
	case "isurl":
		if len(args) != 1 {
			return ""
		}
		return functions.IsURLFunction(args[0])
	case "ispixel":
		if len(args) != 1 {
			return ""
		}
		return functions.IsPixelFunction(args[0])
	case "isem":
		if len(args) != 1 {
			return ""
		}
		return functions.IsEmFunction(args[0])
	case "ispercentage":
		if len(args) != 1 {
			return ""
		}
		return functions.IsPercentageFunction(args[0])
	case "isunit":
		if len(args) != 2 {
			return ""
		}
		return functions.IsUnitFunction(args[0], args[1])
	case "isruleset":
		if len(args) != 1 {
			return ""
		}
		return functions.IsRulesetFunction(args[0])
	case "boolean":
		if len(args) != 1 {
			return ""
		}
		return functions.Boolean(args[0])
	case "length":
		if len(args) != 1 {
			return ""
		}
		return functions.Length(args[0])
	case "extract":
		if len(args) != 2 {
			return ""
		}
		return functions.Extract(args[0], args[1])
	case "range":
		if len(args) < 2 {
			return ""
		}
		if len(args) == 2 {
			return functions.Range(args[0], args[1])
		}
		return functions.Range(args[0], args[1], args[2])
	case "escape":
		if len(args) != 1 {
			return ""
		}
		return functions.Escape(args[0])
	case "e":
		if len(args) != 1 {
			return ""
		}
		return functions.E(args[0])
	case "%":
		if len(args) < 1 {
			return ""
		}
		return functions.Format(args[0], args[1:]...)
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
	r.vars.Set(decl.Name, decl.Value)
}

// renderAtRule renders an at-rule
// renderAtRuleWithContext renders an at-rule with awareness of parent selector context
// This handles LESS-style @media nesting where media queries can contain bare declarations
func (r *Renderer) renderAtRuleWithContext(rule *ast.AtRule, parentSelector string) {
	// For @media with nested declarations in a rule context, we need to bubble up the @media
	// and wrap the declarations in the parent selector
	if rule.Name == "media" && parentSelector != "" && rule.Block != nil {
		if stmts, ok := rule.Block.([]ast.Statement); ok {
			// Check if this @media contains any declaration statements
			hasDeclarations := false
			for _, stmt := range stmts {
				if _, ok := stmt.(*ast.DeclarationStmt); ok {
					hasDeclarations = true
					break
				}
			}

			if hasDeclarations {
				// Bubble up the @media and wrap declarations in parent rule
				r.output.WriteString("@media ")
				r.output.WriteString(rule.Parameters)
				r.output.WriteString(" {\n")

				for _, stmt := range stmts {
					switch s := stmt.(type) {
					case *ast.DeclarationStmt:
						// Render declaration wrapped in parent selector
						r.output.WriteString(parentSelector)
						r.output.WriteString(" {\n")
						r.output.WriteString("  ")
						property := r.resolveInterpolation(s.Declaration.Property)
						r.output.WriteString(property)
						r.output.WriteString(": ")
						r.output.WriteString(r.renderValue(s.Declaration.Value))
						r.output.WriteString(";\n")
						r.output.WriteString("}\n")
					case *ast.Rule:
						// Nested rules inside @media get rendered normally (with parent context)
						r.renderStatement(s, parentSelector)
					default:
						r.renderStatement(s, "")
					}
				}

				r.output.WriteString("}\n")
				return
			}
		}
	}

	// Default at-rule rendering (no parent context or not a special case)
	r.renderAtRule(rule)
}

// renderAtRule renders an at-rule without context
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

// renderDeclarationStmt renders a bare declaration statement (should only appear in at-rule blocks)
func (r *Renderer) renderDeclarationStmt(stmt *ast.DeclarationStmt, parentSelector string) {
	// This shouldn't be called in normal rendering, but if it is, output the declaration
	r.output.WriteString(parentSelector)
	r.output.WriteString(" {\n")
	r.output.WriteString("  ")
	property := r.resolveInterpolation(stmt.Declaration.Property)
	r.output.WriteString(property)
	r.output.WriteString(": ")
	r.output.WriteString(r.renderValue(stmt.Declaration.Value))
	r.output.WriteString(";\n")
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

// evaluateGuard checks if a guard condition is satisfied
func (r *Renderer) evaluateGuard(guard *ast.Guard) bool {
	if guard == nil {
		return true
	}

	// For @when: all conditions must be true (AND logic)
	// For @unless: all conditions must be false (NOT AND logic)
	allSatisfied := true
	for _, cond := range guard.Conditions {
		if !r.evaluateCondition(cond) {
			allSatisfied = false
			break
		}
	}

	if guard.IsWhen {
		return allSatisfied
	} else {
		// @unless: return true if condition is NOT satisfied
		return !allSatisfied
	}
}

// evaluateCondition checks if a single guard condition is satisfied
func (r *Renderer) evaluateCondition(cond *ast.GuardCondition) bool {
	leftVal := r.renderValue(cond.Left)
	rightVal := r.renderValue(cond.Right)

	switch cond.Operator {
	case "=":
		return leftVal == rightVal
	case "!=":
		return leftVal != rightVal
	case "<":
		return r.compareNumeric(leftVal, rightVal) < 0
	case "<=":
		return r.compareNumeric(leftVal, rightVal) <= 0
	case ">":
		return r.compareNumeric(leftVal, rightVal) > 0
	case ">=":
		return r.compareNumeric(leftVal, rightVal) >= 0
	}
	return false
}

// compareNumeric compares two numeric values
// Returns negative if a < b, 0 if a == b, positive if a > b
func (r *Renderer) compareNumeric(a, b string) int {
	aNum := parseNumericValue(a)
	bNum := parseNumericValue(b)

	if aNum < bNum {
		return -1
	} else if aNum > bNum {
		return 1
	}
	return 0
}

// parseNumericValue extracts the numeric part of a value (e.g., "10px" -> 10)
func parseNumericValue(val string) float64 {
	// Extract digits and decimal point
	numStr := ""
	for _, ch := range val {
		if (ch >= '0' && ch <= '9') || ch == '.' || (ch == '-' && numStr == "") {
			numStr += string(ch)
		} else {
			break
		}
	}

	if numStr == "" {
		return 0
	}

	// Simple string to float conversion
	var result float64
	fmt.Sscanf(numStr, "%f", &result)
	return result
}

// bindMixinArguments binds mixin call arguments to parameter names
// Returns a copy of mixin declarations with parameters replaced by argument values
func (r *Renderer) bindMixinArguments(mixin *ast.Rule, args []ast.Value) []ast.Declaration {
	if len(mixin.Parameters) == 0 {
		// No parameters - just return the declarations as-is
		return mixin.Declarations
	}

	// Push a new scope for mixin parameters
	r.vars.Push(nil) // Push a new empty scope

	// Bind arguments to parameters in the new scope
	for i, param := range mixin.Parameters {
		if i < len(args) {
			// Remove @ prefix if present
			paramName := param
			if strings.HasPrefix(paramName, "@") {
				paramName = paramName[1:]
			}
			r.vars.Set(paramName, args[i])
		}
	}

	// Render declarations with bound parameters and create new literals with rendered values
	renderedDecls := make([]ast.Declaration, len(mixin.Declarations))
	for i, decl := range mixin.Declarations {
		// Render the value through the renderer with bound parameters in scope
		renderedValue := r.renderValue(decl.Value)
		// Create a literal with the rendered value
		renderedDecls[i] = ast.Declaration{
			Property: decl.Property,
			Value:    &ast.Literal{Type: ast.KeywordLiteral, Value: renderedValue},
		}
	}

	// Pop the scope
	r.vars.Pop()

	return renderedDecls
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
