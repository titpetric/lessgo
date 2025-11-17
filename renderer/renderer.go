package renderer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
	"github.com/sourcegraph/lessgo/evaluator"
	"github.com/sourcegraph/lessgo/parser"
)

// Renderer converts an AST to CSS
type Renderer struct {
	output   bytes.Buffer
	indent   int
	vars     *parser.Stack          // Stack-based variable scoping
	mixins   map[string][]*ast.Rule // Store mixin definitions by name (can have multiple variants with guards)
	extends  map[string][]string    // Map from extended selector to list of extending selectors
	allRules []*ast.Rule            // Track all rules for extend processing
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		vars:     parser.NewStack(make(map[string]ast.Value)),
		mixins:   make(map[string][]*ast.Rule),
		extends:  make(map[string][]string),
		allRules: []*ast.Rule{},
	}
}

// Render renders the stylesheet to CSS
func (r *Renderer) Render(stylesheet *ast.Stylesheet) string {
	// First pass: collect all rules, mixin definitions, and extends
	r.collectRulesAndMixins(stylesheet)
	r.collectExtends()

	// Second pass: render all statements (mixins are just normal rules)
	for _, rule := range stylesheet.Rules {
		r.renderStatement(rule, "")
	}
	return r.output.String()
}

// collectRulesAndMixins finds all rules and mixin definitions in the stylesheet
func (r *Renderer) collectRulesAndMixins(stylesheet *ast.Stylesheet) {
	for _, stmt := range stylesheet.Rules {
		if rule, ok := stmt.(*ast.Rule); ok {
			// Track all rules for extend processing
			r.allRules = append(r.allRules, rule)
			r.collectNestedRules(rule)

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

// collectNestedRules recursively collects all nested rules
func (r *Renderer) collectNestedRules(rule *ast.Rule) {
	for _, stmt := range rule.Rules {
		if nestedRule, ok := stmt.(*ast.Rule); ok {
			r.allRules = append(r.allRules, nestedRule)
			r.collectNestedRules(nestedRule)
		}
	}
}

// collectExtends builds a map of which selectors are extended by which other selectors
func (r *Renderer) collectExtends() {
	for _, rule := range r.allRules {
		for _, ext := range rule.Extends {
			// For each rule that extends another, record that the extended selector
			// should include the extending selector
			for _, selector := range rule.Selector.Parts {
				r.extends[ext.Selector] = append(r.extends[ext.Selector], selector)
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

	// Render leading comments
	r.renderComments(rule.Comments)

	selector := r.buildSelector(rule.Selector, parentSelector)

	// Apply extends: add any selectors that extend this one
	extendingSelectors := []string{}
	for _, selectorPart := range rule.Selector.Parts {
		if extenders, found := r.extends[selectorPart]; found {
			extendingSelectors = append(extendingSelectors, extenders...)
		}
	}

	if len(extendingSelectors) > 0 {
		// Add extending selectors to this rule's selector
		selector = selector + ",\n" + strings.Join(extendingSelectors, ",\n")
	}

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

// buildSelector builds the full selector from nesting, applying extends
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

	return strings.Join(parts, ",\n")
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
		// StringLiteral should be quoted
		if v.Type == ast.StringLiteral {
			quote := v.QuoteChar
			if quote == "" {
				quote = `"`
			}
			return quote + v.Value + quote
		}
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

	// Handle % format function - needs special wrapping
	if fn.Name == "%" {
		if len(args) > 0 {
			result := Format(args[0], args[1:]...)
			return `"` + result + `"`
		}
		return ""
	}

	// Evaluate all direct functions
	result := r.evaluateFunction(fn.Name, args)
	if result != "" {
		return result
	}

	// No matching function found, output literally
	return fn.Name + "(" + strings.Join(args, ", ") + ")"
}

// evaluateFunction evaluates all built-in LESS functions
func (r *Renderer) evaluateFunction(name string, args []string) string {
	switch name {
	// Math functions
	case "ceil":
		if len(args) > 0 {
			return Ceil(args[0])
		}
	case "floor":
		if len(args) > 0 {
			return Floor(args[0])
		}
	case "round":
		if len(args) > 0 {
			return Round(args[0])
		}
	case "abs":
		if len(args) > 0 {
			return Abs(args[0])
		}
	case "sqrt":
		if len(args) > 0 {
			return Sqrt(args[0])
		}
	case "pow":
		if len(args) >= 2 {
			return Pow(args[0], args[1])
		}
	case "min":
		return Min(args...)
	case "max":
		return Max(args...)
	case "mod":
		if len(args) >= 2 {
			return Mod(args[0], args[1])
		}
	case "sin":
		if len(args) > 0 {
			return Sin(args[0])
		}
	case "cos":
		if len(args) > 0 {
			return Cos(args[0])
		}
	case "tan":
		if len(args) > 0 {
			return Tan(args[0])
		}
	case "asin":
		if len(args) > 0 {
			return Asin(args[0])
		}
	case "acos":
		if len(args) > 0 {
			return Acos(args[0])
		}
	case "atan":
		if len(args) > 0 {
			return Atan(args[0])
		}
	case "pi":
		return Pi()
	case "percentage":
		if len(args) > 0 {
			return Percentage(args[0])
		}

	// String functions
	case "replace":
		if len(args) >= 3 {
			if len(args) > 3 {
				return Replace(args[0], args[1], args[2], args[3])
			}
			return Replace(args[0], args[1], args[2])
		}
	case "escape":
		if len(args) > 0 {
			return Escape(args[0])
		}
	case "e":
		if len(args) > 0 {
			return E(args[0])
		}
	case "format":
		if len(args) > 0 {
			return Format(args[0], args[1:]...)
		}

	// List functions
	case "length":
		if len(args) > 0 {
			return Length(args[0])
		}
	case "extract":
		if len(args) >= 2 {
			return Extract(args[0], args[1])
		}
	case "range":
		if len(args) >= 1 {
			if len(args) >= 3 {
				return Range(args[0], args[1], args[2])
			} else if len(args) >= 2 {
				return Range(args[0], args[1])
			}
			return Range(args[0], "")
		}

	// Color definition functions
	case "rgb":
		if len(args) >= 3 {
			return RGB(args[0], args[1], args[2])
		}
	case "rgba":
		if len(args) >= 4 {
			return RGBA(args[0], args[1], args[2], args[3])
		}
	case "hsl":
		if len(args) >= 3 {
			return HSL(args[0], args[1], args[2])
		}
	case "hsla":
		if len(args) >= 4 {
			return HSLA(args[0], args[1], args[2], args[3])
		}

	// Color channel extraction functions
	case "hue":
		if len(args) > 0 {
			return Hue(args[0])
		}
	case "saturation":
		if len(args) > 0 {
			return Saturation(args[0])
		}
	case "lightness":
		if len(args) > 0 {
			return Lightness(args[0])
		}
	case "red":
		if len(args) > 0 {
			return Red(args[0])
		}
	case "green":
		if len(args) > 0 {
			return Green(args[0])
		}
	case "blue":
		if len(args) > 0 {
			return Blue(args[0])
		}
	case "alpha":
		if len(args) > 0 {
			return Alpha(args[0])
		}
	case "luma":
		if len(args) > 0 {
			return LumaFunction(args[0])
		}
	case "luminance":
		if len(args) > 0 {
			return Luminance(args[0])
		}

	// Color manipulation functions
	case "lighten":
		if len(args) >= 2 {
			return Lighten(args[0], args[1])
		}
	case "darken":
		if len(args) >= 2 {
			return Darken(args[0], args[1])
		}
	case "saturate":
		if len(args) >= 2 {
			return Saturate(args[0], args[1])
		}
	case "desaturate":
		if len(args) >= 2 {
			return Desaturate(args[0], args[1])
		}
	case "spin":
		if len(args) >= 2 {
			return Spin(args[0], args[1])
		}
	case "mix":
		if len(args) >= 2 {
			if len(args) >= 3 {
				return Mix(args[0], args[1], args[2])
			}
			return Mix(args[0], args[1])
		}
	case "tint":
		if len(args) >= 2 {
			return Tint(args[0], args[1])
		}
	case "shade":
		if len(args) >= 2 {
			return Shade(args[0], args[1])
		}
	case "greyscale":
		if len(args) > 0 {
			return Greyscale(args[0])
		}
	case "fade":
		if len(args) >= 2 {
			return Fade(args[0], args[1])
		}
	case "fadein":
		if len(args) >= 2 {
			return Fadein(args[0], args[1])
		}
	case "fadeout":
		if len(args) >= 2 {
			return Fadeout(args[0], args[1])
		}
	case "contrast":
		if len(args) >= 1 {
			return Contrast(args[0], args[1:]...)
		}

	// HSV color functions
	case "hsv":
		if len(args) >= 3 {
			return HSV(args[0], args[1], args[2])
		}
	case "hsva":
		if len(args) >= 4 {
			return HSVA(args[0], args[1], args[2], args[3])
		}
	case "hsvhue":
		if len(args) >= 1 {
			return HSVHue(args[0])
		}
	case "hsvsaturation":
		if len(args) >= 1 {
			return HSVSaturation(args[0])
		}
	case "hsvvalue":
		if len(args) >= 1 {
			return HSVValue(args[0])
		}
	case "argb":
		if len(args) >= 1 {
			return ARGB(args[0])
		}

	// Color blending functions
	case "multiply":
		if len(args) >= 2 {
			return Multiply(args[0], args[1])
		}
	case "screen":
		if len(args) >= 2 {
			return Screen(args[0], args[1])
		}
	case "overlay":
		if len(args) >= 2 {
			return Overlay(args[0], args[1])
		}
	case "softlight":
		if len(args) >= 2 {
			return Softlight(args[0], args[1])
		}
	case "hardlight":
		if len(args) >= 2 {
			return Hardlight(args[0], args[1])
		}
	case "difference":
		if len(args) >= 2 {
			return Difference(args[0], args[1])
		}
	case "exclusion":
		if len(args) >= 2 {
			return Exclusion(args[0], args[1])
		}
	case "average":
		if len(args) >= 2 {
			return Average(args[0], args[1])
		}
	case "negation":
		if len(args) >= 2 {
			return Negation(args[0], args[1])
		}

	// Logical functions
	case "if":
		if len(args) >= 3 {
			return If(args[0], args[1], args[2])
		}

	// Utility functions
	case "color":
		if len(args) > 0 {
			return ColorFunction(args[0])
		}
	case "unit":
		if len(args) > 0 {
			if len(args) > 1 {
				return Unit(args[0], args[1])
			}
			return Unit(args[0], "")
		}
	case "get-unit":
		if len(args) > 0 {
			return GetUnit(args[0])
		}
	case "convert":
		if len(args) >= 2 {
			return Convert(args[0], args[1])
		}
	}
	return ""
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
// Type-checking functions need special AST-based evaluation to work with LESS type system
func isTypeCheckingFunction(name string) bool {
	switch name {
	case "isnumber", "isstring", "iscolor", "iskeyword", "isurl", "ispixel",
		"isem", "ispercentage", "isunit", "isruleset", "islist", "boolean", "length", "isdefined":
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
	var expandedFromVar bool   // Did we expand any arguments from variables?
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
	// EXCEPT for functions that actually operate on lists (like length)
	if expandedFromVar && len(expandedArgs) != len(astArgs) && fn.Name != "length" {
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
		return IsURLFunction(args[0])
	case "ispixel":
		if len(args) != 1 {
			return ""
		}
		return IsPixelFunction(args[0])
	case "isem":
		if len(args) != 1 {
			return ""
		}
		return IsEmFunction(args[0])
	case "ispercentage":
		if len(args) != 1 {
			return ""
		}
		return IsPercentageFunction(args[0])
	case "isunit":
		if len(args) != 2 {
			return ""
		}
		return IsUnitFunction(args[0], args[1])
	case "isruleset":
		if len(args) != 1 {
			return ""
		}
		return IsRulesetFunction(args[0])
	case "islist":
		if len(astArgs) != 1 {
			return ""
		}
		if r.isListAST(astArgs[0]) {
			return "true"
		}
		return "false"
	case "boolean":
		if len(fn.Arguments) != 1 {
			return ""
		}
		// Try to evaluate as an expression first
		// This handles cases like boolean(@v > 0)
		exprResult := r.evaluateBooleanExpression(fn.Arguments[0])
		if exprResult != "" {
			return exprResult
		}
		// Fall back to simple boolean evaluation
		return Boolean(args[0])
	case "length":
		if len(args) != 1 {
			return ""
		}
		return r.lengthAST(astArgs[0])
	case "isdefined":
		// isdefined() is just output literally with variable substitution
		// Build the function call with the resolved arguments
		argStrs := []string{}
		for _, arg := range fn.Arguments {
			argStrs = append(argStrs, r.renderValue(arg))
		}
		return fn.Name + "(" + strings.Join(argStrs, ", ") + ")"
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
			return IsColor(val.Value)
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

// evaluateBooleanExpression evaluates a boolean expression like @v > 0 or luma(@v) > 50%
// Returns "true" or "false" if the expression can be evaluated, empty string otherwise
func (r *Renderer) evaluateBooleanExpression(value ast.Value) string {
	// Check if this is an expression (binary operation, comparison, etc)
	binOp, ok := value.(*ast.BinaryOp)
	if !ok {
		// Not an expression, return empty to fall back to simple evaluation
		return ""
	}

	// Build the expression string, evaluating function calls
	// Do this BEFORE building varMap to avoid infinite recursion
	leftStr := r.evaluateExpressionValue(binOp.Left)
	rightStr := r.evaluateExpressionValue(binOp.Right)

	exprStr := leftStr + " " + binOp.Operator + " " + rightStr

	// Get variable values as strings, but only render simple values
	// to avoid infinite recursion with circular variable references
	varMap := make(map[string]string)
	envMap := r.vars.EnvMap()
	for varName, varValue := range envMap {
		// Only include simple values, skip function calls and complex expressions
		switch varValue.(type) {
		case *ast.Literal:
			varMap[varName] = r.renderValue(varValue)
		}
	}

	// Use the evaluator to evaluate the expression
	eval := evaluator.NewEvaluator(varMap)
	result, err := eval.EvalBool(exprStr)
	if err != nil {
		// Failed to evaluate, return empty
		return ""
	}

	if result {
		return "true"
	}
	return "false"
}

// evaluateExpressionValue evaluates a value that may contain function calls for use in expressions
func (r *Renderer) evaluateExpressionValue(val ast.Value) string {
	// Check if it's a function call that should be evaluated
	if fn, ok := val.(*ast.FunctionCall); ok {
		// Try to evaluate the function
		args := []string{}
		for _, arg := range fn.Arguments {
			args = append(args, r.renderValue(arg))
		}

		// Evaluate functions that return values usable in comparisons
		switch fn.Name {
		case "luma":
			if len(args) == 1 {
				result := LumaFunction(args[0])
				if result != "" {
					// Extract just the number from "0.00%" format
					return strings.TrimSuffix(result, "%")
				}
			}
		case "lighten", "darken", "saturate", "desaturate":
			// These return colors, not directly comparable numerically
			return r.renderValue(val)
		}
	}
	return r.renderValue(val)
}

// renderBinaryOp evaluates and renders a binary operation
func (r *Renderer) renderBinaryOp(op *ast.BinaryOp) string {
	leftStr := r.renderValue(op.Left)
	rightStr := r.renderValue(op.Right)

	// For comparison operators, evaluate numerically if possible
	if isComparisonOperator(op.Operator) {
		leftNum, _ := parseNumberWithUnit(leftStr)
		rightNum, _ := parseNumberWithUnit(rightStr)

		if leftNum != nil && rightNum != nil {
			// Compare as numbers
			var result bool
			switch op.Operator {
			case ">":
				result = *leftNum > *rightNum
			case "<":
				result = *leftNum < *rightNum
			case ">=":
				result = *leftNum >= *rightNum
			case "<=":
				result = *leftNum <= *rightNum
			case "==":
				result = *leftNum == *rightNum
			case "!=":
				result = *leftNum != *rightNum
			}
			if result {
				return "true"
			}
			return "false"
		}
		// Compare as strings if not numbers
		var result bool
		switch op.Operator {
		case ">":
			result = leftStr > rightStr
		case "<":
			result = leftStr < rightStr
		case ">=":
			result = leftStr >= rightStr
		case "<=":
			result = leftStr <= rightStr
		case "==":
			result = leftStr == rightStr
		case "!=":
			result = leftStr != rightStr
		}
		if result {
			return "true"
		}
		return "false"
	}

	// Try to evaluate the operation for arithmetic
	result := r.evaluateBinaryOp(op)
	if result != "" {
		return result
	}
	// Fallback to rendering as-is if we can't evaluate
	return leftStr + " " + op.Operator + " " + rightStr
}

// isComparisonOperator checks if an operator is a comparison operator
func isComparisonOperator(op string) bool {
	switch op {
	case ">", "<", ">=", "<=", "==", "!=":
		return true
	}
	return false
}

// evaluateBinaryOp evaluates a binary operation and returns the result, or empty string if not evaluable
func (r *Renderer) evaluateBinaryOp(op *ast.BinaryOp) string {
	leftStr := r.renderValue(op.Left)
	rightStr := r.renderValue(op.Right)

	// Try to parse as numbers with units for arithmetic operators
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
	// Render leading comments
	r.renderComments(decl.Comments)

	r.vars.Set(decl.Name, decl.Value)
}

// renderComments renders a list of comments
func (r *Renderer) renderComments(comments []*ast.Comment) {
	if len(comments) == 0 {
		return
	}

	// Render each comment converted to /* */ format
	for _, comment := range comments {
		r.output.WriteString("/* ")
		r.output.WriteString(comment.Text)
		r.output.WriteString(" */\n")
	}
}

// renderAtRule renders an at-rule
// renderAtRuleWithContext renders an at-rule with awareness of parent selector context
// This handles LESS-style @media/@supports nesting where queries can contain bare declarations
func (r *Renderer) renderAtRuleWithContext(rule *ast.AtRule, parentSelector string) {
	// For at-rules like @media/@supports with nested declarations in a rule context,
	// we need to bubble up the at-rule and wrap the declarations in the parent selector
	shouldBubble := (rule.Name == "media" || rule.Name == "supports") && parentSelector != "" && rule.Block != nil
	if shouldBubble {
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
				// Bubble up the at-rule and wrap declarations in parent rule
				r.output.WriteString("@")
				r.output.WriteString(rule.Name)
				r.output.WriteString(" ")
				r.output.WriteString(rule.Parameters)
				r.output.WriteString(" {\n")

				for _, stmt := range stmts {
					switch s := stmt.(type) {
					case *ast.DeclarationStmt:
						// Render declaration wrapped in parent selector with indentation
						r.output.WriteString("  ")
						r.output.WriteString(parentSelector)
						r.output.WriteString(" {\n")
						r.output.WriteString("    ")
						property := r.resolveInterpolation(s.Declaration.Property)
						r.output.WriteString(property)
						r.output.WriteString(": ")
						r.output.WriteString(r.renderValue(s.Declaration.Value))
						r.output.WriteString(";\n")
						r.output.WriteString("  }\n")
					case *ast.Rule:
						// Nested rules inside @media get rendered normally (with parent context)
						// Need to add indentation for rules inside @media
						selector := r.buildSelector(s.Selector, parentSelector)
						if len(s.Declarations) > 0 {
							r.output.WriteString("  ")
							r.output.WriteString(selector)
							r.output.WriteString(" {\n")
							for _, decl := range s.Declarations {
								r.output.WriteString("    ")
								property := r.resolveInterpolation(decl.Property)
								r.output.WriteString(property)
								r.output.WriteString(": ")
								r.output.WriteString(r.renderValue(decl.Value))
								r.output.WriteString(";\n")
							}
							r.output.WriteString("  }\n")
						}
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
