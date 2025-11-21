package dst

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/titpetric/lessgo/renderer"
)

// Renderer walks a DST Document and renders it to CSS
type Renderer struct {
	output bytes.Buffer
	// Variables and mixins collected during rendering
	variables map[string]string
	mixins    map[string]*Node // selector -> mixin node
	// Current selector context (for nested rules)
	selectorStack []string
	// All nodes for lookups
	allNodes []*Node
}

// NewRenderer creates a new renderer
func NewRenderer() *Renderer {
	return &Renderer{
		variables:     make(map[string]string),
		mixins:        make(map[string]*Node),
		selectorStack: make([]string, 0),
	}
}

// Render renders a Document to CSS string
func (r *Renderer) Render(doc *Document) string {
	r.output.Reset()
	r.variables = make(map[string]string)
	r.mixins = make(map[string]*Node)
	r.selectorStack = make([]string, 0)

	// First pass: collect raw variables and mixins (without evaluation)
	for _, node := range doc.Nodes {
		r.collectRawVariablesAndMixins(node)
	}

	// Second pass: evaluate variable values (now that all raw variables are collected)
	r.evaluateVariables()

	// Third pass: render
	for _, node := range doc.Nodes {
		r.renderNode(node)
	}

	return r.output.String()
}

// collectRawVariablesAndMixins recursively collects raw (unevaluated) variables and mixins
func (r *Renderer) collectRawVariablesAndMixins(node *Node) {
	if node == nil {
		return
	}

	if node.Type == NodeVariable {
		// Variable: store name without @ prefix, raw value
		varName := strings.TrimPrefix(node.Name, "@")
		r.variables[varName] = node.Value
	}

	if node.Type == NodeDeclaration {
		// Check if this looks like a mixin definition
		// Mixins start with . or # and typically don't have pseudo-selectors
		if r.isMixinDefinition(node) {
			// Store by selector
			for _, sel := range node.Names() {
				r.mixins[sel] = node
			}
		}
	}

	// Recurse into children
	for _, child := range node.Children {
		r.collectRawVariablesAndMixins(child)
	}
}

// evaluateVariables evaluates all collected variable values (handles substitution and functions)
func (r *Renderer) evaluateVariables() {
	// Iterate multiple times to handle variable-to-variable references
	maxIterations := 10
	for iteration := 0; iteration < maxIterations; iteration++ {
		changed := false
		for varName, varValue := range r.variables {
			// Evaluate this variable's value
			evaluatedValue := r.evaluateFunctionsOnly(varValue)
			if evaluatedValue != varValue {
				r.variables[varName] = evaluatedValue
				changed = true
			}
		}
		if !changed {
			break
		}
	}
}

// evaluateFunctionsOnly evaluates functions and variable references in a value
// without recursing through evaluateValue (to avoid infinite loops)
// This is used during variable evaluation to avoid infinite loops with variable references
func (r *Renderer) evaluateFunctionsOnly(value string) string {
	result := value
	// Just evaluate functions - don't substitute variables yet, to avoid breaking argument parsing
	// Variables will be substituted inside evalOneFunction -> callFunction -> evaluateValue
	result = r.evaluateFunctions(result)
	return result
}

// isMixinDefinition checks if a declaration looks like a mixin
// Mixins are definitions that look like they expect to be called, but pure selectors are just rules
// Since we can't tell from the DST alone, we say nothing is a mixin - mixins are called elsewhere
func (r *Renderer) isMixinDefinition(node *Node) bool {
	// For now, treat everything as a potential selector/rule, not a mixin definition
	// Mixins will be handled when they are invoked
	return false
}

// renderNode renders a single node based on its type
func (r *Renderer) renderNode(node *Node) {
	if node == nil {
		return
	}

	switch node.Type {
	case NodeDeclaration:
		r.renderDeclaration(node)
	case NodeProperty:
		r.renderProperty(node)
	case NodeVariable:
		// Variables are collected in first pass, skip rendering
	case NodeAtRule:
		r.renderAtRule(node)
	case NodeMixin:
		// Mixins are not rendered directly; they are evaluated/called
	case NodeComment:
		r.renderComment(node)
	}
}

// renderDeclaration renders a declaration (rule with selector and properties/nested rules)
func (r *Renderer) renderDeclaration(node *Node) {
	if node == nil || node.SelectorsRaw == "" {
		return
	}

	// Skip if this is a mixin definition
	if r.isMixinDefinition(node) {
		return
	}

	// Get selectors for this node
	selectors := node.Names()
	if len(selectors) == 0 {
		return
	}

	// Separate properties, comments, and nested rules
	properties := make([]*Node, 0)
	comments := make([]*Node, 0)
	nestedRules := make([]*Node, 0)

	for _, child := range node.Children {
		if child.Type == NodeProperty {
			properties = append(properties, child)
		} else if child.Type == NodeComment {
			comments = append(comments, child)
		} else if child.Type == NodeDeclaration {
			nestedRules = append(nestedRules, child)
		}
	}

	// Build full selector list with parent context
	fullSelectors := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		fullSelectors = append(fullSelectors, r.buildSelector(selector))
	}

	// If there are properties, render as one rule with all selectors (comma-separated)
	if len(properties) > 0 || len(comments) > 0 {
		r.output.WriteString(strings.Join(fullSelectors, ",\n"))
		r.output.WriteString(" {\n")

		for _, prop := range properties {
			r.renderPropertyIndented(prop, 2)
		}

		for _, comment := range comments {
			r.renderCommentIndented(comment, 2)
		}

		r.output.WriteString("}\n")
	}

	// Render nested rules with updated selector context
	if len(nestedRules) > 0 {
		for _, fullSelector := range fullSelectors {
			r.selectorStack = append(r.selectorStack, fullSelector)
			for _, nested := range nestedRules {
				r.renderDeclaration(nested)
			}
			r.selectorStack = r.selectorStack[:len(r.selectorStack)-1]
		}
	}
}

// renderProperty renders a property declaration with indentation
func (r *Renderer) renderPropertyIndented(node *Node, indent int) {
	if node == nil || node.Name == "" {
		return
	}

	indentStr := strings.Repeat(" ", indent)
	r.output.WriteString(indentStr)
	r.output.WriteString(node.Name)
	r.output.WriteString(": ")
	r.output.WriteString(r.evaluateValue(node.Value))
	r.output.WriteString(";\n")
}

// renderProperty renders a standalone property (shouldn't normally occur at top level)
func (r *Renderer) renderProperty(node *Node) {
	r.renderPropertyIndented(node, 0)
}

// renderAtRule renders an at-rule (@media, @import, etc)
func (r *Renderer) renderAtRule(node *Node) {
	if node == nil {
		return
	}

	r.output.WriteString("@")
	r.output.WriteString(node.Name)

	if node.Value != "" {
		r.output.WriteString(" ")
		r.output.WriteString(node.Value)
	}

	// At-rules like @import end with ;
	if node.Name == "import" && len(node.Children) == 0 {
		r.output.WriteString(";\n")
		return
	}

	r.output.WriteString(" {\n")

	// Render children
	for _, child := range node.Children {
		if child.Type == NodeDeclaration {
			// Indent nested rules in at-rule
			r.renderDeclarationIndented(child, 2)
		}
	}

	r.output.WriteString("}\n")
}

// renderDeclarationIndented renders a declaration with indentation
func (r *Renderer) renderDeclarationIndented(node *Node, indent int) {
	if node == nil || node.SelectorsRaw == "" {
		return
	}

	selectors := node.Names()
	indentStr := strings.Repeat(" ", indent)

	for _, selector := range selectors {
		r.output.WriteString(indentStr)
		r.output.WriteString(selector)
		r.output.WriteString(" {\n")

		for _, child := range node.Children {
			if child.Type == NodeProperty {
				r.renderPropertyIndented(child, indent+2)
			}
		}

		r.output.WriteString(indentStr)
		r.output.WriteString("}\n")
	}
}

// renderComment renders a comment at top level (only multiline comments)
func (r *Renderer) renderComment(node *Node) {
	if node == nil || node.Value == "" {
		return
	}
	// Skip single-line comments (they start with //)
	if strings.HasPrefix(strings.TrimSpace(node.Value), "//") {
		return
	}
	r.output.WriteString(node.Value)
	r.output.WriteString("\n")
}

// renderCommentIndented renders a comment with indentation (only multiline comments)
func (r *Renderer) renderCommentIndented(node *Node, indent int) {
	if node == nil || node.Value == "" {
		return
	}
	// Skip single-line comments (they start with //)
	if strings.HasPrefix(strings.TrimSpace(node.Value), "//") {
		return
	}
	indentStr := strings.Repeat(" ", indent)
	r.output.WriteString(indentStr)
	r.output.WriteString(node.Value)
	r.output.WriteString("\n")
}

// buildSelector combines parent context with current selector
func (r *Renderer) buildSelector(selector string) string {
	// Clean up selector - remove comments and trim
	selector = cleanSelector(selector)
	
	if len(r.selectorStack) == 0 {
		return selector
	}

	// Handle parent selector &
	if strings.HasPrefix(selector, "&") {
		// Replace & with parent selector
		parent := r.selectorStack[len(r.selectorStack)-1]
		return strings.Replace(selector, "&", parent, 1)
	}

	// Descendant combinator (space)
	parent := r.selectorStack[len(r.selectorStack)-1]
	return parent + " " + selector
}

// cleanSelector removes comments from selectors
func cleanSelector(selector string) string {
	// Remove lines starting with //
	lines := strings.Split(selector, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//") && trimmed != "" {
			cleaned = append(cleaned, line)
		}
	}
	result := strings.Join(cleaned, "\n")
	// Clean up extra whitespace
	result = strings.TrimSpace(result)
	// Replace multiple spaces with single space
	result = strings.Join(strings.Fields(result), " ")
	return result
}

// evaluateValue substitutes variables and evaluates functions in a value
func (r *Renderer) evaluateValue(value string) string {
	result := value

	// Simple variable substitution: @varname -> value
	// Sort by length (longest first) to avoid partial replacements
	// e.g., @bg-light must be replaced before @bg
	varNames := make([]string, 0, len(r.variables))
	for varName := range r.variables {
		varNames = append(varNames, varName)
	}
	// Sort by length descending
	for i := 0; i < len(varNames); i++ {
		for j := i + 1; j < len(varNames); j++ {
			if len(varNames[i]) < len(varNames[j]) {
				varNames[i], varNames[j] = varNames[j], varNames[i]
			}
		}
	}

	for _, varName := range varNames {
		// Replace @varname with the variable value
		result = strings.ReplaceAll(result, "@"+varName, r.variables[varName])
	}

	// Evaluate math operations (must be after variable substitution)
	result = r.evaluateMath(result)

	// Basic function evaluation
	result = r.evaluateFunctions(result)

	return result
}

// evaluateFunctions evaluates LESS functions in values
func (r *Renderer) evaluateFunctions(value string) string {
	// Process functions from innermost to outermost
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		newValue := r.evalOneFunction(value)
		if newValue == value {
			break // No function was evaluated
		}
		value = newValue
	}
	return value
}

// evalOneFunction evaluates one function call in the value
func (r *Renderer) evalOneFunction(value string) string {
	// Find function call: functionname(...)
	// We need to handle nested parentheses

	// First, find where a function call starts
	funcPattern := regexp.MustCompile(`(\w+)\s*\(`)
	matches := funcPattern.FindStringSubmatchIndex(value)
	if matches == nil {
		return value
	}

	funcName := value[matches[2]:matches[3]]
	matchEnd := matches[1] // End of the entire regex match (position of opening paren)

	// Find the actual opening paren position (it's just before the end of the match)
	parenStart := matchEnd - 1 // Position of the opening paren

	// Find matching closing paren, handling nesting
	parenDepth := 0
	parenEnd := -1
	for i := parenStart; i < len(value); i++ {
		if value[i] == '(' {
			parenDepth++
		} else if value[i] == ')' {
			parenDepth--
			if parenDepth == 0 {
				parenEnd = i
				break
			}
		}
	}

	if parenEnd == -1 {
		// No matching paren found
		return value
	}

	// Extract arguments
	argsStr := value[parenStart+1 : parenEnd]

	// Parse arguments
	args := r.parseArguments(argsStr)

	// Evaluate the function
	result := r.callFunction(funcName, args)
	if result == "" {
		return value // Function not recognized, return as-is
	}

	// Replace the function call with the result
	return value[:matches[0]] + result + value[parenEnd+1:]
}

// parseArguments splits function arguments, respecting nested parens
func (r *Renderer) parseArguments(argsStr string) []string {
	var args []string
	var current string
	parenDepth := 0

	for _, ch := range argsStr {
		switch ch {
		case '(':
			parenDepth++
			current += string(ch)
		case ')':
			parenDepth--
			current += string(ch)
		case ',':
			if parenDepth == 0 {
				args = append(args, strings.TrimSpace(current))
				current = ""
			} else {
				current += string(ch)
			}
		default:
			current += string(ch)
		}
	}
	if current != "" {
		args = append(args, strings.TrimSpace(current))
	}
	return args
}

// callFunction calls a LESS function
func (r *Renderer) callFunction(name string, args []string) string {
	name = strings.ToLower(name)

	// Substitute variables in arguments
	for i, arg := range args {
		args[i] = r.evaluateValue(arg)
	}

	switch name {
	// Color definition functions
	case "rgb":
		if len(args) >= 3 {
			return renderer.RGB(args[0], args[1], args[2])
		}
	case "rgba":
		if len(args) >= 4 {
			return renderer.RGBA(args[0], args[1], args[2], args[3])
		}
	case "hsl":
		if len(args) >= 3 {
			return renderer.HSL(args[0], args[1], args[2])
		}
	case "hsla":
		if len(args) >= 4 {
			return renderer.HSLA(args[0], args[1], args[2], args[3])
		}

	// Color channel extraction
	case "hue":
		if len(args) >= 1 {
			return renderer.Hue(args[0])
		}
	case "saturation":
		if len(args) >= 1 {
			return renderer.Saturation(args[0])
		}
	case "lightness":
		if len(args) >= 1 {
			return renderer.Lightness(args[0])
		}
	case "luma", "luminance":
		if len(args) >= 1 {
			return renderer.LumaFunction(args[0])
		}
	case "red":
		if len(args) >= 1 {
			return renderer.Red(args[0])
		}
	case "green":
		if len(args) >= 1 {
			return renderer.Green(args[0])
		}
	case "blue":
		if len(args) >= 1 {
			return renderer.Blue(args[0])
		}
	case "alpha":
		if len(args) >= 1 {
			return renderer.Alpha(args[0])
		}

	// Color operations
	case "lighten":
		if len(args) >= 2 {
			return renderer.Lighten(args[0], args[1])
		}
	case "darken":
		if len(args) >= 2 {
			return renderer.Darken(args[0], args[1])
		}
	case "saturate":
		if len(args) >= 2 {
			return renderer.Saturate(args[0], args[1])
		}
	case "desaturate":
		if len(args) >= 2 {
			return renderer.Desaturate(args[0], args[1])
		}
	case "spin":
		if len(args) >= 2 {
			return renderer.Spin(args[0], args[1])
		}
	case "mix":
		if len(args) >= 2 {
			weight := "50%"
			if len(args) >= 3 {
				weight = args[2]
			}
			return renderer.Mix(args[0], args[1], weight)
		}
	case "greyscale":
		if len(args) >= 1 {
			return renderer.Greyscale(args[0])
		}

	// Math functions
	case "ceil":
		if len(args) >= 1 {
			return renderer.Ceil(args[0])
		}
	case "floor":
		if len(args) >= 1 {
			return renderer.Floor(args[0])
		}
	case "round":
		if len(args) >= 1 {
			return renderer.Round(args[0])
		}
	case "abs":
		if len(args) >= 1 {
			return renderer.Abs(args[0])
		}
	case "sqrt":
		if len(args) >= 1 {
			return renderer.Sqrt(args[0])
		}
	case "pow":
		if len(args) >= 2 {
			return renderer.Pow(args[0], args[1])
		}
	case "min":
		return renderer.Min(args...)
	case "max":
		return renderer.Max(args...)
	case "mod":
		if len(args) >= 2 {
			return renderer.Mod(args[0], args[1])
		}
	case "pi":
		return renderer.Pi()
	case "percentage":
		if len(args) >= 1 {
			return renderer.Percentage(args[0])
		}

	// Trigonometric functions
	case "sin":
		if len(args) >= 1 {
			return renderer.Sin(args[0])
		}
	case "cos":
		if len(args) >= 1 {
			return renderer.Cos(args[0])
		}
	case "tan":
		if len(args) >= 1 {
			return renderer.Tan(args[0])
		}
	case "asin":
		if len(args) >= 1 {
			return renderer.Asin(args[0])
		}
	case "acos":
		if len(args) >= 1 {
			return renderer.Acos(args[0])
		}
	case "atan":
		if len(args) >= 1 {
			return renderer.Atan(args[0])
		}

	// Logical functions
	case "if":
		if len(args) >= 3 {
			return renderer.If(args[0], args[1], args[2])
		}
	case "boolean":
		if len(args) >= 1 {
			return renderer.Boolean(args[0])
		}

	// String functions
	case "escape":
		if len(args) >= 1 {
			return renderer.Escape(args[0])
		}
	case "e":
		if len(args) >= 1 {
			return renderer.E(args[0])
		}
	case "replace":
		if len(args) >= 3 {
			return renderer.Replace(args[0], args[1], args[2])
		}

	// List functions
	case "length":
		if len(args) >= 1 {
			return renderer.Length(args[0])
		}
	case "extract":
		if len(args) >= 2 {
			return renderer.Extract(args[0], args[1])
		}
	case "range":
		if len(args) >= 1 {
			start := args[0]
			end := ""
			var step []string
			if len(args) >= 2 {
				end = args[1]
			}
			if len(args) >= 3 {
				step = args[2:]
			}
			return renderer.Range(start, end, step...)
		}

	// Type checking functions
	case "isnumber":
		if len(args) >= 1 {
			return renderer.IsNumberFunction(args[0])
		}
	case "isstring":
		if len(args) >= 1 {
			return renderer.IsStringFunction(args[0])
		}
	case "iscolor":
		if len(args) >= 1 {
			return renderer.IsColorFunction(args[0])
		}
	case "iskeyword":
		if len(args) >= 1 {
			return renderer.IsKeywordFunction(args[0])
		}
	case "isurl":
		if len(args) >= 1 {
			return renderer.IsURLFunction(args[0])
		}
	case "ispixel":
		if len(args) >= 1 {
			return renderer.IsPixelFunction(args[0])
		}
	case "isem":
		if len(args) >= 1 {
			return renderer.IsEmFunction(args[0])
		}
	case "ispercentage":
		if len(args) >= 1 {
			return renderer.IsPercentageFunction(args[0])
		}
	case "isunit":
		if len(args) >= 2 {
			return renderer.IsUnitFunction(args[0], args[1])
		}
	case "isruleset":
		if len(args) >= 1 {
			return renderer.IsRulesetFunction(args[0])
		}
	case "isdefined":
		if len(args) >= 1 {
			if renderer.IsDefined(args[0]) {
				return "true"
			}
			return "false"
		}

	// Color definition
	case "hsv":
		if len(args) >= 3 {
			return renderer.HSV(args[0], args[1], args[2])
		}
	case "hsva":
		if len(args) >= 4 {
			return renderer.HSVA(args[0], args[1], args[2], args[3])
		}
	case "argb":
		if len(args) >= 1 {
			return renderer.ARGB(args[0])
		}

	// More color channels
	case "hsvhue":
		if len(args) >= 1 {
			return renderer.HSVHue(args[0])
		}
	case "hsvsaturation":
		if len(args) >= 1 {
			return renderer.HSVSaturation(args[0])
		}
	case "hsvvalue":
		if len(args) >= 1 {
			return renderer.HSVValue(args[0])
		}

	// More color operations
	case "fade":
		if len(args) >= 2 {
			return renderer.Fade(args[0], args[1])
		}
	case "fadein":
		if len(args) >= 2 {
			return renderer.Fadein(args[0], args[1])
		}
	case "fadeout":
		if len(args) >= 2 {
			return renderer.Fadeout(args[0], args[1])
		}
	case "tint":
		if len(args) >= 2 {
			return renderer.Tint(args[0], args[1])
		}
	case "shade":
		if len(args) >= 2 {
			return renderer.Shade(args[0], args[1])
		}
	case "contrast":
		if len(args) >= 1 {
			return renderer.Contrast(args[0])
		}

	// Color blending
	case "multiply":
		if len(args) >= 2 {
			return renderer.Multiply(args[0], args[1])
		}
	case "screen":
		if len(args) >= 2 {
			return renderer.Screen(args[0], args[1])
		}
	case "overlay":
		if len(args) >= 2 {
			return renderer.Overlay(args[0], args[1])
		}
	case "softlight":
		if len(args) >= 2 {
			return renderer.Softlight(args[0], args[1])
		}
	case "hardlight":
		if len(args) >= 2 {
			return renderer.Hardlight(args[0], args[1])
		}
	case "difference":
		if len(args) >= 2 {
			return renderer.Difference(args[0], args[1])
		}
	case "exclusion":
		if len(args) >= 2 {
			return renderer.Exclusion(args[0], args[1])
		}
	case "average":
		if len(args) >= 2 {
			return renderer.Average(args[0], args[1])
		}
	case "negation":
		if len(args) >= 2 {
			return renderer.Negation(args[0], args[1])
		}

	// Unit functions
	case "unit":
		if len(args) >= 2 {
			return renderer.Unit(args[0], args[1])
		}
	case "get-unit":
		if len(args) >= 1 {
			return renderer.GetUnit(args[0])
		}
	case "convert":
		if len(args) >= 2 {
			return renderer.Convert(args[0], args[1])
		}

	// Misc functions
	case "color":
		if len(args) >= 1 {
			return renderer.ColorFunction(args[0])
		}
	case "format":
		if len(args) >= 1 {
			return renderer.Format(args[0], args[1:]...)
		}
	}

	return ""
}

// evaluateMath evaluates mathematical operations in a value
// e.g., "10px * 2" -> "20px", "@base + 5px" -> result
func (r *Renderer) evaluateMath(value string) string {
	// Check if the value contains math operators
	if !strings.ContainsAny(value, "+-*/%") {
		return value
	}

	// Extract and evaluate the math expression
	// Split by whitespace to identify the expression
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return value // Not enough parts for a math expression
	}

	// Try to parse and evaluate the expression
	result := renderer.EvaluateExpression(value)
	if result != "" {
		return result
	}

	return value
}

// Variables returns the collected variables (for testing/debugging)
func (r *Renderer) Variables() map[string]string {
	return r.variables
}

// Mixins returns the collected mixins (for testing/debugging)
func (r *Renderer) Mixins() map[string]*Node {
	return r.mixins
}
