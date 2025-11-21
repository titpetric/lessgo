// dst.go provides a simple Document Structure Tree for LESS
// This is a simpler alternative to the full AST, treating everything as nodes

package parser

import (
	"fmt"
	"strings"
)

// Node represents any element in the LESS document structure
type Node struct {
	Type      string  // "stylesheet", "rule", "declaration", "variable", "atrule", etc
	Name      string  // selector, property name, variable name, etc
	Value     string  // declaration value, variable value, etc
	Children  []*Node // nested rules, declarations, etc
	Parent    *Node
	RawSource string
}

// NewNode creates a new DST node
func NewNode(nodeType string) *Node {
	return &Node{
		Type:     nodeType,
		Children: make([]*Node, 0),
	}
}

// AddChild adds a child node and sets parent
func (n *Node) AddChild(child *Node) {
	if child == nil {
		return
	}
	child.Parent = n
	n.Children = append(n.Children, child)
}

// DSTParser converts tokens to a simple node tree
type DSTParser struct {
	tokens []Token
	pos    int
	source string
}

// NewDSTParser creates a new DST parser
func NewDSTParser(tokens []Token, source string) *DSTParser {
	return &DSTParser{
		tokens: tokens,
		pos:    0,
		source: source,
	}
}

// Parse parses tokens into a document tree
func (p *DSTParser) Parse() (*Node, error) {
	root := NewNode("stylesheet")

	for !p.isAtEnd() {
		// Skip whitespace/newlines/semicolons
		for p.pos < len(p.tokens) && (p.peek().Type == TokenNewline || p.peek().Type == TokenSemicolon) {
			p.advance()
		}

		if p.isAtEnd() {
			break
		}

		node, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if node != nil {
			root.AddChild(node)
		}
	}

	return root, nil
}

// parseStatement parses a top-level statement
func (p *DSTParser) parseStatement() (*Node, error) {
	tok := p.peek()

	// Comment
	if tok.Type == TokenCommentOneline {
		return p.parseCommentOneline()
	}
	if tok.Type == TokenCommentMultline {
		return p.parseCommentMultiline()
	}

	// Variable declaration: @name: value;
	if tok.Type == TokenVariable {
		return p.parseVariable()
	}

	// At-rule: @media, etc
	if isAtRuleKeyword(tok.Value) {
		return p.parseAtRule()
	}

	// Rule: selector { ... }
	return p.parseRule()
}

// parseCommentOneline parses a single-line comment
func (p *DSTParser) parseCommentOneline() (*Node, error) {
	tok := p.advance()
	node := NewNode("comment_oneline")
	node.Value = tok.Value
	return node, nil
}

// parseCommentMultiline parses a multi-line comment
func (p *DSTParser) parseCommentMultiline() (*Node, error) {
	tok := p.advance()
	node := NewNode("comment_multiline")
	node.Value = tok.Value
	return node, nil
}

// parseVariable parses @name: value;
func (p *DSTParser) parseVariable() (*Node, error) {
	varNode := NewNode("variable")
	varNode.Name = p.advance().Value // @name

	if !p.match(TokenColon) {
		return nil, fmt.Errorf("expected ':' after variable at %v", p.peek())
	}

	// Collect value until ; or {
	value := p.collectUntil(TokenSemicolon, TokenLBrace, TokenEOF)
	varNode.Value = strings.TrimSpace(value)
	p.match(TokenSemicolon)

	return varNode, nil
}

// parseRule parses selector { ... }
func (p *DSTParser) parseRule() (*Node, error) {
	// Collect selector tokens until {
	selector := p.collectUntil(TokenLBrace, TokenEOF)
	if !p.match(TokenLBrace) {
		return nil, fmt.Errorf("expected '{{' after selector at %v", p.peek())
	}

	ruleNode := NewNode("rule")
	ruleNode.Name = strings.TrimSpace(selector)

	// Parse children until }
	for !p.check(TokenRBrace) && !p.isAtEnd() {
		// Skip whitespace
		for p.pos < len(p.tokens) && p.peek().Type == TokenNewline {
			p.advance()
		}

		if p.check(TokenRBrace) {
			break
		}

		child, err := p.parseDeclarationOrRule()
		if err != nil {
			return nil, err
		}
		if child != nil {
			ruleNode.AddChild(child)
		}
	}

	if !p.match(TokenRBrace) {
		return nil, fmt.Errorf("expected '}}' at %v", p.peek())
	}

	return ruleNode, nil
}

// parseDeclarationOrRule parses a property: value; or nested rule
func (p *DSTParser) parseDeclarationOrRule() (*Node, error) {
	tok := p.peek()

	// Comment: only render multiline comments
	if tok.Type == TokenCommentMultline {
		return p.parseCommentMultiline()
	}
	// Skip single-line comments
	if tok.Type == TokenCommentOneline {
		p.advance()
		return nil, nil
	}

	// Parent selector & is always a nested rule
	if tok.Type == TokenAmpersand {
		return p.parseRule()
	}

	// Nested rule or at-rule
	if tok.Type == TokenDot || tok.Type == TokenHash || tok.Type == TokenIdent || tok.Type == TokenFunction {
		// Look ahead for : to distinguish property declaration from nested rule
		// Check if the next occurrence of : is a declaration colon (property: value)
		// or a pseudo-selector colon (&:hover, a:hover, etc)
		isDeclaration := false
		checkPos := p.pos
		tokCount := 0
		for checkPos < len(p.tokens) && tokCount < 3 {
			if p.tokens[checkPos].Type == TokenColon {
				isDeclaration = true
				break
			}
			if p.tokens[checkPos].Type == TokenLBrace {
				break
			}
			checkPos++
			tokCount++
		}

		if !isDeclaration {
			return p.parseRule()
		}
	}

	// Variable inside rule
	if tok.Type == TokenVariable {
		return p.parseVariable()
	}

	// Declaration: property: value;
	property := p.advance().Value

	if !p.match(TokenColon) {
		return nil, fmt.Errorf("expected ':' after property '%s' at %v", property, p.peek())
	}

	// Collect value until ; or }
	value := p.collectUntil(TokenSemicolon, TokenRBrace, TokenEOF)
	p.match(TokenSemicolon)

	declNode := NewNode("declaration")
	declNode.Name = strings.TrimSpace(property)
	declNode.Value = strings.TrimSpace(value)

	return declNode, nil
}

// parseAtRule parses @media, @import, etc
func (p *DSTParser) parseAtRule() (*Node, error) {
	nameToken := p.advance()
	atNode := NewNode("atrule")
	atNode.Name = nameToken.Value

	// Collect parameters until {
	params := p.collectUntil(TokenLBrace, TokenEOF)
	atNode.Value = strings.TrimSpace(params)

	if !p.match(TokenLBrace) {
		return nil, fmt.Errorf("expected '{{' in @%s at %v", atNode.Name, p.peek())
	}

	// Parse rules inside at-rule
	for !p.check(TokenRBrace) && !p.isAtEnd() {
		for p.pos < len(p.tokens) && p.peek().Type == TokenNewline {
			p.advance()
		}
		if p.check(TokenRBrace) {
			break
		}

		child, err := p.parseRule()
		if err != nil {
			return nil, err
		}
		if child != nil {
			atNode.AddChild(child)
		}
	}

	if !p.match(TokenRBrace) {
		return nil, fmt.Errorf("expected '}}' in @%s at %v", atNode.Name, p.peek())
	}

	return atNode, nil
}

// Helper methods

func (p *DSTParser) peek() Token {
	if p.isAtEnd() {
		return Token{Type: TokenEOF, Value: ""}
	}
	return p.tokens[p.pos]
}

func (p *DSTParser) advance() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *DSTParser) check(tokenType TokenType) bool {
	return p.peek().Type == tokenType
}

func (p *DSTParser) match(tokenType TokenType) bool {
	if p.check(tokenType) {
		p.advance()
		return true
	}
	return false
}

func (p *DSTParser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.peek().Type == TokenEOF
}

// collectUntil collects tokens until one of the given types is encountered
func (p *DSTParser) collectUntil(stopTypes ...TokenType) string {
	var result string
	for !p.isAtEnd() {
		tok := p.peek()

		// Check if we've hit a stop token
		for _, stopType := range stopTypes {
			if tok.Type == stopType {
				return result
			}
		}

		// Add token value
		if result != "" && needsSpaceBeforeToken(tok) && !strings.HasSuffix(result, " ") {
			result += " "
		}
		result += tok.Value
		p.advance()
	}
	return result
}

func needsSpaceBeforeToken(tok Token) bool {
	switch tok.Type {
	case TokenComma, TokenSemicolon, TokenRBrace, TokenRParen, TokenRBracket:
		return false
	}
	return true
}

func isAtRuleKeyword(value string) bool {
	keywords := map[string]bool{
		"media":     true,
		"import":    true,
		"supports":  true,
		"keyframes": true,
		"charset":   true,
		"font-face": true,
	}
	return keywords[value]
}
