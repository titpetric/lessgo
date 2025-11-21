package dst

import (
	"bytes"
	"strings"
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

	// First pass: collect variables and mixins
	for _, node := range doc.Nodes {
		r.collectVariablesAndMixins(node)
	}

	// Second pass: render
	for _, node := range doc.Nodes {
		r.renderNode(node)
	}

	return r.output.String()
}

// collectVariablesAndMixins recursively collects variables and mixins
func (r *Renderer) collectVariablesAndMixins(node *Node) {
	if node == nil {
		return
	}

	if node.Type == NodeVariable {
		// Variable: store name without @ prefix
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
		r.collectVariablesAndMixins(child)
	}
}

// isMixinDefinition checks if a declaration looks like a mixin
// Mixins start with . or # and have no @media, :pseudo, > combinator, etc
func (r *Renderer) isMixinDefinition(node *Node) bool {
	if node == nil {
		return false
	}

	selectors := node.Names()
	if len(selectors) == 0 {
		return false
	}

	// Check first selector
	sel := selectors[0]

	// Mixins start with . or #
	if !strings.HasPrefix(sel, ".") && !strings.HasPrefix(sel, "#") {
		return false
	}

	// Not a mixin if it contains pseudo-selectors (but ::before is ok for nested rules)
	if strings.Contains(sel, ":") && !strings.Contains(sel, "::") {
		return false
	}

	// Not a mixin if it contains combinators like >, +, ~
	if strings.Contains(sel, ">") || strings.Contains(sel, "+") || strings.Contains(sel, "~") {
		return false
	}

	return true
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

	// Separate properties from nested rules
	properties := make([]*Node, 0)
	nestedRules := make([]*Node, 0)

	for _, child := range node.Children {
		if child.Type == NodeProperty {
			properties = append(properties, child)
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
	if len(properties) > 0 {
		r.output.WriteString(strings.Join(fullSelectors, ",\n"))
		r.output.WriteString(" {\n")

		for _, prop := range properties {
			r.renderPropertyIndented(prop, 2)
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

// renderComment renders a comment
func (r *Renderer) renderComment(node *Node) {
	if node == nil || node.Value == "" {
		return
	}
	r.output.WriteString(node.Value)
	r.output.WriteString("\n")
}

// buildSelector combines parent context with current selector
func (r *Renderer) buildSelector(selector string) string {
	// Normalize selector: handle pseudo-elements with space (:: before -> ::before)
	selector = r.normalizePseudoElements(selector)

	if len(r.selectorStack) == 0 {
		return selector
	}

	// Handle parent selector &
	if strings.HasPrefix(selector, "&") {
		// Replace & with parent selector
		parent := r.selectorStack[len(r.selectorStack)-1]
		parent = r.normalizePseudoElements(parent)
		return strings.Replace(selector, "&", parent, 1)
	}

	// Descendant combinator (space)
	parent := r.selectorStack[len(r.selectorStack)-1]
	parent = r.normalizePseudoElements(parent)
	return parent + " " + selector
}

// normalizePseudoElements removes spaces around :: in pseudo-elements
func (r *Renderer) normalizePseudoElements(selector string) string {
	// Remove space before and after ::
	// E.g., "blockquote:: before" -> "blockquote::before"
	selector = strings.ReplaceAll(selector, ": :", "::")
	selector = strings.ReplaceAll(selector, " :", ":")
	selector = strings.ReplaceAll(selector, ": ", ":")
	return selector
}

// evaluateValue substitutes variables and evaluates functions in a value
func (r *Renderer) evaluateValue(value string) string {
	result := value

	// Simple variable substitution: @varname -> value
	for varName, varValue := range r.variables {
		// Replace @varname with the variable value
		result = strings.ReplaceAll(result, "@"+varName, varValue)
	}

	// Basic function evaluation
	result = r.evaluateFunctions(result)

	return result
}

// evaluateFunctions evaluates simple LESS functions
func (r *Renderer) evaluateFunctions(value string) string {
	// Color functions: rgb(), rgba(), hsl(), hsla(), etc
	value = r.evaluateColorFunctions(value)

	// Math functions that can be simplified
	value = r.evaluateMathFunctions(value)

	return value
}

// evaluateColorFunctions handles color definition functions
func (r *Renderer) evaluateColorFunctions(value string) string {
	// rgb(r, g, b) -> #rrggbb
	// rgba(r, g, b, a) -> rgba(r, g, b, a)
	// For now, just return as-is - complex evaluation happens in evaluator
	return value
}

// evaluateMathFunctions handles math operations
func (r *Renderer) evaluateMathFunctions(value string) string {
	// lighten(color, amount), darken(color, amount), etc.
	// These need more complex evaluation - for now, return as-is
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
