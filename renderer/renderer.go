package renderer

import (
	"fmt"
	"strconv"

	"github.com/expr-lang/expr"
	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/evaluator"
	"github.com/titpetric/lessgo/expression"
	"github.com/titpetric/lessgo/expression/functions"
	"github.com/titpetric/lessgo/internal/strings"
)

// Renderer converts a DST into CSS output
type Renderer struct {
	resolver     *Resolver
	mixins       map[string][]*dst.Block
	mediaQueries []*MediaQuery                 // Collected media queries to render after main content
	extends      map[string][]string           // Tracks extends: extended selector -> list of extending selectors
	blockVars    map[string]*dst.BlockVariable // Detached rulesets: @var: { ... }

	// Pre-allocated buffers for zero-alloc splitting
	selectorBuf []string // For selector splitting (comma-separated)
}

// MediaQuery represents an @media block with its selector context
type MediaQuery struct {
	Condition     string     // The @media condition (e.g., "(max-width: 600px)")
	ParentSelName string     // The parent selector context
	Children      []dst.Node // The declarations within the media query
	Depth         int        // The nesting depth
}

// RenderInterface separates the responsibility to render the syntax tree
// into two steps. In the `Eval` step, the nested structures are traversed
// and then produce dst.Nodes with flattened CSS. The Render function is
// thus much more simpler, as it doesn't have to consider depth, but just
// prints a series of flattened *dst.Node values.
type RenderInterface interface {
	Eval(nodes []dst.Node) []NodeContext
	Render(nodes []NodeContext) []byte
}

type NodeContext struct {
	Buf *strings.Builder

	Node dst.Node

	Stack *Stack

	// SelNames should update by depth, for the root:
	// [a, b, p] ; with &.active; [a.active, b.active, p.active].
	// any more nested block nodes. If we kept
	// a parent Block (and a chain of parents), then
	// [a, b, p] + [&.active] should produce the resulting
	// css tree in a simple loop:
	//
	// result := []*Node
	// for _, v := range node.Names() {
	//    for _, j := range node.Children() {
	//       for _, y := range j.Names() {
	//           ctx := NodeContext{
	//               Depth: parentCtx.Depth+1,
	//               SelName: selector(v, y),
	//           }
	//           err := x.renderNode(ctx, j) // eventually render block
	//           if err != nil { // bubble up errors
	//           }
	//           result = append(result, r)
	//       }
	//    }
	// }
	SelName string
	BaseDir string // Base directory for resolving relative file paths
}

func (n *NodeContext) Depth() int {
	return n.Stack.Depth()
}

// NewRenderer creates a new CSS renderer
func NewRenderer() *Renderer {
	return &Renderer{
		mixins:       make(map[string][]*dst.Block),
		mediaQueries: make([]*MediaQuery, 0),
		extends:      make(map[string][]string),
		blockVars:    make(map[string]*dst.BlockVariable),
		selectorBuf:  make([]string, 0, 16),
	}
}

// Render converts a File into CSS output, resolving variables and expressions
func (r *Renderer) Render(file *dst.File) (string, error) {
	return r.RenderWithBaseDir(file, "")
}

// RenderWithBaseDir converts a File into CSS output with a base directory for file resolution
func (r *Renderer) RenderWithBaseDir(file *dst.File, baseDir string) (string, error) {
	// Set the base directory for image functions
	functions.BaseDir = baseDir

	r.resolver = NewResolver(file)

	// First pass: collect mixin definitions, extends, and block variables
	r.collectMixinsAndExtends(file.Nodes)
	r.collectBlockVariables(file.Nodes)

	// Second pass: render nodes
	ctx := &NodeContext{
		Buf:     &strings.Builder{},
		Node:    nil,
		Stack:   NewStack(),
		SelName: "",
		BaseDir: baseDir,
	}

	// Store block variables in the global scope so they can be checked with isruleset()
	for name := range r.blockVars {
		// Store as a special string value that isruleset() can recognize
		// The actual expansion happens via @var() calls, so we just store a placeholder
		ctx.Stack.SetGlobal(name, "{}")
	}

	if err := r.renderNodes(ctx, nil, "", file.Nodes); err != nil {
		return "", err
	}

	return ctx.Buf.String(), nil
}

// collectMixinsAndExtends walks the AST to find mixin definitions and extends declarations
func (r *Renderer) collectMixinsAndExtends(nodes []dst.Node) {
	r.collectMixinsAndExtendsWithPrefix(nodes, "")
}

// collectBlockVariables walks the AST to find block variable definitions
func (r *Renderer) collectBlockVariables(nodes []dst.Node) {
	for _, node := range nodes {
		if blockVar, ok := node.(*dst.BlockVariable); ok {
			r.blockVars[blockVar.Name] = blockVar
		}
	}
}

// collectMixinsAndExtendsWithPrefix collects mixins with optional namespace prefix
func (r *Renderer) collectMixinsAndExtendsWithPrefix(nodes []dst.Node, prefix string) {
	for _, node := range nodes {
		if block, ok := node.(*dst.Block); ok {
			// Register blocks that are used as mixins (have selectors starting with . or #)
			// This includes both parametric mixins (.name(params)) and non-parametric (.name())

			isMixin := block.IsMixinFunction

			for _, sel := range block.SelNames {
				if strings.HasPrefix(sel, ".") || strings.HasPrefix(sel, "#") {
					isMixin = true
					break
				}
			}

			// For each selector in this block, check if we need to register it as mixin
			// and also register any nested mixins with namespace prefix
			if isMixin {
				for _, sel := range block.SelNames {
					// Register by simple name
					r.mixins[sel] = append(r.mixins[sel], block)
					// Also register by namespaced name if we have a prefix
					if prefix != "" {
						namespacedName := prefix + " > " + sel
						r.mixins[namespacedName] = append(r.mixins[namespacedName], block)
					}
				}
			}

			// Collect extends from this block
			for _, child := range block.Children {
				if mixin, ok := child.(*dst.MixinCall); ok {
					// Check if this is an &:extend call
					if strings.HasPrefix(mixin.Name, "&:extend") && len(mixin.Args) > 0 {
						// Each arg is a selector being extended
						// (parser splits by comma, but may also parse as single arg with spaces)
						for _, arg := range mixin.Args {
							// Parse the argument in case it contains multiple selectors (zero-alloc)
							extendedSelectors := r.parseExtendSelectors(arg)

							// For each selector being extended, track that this block extends it
							for _, extendedSel := range extendedSelectors {
								for _, sel := range block.SelNames {
									r.extends[extendedSel] = append(r.extends[extendedSel], sel)
								}
							}
						}
					}
				}
			}

			// Recursively collect nested blocks with namespace prefix
			// For nested blocks, the namespace is the current selector
			for _, sel := range block.SelNames {
				newPrefix := prefix
				if newPrefix != "" {
					newPrefix = newPrefix + " " + sel
				} else {
					newPrefix = sel
				}
				r.collectMixinsAndExtendsWithPrefix(block.Children, newPrefix)
			}
		}
	}
}

// renderNodes renders a slice of nodes
func (r *Renderer) renderNodes(parentCtx *NodeContext, parent dst.Node, selName string, nodes []dst.Node) error {
	for _, node := range nodes {
		// For blocks at top level (parent == nil), render them as-is with their selector grouping
		// For nested content, render for each parent selector if we're in nested context
		_, isBlock := node.(*dst.Block)
		if parent == nil && isBlock {
			// Top-level block: render as-is (renderBlock handles multiple selectors)
			if err := r.renderNode(parentCtx, nil, "", node); err != nil {
				return err
			}
		} else {
			// Nested content or non-block nodes
			names := node.Names()

			// Handle nodes without selectors (Decl, Comment, MixinCall at block level)
			if len(names) == 0 {
				if err := r.renderNode(parentCtx, nil, "", node); err != nil {
					return err
				}
				continue
			}

			// Handle nodes with selectors (Block with nested structure)
			for _, name := range names {
				if err := r.renderNode(parentCtx, parent, selector(selName, name), node); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// renderNode renders a single node
func (r *Renderer) renderNode(parentCtx *NodeContext, parent dst.Node, selName string, node dst.Node) error {
	ctx := parentCtx
	if parent != nil {
		ctx = &NodeContext{
			Buf:     ctx.Buf,
			Stack:   ctx.Stack,
			Node:    parent,
			SelName: selName,
			BaseDir: parentCtx.BaseDir,
		}
	}

	switch n := node.(type) {
	case *dst.Comment:
		return r.renderComment(ctx, n)
	case *dst.Decl:
		return r.renderDecl(ctx, n)
	case *dst.Block:
		return r.renderBlock(ctx, n)
	case *dst.MixinCall:
		return r.renderMixinCall(ctx, n)
	case *dst.BlockVariable:
		// BlockVariable nodes are only definitions, they don't render directly
		// They are invoked via @varname() calls
		return nil
	case *dst.Each:
		return r.renderEach(ctx, n)
	case *dst.Import:
		return r.renderImport(ctx, n)
	}
	return nil
}

// renderImport renders an @import statement (for URL imports that pass through)
func (r *Renderer) renderImport(ctx *NodeContext, i *dst.Import) error {
	ctx.Buf.WriteString("@import \"")
	ctx.Buf.WriteString(i.Path)
	ctx.Buf.WriteString("\";\n")
	return nil
}

// renderComment renders a comment node
func (r *Renderer) renderComment(ctx *NodeContext, c *dst.Comment) error {
	// Skip single-line comments (// style) - lessc omits them from CSS output
	if !c.Multiline {
		return nil
	}

	r.writeIndent(ctx.Buf, ctx.Depth()-1)
	ctx.Buf.WriteString("/* ")
	ctx.Buf.WriteString(c.Text)
	ctx.Buf.WriteString(" */")
	ctx.Buf.WriteString("\n")
	return nil
}

// renderDecl renders a declaration node
func (r *Renderer) renderDecl(ctx *NodeContext, d *dst.Decl) error {
	// Check if this is a block variable call (@var();)
	if len(d.Key) > 0 && d.Key[0:1] == "@" && !strings.Contains(d.Key, "{") && strings.TrimSpace(d.Value) == "()" {
		// Block variable invocation - expand the block variable's children
		varName := strings.TrimPrefix(d.Key, "@")

		if blockVar, ok := r.blockVars[varName]; ok {
			// Render the block variable's children
			return r.renderNodes(ctx, nil, ctx.SelName, blockVar.Children)
		}

		// Block variable not found, skip silently
		return nil
	}

	// Check if this is a variable assignment (@var: value;) not an interpolated property
	// Variable assignments have simple identifiers like @varname, not @{...}
	if len(d.Key) > 0 && d.Key[0:1] == "@" && !strings.Contains(d.Key, "{") {
		// Variable assignment - store in stack and don't output to CSS
		varName := strings.TrimPrefix(d.Key, "@")
		value := d.Value

		// Evaluate the value to resolve any functions or expressions
		resolved, err := r.resolver.ResolveValue(ctx.Stack, value)
		if err == nil {
			value = resolved
		}

		ctx.Stack.Set(varName, value)

		return nil
	}

	// Write declaration with proper indentation
	r.writeIndent(ctx.Buf, ctx.Depth()-1)

	// Apply variable interpolation to property name
	key := r.resolver.InterpolateVariables(ctx.Stack, d.Key)
	ctx.Buf.WriteString(key)
	ctx.Buf.WriteString(": ")

	// Skip resolution for CSS3 custom properties (starting with --)
	value := d.Value
	if !strings.HasPrefix(d.Key, "--") {
		resolved, err := r.resolver.ResolveValue(ctx.Stack, value)
		if err != nil {
			return err
		}
		value = resolved
	}

	ctx.Buf.WriteString(value)
	ctx.Buf.WriteString(";\n")

	return nil
}

// renderBlock renders a block node with nested children
func (r *Renderer) renderBlock(ctx *NodeContext, b *dst.Block) error {
	// Skip parametric mixin definitions (they're only invoked, not output)
	if b.IsMixinFunction {
		return nil
	}

	// If block has guard, evaluate it
	satisfied, err := r.evaluateGuard(ctx.Stack, b.Guard)
	if !satisfied || err != nil {
		return nil
	}

	// Handle top-level @media blocks specially
	if len(b.SelNames) > 0 && strings.HasPrefix(b.SelNames[0], "@media") && ctx.SelName == "" {
		return r.renderTopLevelMediaBlock(ctx, b)
	}

	// Compute the full selector names for this block (combining parent context)
	fullSelNames := make([]string, 0, len(b.Names())*2) // preallocate with capacity for names + extends
	for _, name := range b.Names() {
		// Apply variable interpolation to selector names
		interpolatedName := r.resolver.InterpolateVariables(ctx.Stack, name)
		fullSelName := selector(ctx.SelName, interpolatedName)
		fullSelNames = append(fullSelNames, fullSelName)

		// Check if this selector is extended by other selectors
		if extenders, ok := r.extends[fullSelName]; ok {
			for _, extender := range extenders {
				if !contains(fullSelNames, extender) {
					fullSelNames = append(fullSelNames, extender)
				}
			}
		}
	}

	// Separate declarations, nested blocks, and media queries
	decls := make([]dst.Node, 0, len(b.Children))
	nestedBlocks := make([]dst.Node, 0, len(b.Children))
	mediaBlocks := make([]*dst.Block, 0, len(b.Children))
	var realDeclCount int // Count of non-variable declarations

	for _, child := range b.Children {
		if block, isBlock := child.(*dst.Block); isBlock {
			// Check if this is a media query block
			isMedia := len(block.SelNames) > 0 && strings.HasPrefix(block.SelNames[0], "@media")
			if isMedia {
				mediaBlocks = append(mediaBlocks, block)
			} else {
				nestedBlocks = append(nestedBlocks, child)
			}
		} else if mixin, isMixin := child.(*dst.MixinCall); isMixin {
			// Skip &:extend mixin calls - they're already processed during collection
			if strings.HasPrefix(mixin.Name, "&:extend") {
				continue
			}
			decls = append(decls, child)
			realDeclCount++
		} else if decl, isDecl := child.(*dst.Decl); isDecl {
			decls = append(decls, child)
			// Count as real decl if it's not a variable assignment
			// Block variable calls (@var();) are counted as real since they expand to CSS
			if !strings.HasPrefix(decl.Key, "@") || strings.Contains(decl.Key, "{") || (strings.TrimSpace(decl.Value) == "()") {
				realDeclCount++
			}
		} else {
			decls = append(decls, child)
			realDeclCount++
		}
	}

	// Only render the block opening/closing if it has real declarations (not just variables)
	if realDeclCount > 0 {
		r.writeIndent(ctx.Buf, ctx.Depth()-1)
		for i, fullSel := range fullSelNames {
			if i > 0 {
				ctx.Buf.WriteString(",\n")
				r.writeIndent(ctx.Buf, ctx.Depth()-1)
			}
			ctx.Buf.WriteString(fullSel)
		}

		ctx.Buf.WriteString(" {\n")

		// Push new scope for block-level variables (used when rendering declarations)
		ctx.Stack.Push()

		// Render declarations - they will use the increased stack depth for indentation
		if err := r.renderNodes(ctx, b, ctx.SelName, decls); err != nil {
			ctx.Stack.Pop()
			return err
		}

		ctx.Stack.Pop()

		r.writeIndent(ctx.Buf, ctx.Depth()-1)
		ctx.Buf.WriteString("}\n")

		// Render media queries right after the block's declarations
		for _, fullSelName := range fullSelNames {
			if err := r.renderMediaQueriesForSelector(ctx, fullSelName, mediaBlocks); err != nil {
				return err
			}
		}
	}

	// Render nested blocks at parent level with combined selectors
	// When we have multiple parent selectors, we need to collect all nested selectors
	// and group them together with commas
	if len(nestedBlocks) > 0 && len(fullSelNames) > 1 {
		// Collect all nested selectors for grouped output
		allNestedSelectors := make([]string, 0, len(fullSelNames)*len(nestedBlocks))
		for _, parentSelName := range fullSelNames {
			for _, nestedBlock := range nestedBlocks {
				if nestedBlockNames := nestedBlock.Names(); len(nestedBlockNames) > 0 {
					for _, nestedName := range nestedBlockNames {
						interpolatedName := r.resolver.InterpolateVariables(ctx.Stack, nestedName)
						fullNestedSel := selector(parentSelName, interpolatedName)
						allNestedSelectors = append(allNestedSelectors, fullNestedSel)
					}
				}
			}
		}

		// Render grouped nested selectors if we have any
		if len(allNestedSelectors) > 0 {
			r.writeIndent(ctx.Buf, ctx.Depth()-1)
			for i, sel := range allNestedSelectors {
				if i > 0 {
					ctx.Buf.WriteString(",\n")
					r.writeIndent(ctx.Buf, ctx.Depth()-1)
				}
				ctx.Buf.WriteString(sel)
			}
			ctx.Buf.WriteString(" {\n")

			// Push scope and render nested block declarations
			ctx.Stack.Push()
			for _, nestedBlock := range nestedBlocks {
				if blockNode, ok := nestedBlock.(*dst.Block); ok {
					// Render only the declarations (children that aren't blocks)
					for _, child := range blockNode.Children {
						if _, isBlock := child.(*dst.Block); !isBlock {
							if err := r.renderNode(ctx, blockNode, ctx.SelName, child); err != nil {
								ctx.Stack.Pop()
								return err
							}
						}
					}
				}
			}
			ctx.Stack.Pop()

			r.writeIndent(ctx.Buf, ctx.Depth()-1)
			ctx.Buf.WriteString("}\n")
		}
	} else {
		// Single parent selector or no nested blocks: render normally
		for _, parentSelName := range fullSelNames {
			// Create a new context with the combined parent selector
			nestedCtx := &NodeContext{
				Buf:     ctx.Buf,
				Stack:   ctx.Stack,
				Node:    b,
				SelName: parentSelName,
				BaseDir: ctx.BaseDir,
			}

			for _, nestedBlock := range nestedBlocks {
				if err := r.renderNode(nestedCtx, b, parentSelName, nestedBlock); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// evaluateGuard
func (r *Renderer) evaluateGuard(stack *Stack, g *dst.Guard) (bool, error) {
	if !g.Valid() {
		return true, nil
	}

	// Strip outer parentheses from guard condition if present (accounting for nested parens)
	condition := strings.TrimSpace(g.Condition)
	if strings.HasPrefix(condition, "(") && strings.HasSuffix(condition, ")") {
		// Check that the closing paren matches the opening one
		depth := 0
		allWrapped := true
		for i, ch := range condition {
			if ch == '(' {
				depth++
			} else if ch == ')' {
				depth--
				// If we hit zero before the last character, it's not fully wrapped
				if depth == 0 && i < len(condition)-1 {
					allWrapped = false
					break
				}
			}
		}
		if allWrapped && depth == 0 {
			condition = condition[1 : len(condition)-1]
		}
	}

	// Parse the LESS guard condition tokens to prepare for evaluation
	tokens, err := evaluator.Tokenize(condition)
	if err != nil {
		return false, err
	}

	// Build a Go expression from the tokens
	exprParts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		switch t.Type {
		case evaluator.TokenIdent:
			// Variable reference - drop the @ and use the variable name
			exprParts = append(exprParts, strings.TrimPrefix(t.Text, "@"))
		case evaluator.TokenOp:
			// Map = to ==
			if t.Text == "=" {
				exprParts = append(exprParts, "==")
			} else {
				exprParts = append(exprParts, t.Text)
			}
		case evaluator.TokenValue:
			// Try to parse as a number (with or without units), otherwise quote as string
			numVal := parseNumberForGuard(t.Text)
			if numVal != nil {
				exprParts = append(exprParts, fmt.Sprint(numVal))
			} else {
				// It's a non-numeric value, quote it
				exprParts = append(exprParts, fmt.Sprintf("%q", t.Text))
			}
		case evaluator.TokenParen:
			exprParts = append(exprParts, t.Text)
		}
	}

	goExpr := strings.Join(exprParts, " ")

	vars := stack.All()

	// Don't use Eval/EvalBool - they call preprocessExpression which re-parses the expression
	// Instead compile and run directly with expr library, using pre-processed variables
	evalVars := make(map[string]interface{})
	for k, v := range vars {
		// Convert string variables to appropriate types for expr evaluation
		numVal := parseNumberForGuard(v)
		if numVal != nil {
			evalVars[k] = numVal
		} else if v == "true" {
			evalVars[k] = true
		} else if v == "false" {
			evalVars[k] = false
		} else {
			evalVars[k] = v
		}
	}

	program, err := expr.Compile(goExpr, expr.AllowUndefinedVariables())
	if err != nil {
		return false, err
	}

	result, err := expr.Run(program, evalVars)
	if err != nil {
		return false, err
	}

	// Convert result to boolean
	var res bool
	switch v := result.(type) {
	case bool:
		res = v
	case string:
		res = v == "true"
	case float64:
		res = v != 0
	default:
		res = false
	}

	return res, nil
}

// renderMixinCall renders a mixin call by expanding it
func (r *Renderer) renderMixinCall(ctx *NodeContext, m *dst.MixinCall) error {
	// Find the mixin definition
	blocks, ok := r.mixins[m.Name]
	if !ok {
		return nil
	}

	// First pass: find exact arity match (pattern matching by argument count)
	// LESS supports mixin overloading by arity
	var bestMatch *dst.Block
	for _, b := range blocks {
		// Check if this mixin matches the argument count
		if len(b.Params) == len(m.Args) {
			bestMatch = b
			break // Use first exact match
		}
	}

	// If no exact match found, look for best fallback (mixin with fewer params can accept extra args)
	if bestMatch == nil {
		for _, b := range blocks {
			// Use first mixin that has params (or no params if call has no args)
			if len(b.Params) == 0 && len(m.Args) == 0 {
				bestMatch = b
				break
			} else if len(b.Params) > 0 && len(m.Args) > 0 {
				bestMatch = b
				break
			}
		}
	}

	// If we found a matching mixin, render it
	if bestMatch != nil {
		// Set parameters as variables in the current scope (Stack push/pop is handled by renderBlock)
		if len(bestMatch.Params) > 0 && len(m.Args) > 0 {
			for i, param := range bestMatch.Params {
				if i < len(m.Args) {
					// Remove @ from parameter name
					paramName := strings.TrimPrefix(param, "@")
					argValue := m.Args[i]

					// Try to resolve the argument value (in case it contains expressions like @var or operations)
					resolved, err := r.resolver.ResolveValue(ctx.Stack, argValue)
					if err == nil {
						argValue = resolved
					}

					ctx.Stack.Set(paramName, argValue)
				}
			}
		}

		// If block has guard, evaluate it
		satisfied, err := r.evaluateGuard(ctx.Stack, bestMatch.Guard)
		if !satisfied || err != nil {
			return nil
		}

		// Render mixin children
		// Pass parent=nil so blocks within the mixin are rendered at the correct nesting level
		// If we're rendering a top-level mixin call, children should be top-level
		// If we're in a nested context, children should be nested under the current selName
		if err := r.renderNodes(ctx, nil, ctx.SelName, bestMatch.Children); err != nil {
			return err
		}
	}

	return nil
}

// parseExtendSelectors parses a selector string that may contain multiple selectors
// e.g., ".base, .success" or ".base .success" (zero-alloc)
func (r *Renderer) parseExtendSelectors(selString string) []string {
	// First try splitting by comma (explicit list)
	selString = strings.TrimSpace(selString)

	if strings.Contains(selString, ",") {
		strings.SplitCommaNoAlloc(selString, &r.selectorBuf)
		// Make a copy since the buffer will be reused
		result := make([]string, len(r.selectorBuf))
		copy(result, r.selectorBuf)
		return result
	}

	// No commas - return the whole string (might be a single selector or selector with space)
	return []string{selString}
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// writeIndent writes the current indentation
func (r *Renderer) writeIndent(buf *strings.Builder, depth int) {
	for i := 0; i < depth; i++ {
		buf.WriteString("  ")
	}
}

// renderMediaQueriesForSelector renders media query blocks for a specific parent selector
func (r *Renderer) renderMediaQueriesForSelector(ctx *NodeContext, parentSelName string, mediaBlocks []*dst.Block) error {
	for _, mediaBlock := range mediaBlocks {
		if len(mediaBlock.SelNames) == 0 {
			continue
		}

		condition := mediaBlock.SelNames[0] // "@media ..."

		// Write the media query
		r.writeIndent(ctx.Buf, ctx.Depth()-1)
		ctx.Buf.WriteString(condition)
		ctx.Buf.WriteString(" {\n")

		// Push scope for media query content
		ctx.Stack.Push()

		// Write the parent selector within the media query
		r.writeIndent(ctx.Buf, ctx.Depth()-1)
		ctx.Buf.WriteString(parentSelName)
		ctx.Buf.WriteString(" {\n")

		// Render the children
		mediaCtx := &NodeContext{
			Buf:     ctx.Buf,
			Stack:   ctx.Stack,
			Node:    nil,
			SelName: parentSelName,
			BaseDir: ctx.BaseDir,
		}

		// Push another scope for proper indentation
		ctx.Stack.Push()

		for _, child := range mediaBlock.Children {
			if err := r.renderNode(mediaCtx, nil, "", child); err != nil {
				ctx.Stack.Pop()
				ctx.Stack.Pop()
				return err
			}
		}

		ctx.Stack.Pop()

		r.writeIndent(ctx.Buf, ctx.Depth()-1)
		ctx.Buf.WriteString("}\n")

		ctx.Stack.Pop()

		r.writeIndent(ctx.Buf, ctx.Depth()-1)
		ctx.Buf.WriteString("}\n")
	}

	return nil
}

// renderTopLevelMediaBlock renders a top-level @media block (not nested inside another selector)
func (r *Renderer) renderTopLevelMediaBlock(ctx *NodeContext, b *dst.Block) error {
	condition := b.SelNames[0] // "@media ..."

	// Write the media query opening
	ctx.Buf.WriteString(condition)
	ctx.Buf.WriteString(" {\n")

	// Push scope for media query content
	ctx.Stack.Push()

	// Render the children (which are blocks like .container, h1, etc.)
	for _, child := range b.Children {
		if err := r.renderNode(ctx, nil, "", child); err != nil {
			ctx.Stack.Pop()
			return err
		}
	}

	ctx.Stack.Pop()

	ctx.Buf.WriteString("}\n")

	return nil
}

// parseNumberForGuard tries to parse a value as a number, stripping CSS units
func parseNumberForGuard(value string) interface{} {
	value = strings.TrimSpace(value)

	// Try direct float parse
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num
	}

	// Try stripping common CSS units
	units := []string{"px", "em", "rem", "pt", "cm", "mm", "in", "pc", "ex", "ch", "vw", "vh", "vmin", "vmax"}
	for _, unit := range units {
		if strings.HasSuffix(value, unit) {
			numStr := strings.TrimSuffix(value, unit)
			if num, err := strconv.ParseFloat(numStr, 64); err == nil {
				return num
			}
		}
	}

	return nil
}

// renderEach renders an each() loop
// Format: each(range(3), { .col-@{value} { height: (@value * 50px); } });
// Expands to render the children block once for each value in the list
func (r *Renderer) renderEach(ctx *NodeContext, e *dst.Each) error {
	// Evaluate the list expression to get the values
	vars := ctx.Stack.All()
	eval, err := expression.NewEvaluator(vars)
	if err != nil {
		return err
	}

	// Evaluate the list expression
	result, err := eval.Eval(e.ListExpr)
	if err != nil {
		return err
	}

	// Get the list values - range returns "1, 2, 3"
	listStr := result.String()
	splitValues := strings.Split(listStr, ",")
	values := make([]string, len(splitValues))

	// Split by comma to get individual values
	for i, v := range splitValues {
		values[i] = strings.TrimSpace(v)
	}

	// For each value, render the children with @value set
	for _, val := range values {
		// Update the loop variable in the current scope
		// Don't push a new frame - just update the variable
		oldVal, hadOld := ctx.Stack.Get(e.VarName)
		ctx.Stack.Set(e.VarName, val)

		// Render children
		if err := r.renderNodes(ctx, nil, "", e.Children); err != nil {
			return err
		}

		// Restore old value if it existed
		if hadOld {
			ctx.Stack.Set(e.VarName, oldVal)
		}
	}

	return nil
}
