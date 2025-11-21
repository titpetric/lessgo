package dst

import (
	"fmt"
	"strings"

	"github.com/titpetric/lessgo/parser"
)

// Parser converts tokens to a Document DST
type Parser struct {
	tokens []parser.Token
	pos    int
	source string
}

// NewParser creates a new DST parser
func NewParser(tokens []parser.Token, source string) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		source: source,
	}
}

// Parse parses tokens into a Document
func (p *Parser) Parse() (*Document, error) {
	doc := NewDocument()

	for !p.isAtEnd() {
		p.skipWhitespace()

		if p.isAtEnd() {
			break
		}

		node, err := p.parseTopLevelNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			doc.AddNode(node)
		}
	}

	return doc, nil
}

// parseTopLevelNode parses a top-level statement
func (p *Parser) parseTopLevelNode() (*Node, error) {
	tok := p.peek()

	// Variable: @name: value;
	if tok.Type == parser.TokenVariable {
		return p.parseVariable()
	}

	// At-rule: @media, @import, etc
	if tok.Type == parser.TokenAt {
		return p.parseAtRule()
	}

	// Declaration: selector { ... }
	// Try to parse as declaration, skip if no brace found
	node, err := p.parseDeclaration()
	if err != nil {
		return nil, err
	}
	if node != nil {
		return node, nil
	}

	// If we couldn't parse a declaration, skip this statement
	p.skipUntilStatementEnd()
	return nil, nil
}

// parseDeclarationOrNestedNode parses within a block
func (p *Parser) parseDeclarationOrNestedNode() (*Node, error) {
	tok := p.peek()

	// Variable
	if tok.Type == parser.TokenVariable {
		return p.parseVariable()
	}

	// Check if this is a property or declaration
	// Properties have: name : value;
	// Declarations have: selector { ... }
	// Pseudo-selectors: &:hover { ... } - have { but also : so prioritize {

	// Look ahead for : or { within reasonable distance
	savePos := p.pos
	colonPos := -1
	bracePos := -1

	for i := 0; i < 20 && !p.isAtEnd(); i++ {
		tok := p.peek()
		
		if tok.Type == parser.TokenColon && colonPos == -1 {
			colonPos = i
		}
		if tok.Type == parser.TokenLBrace && bracePos == -1 {
			bracePos = i
			// Found opening brace - we can stop here, it's definitely a declaration
			break
		}
		if tok.Type == parser.TokenSemicolon {
			// Found semicolon before any brace - it's a property
			p.pos = savePos
			return p.parseProperty()
		}
		if tok.Type == parser.TokenRBrace {
			// End of block, no property or rule
			p.pos = savePos
			return nil, nil
		}
		p.advance()
	}

	p.pos = savePos

	// Decide based on what we found
	// If we have a brace, treat as declaration (even if we also have colon for pseudo-selector)
	if bracePos != -1 {
		return p.parseDeclaration()
	}
	
	// Otherwise if we have colon, treat as property
	if colonPos != -1 {
		return p.parseProperty()
	}

	// Neither found, skip
	p.skipUntilStatementEnd()
	return nil, nil
}

// parseDeclaration parses selector { ... }
func (p *Parser) parseDeclaration() (*Node, error) {
	node := NewNodeWithPosition(NodeDeclaration, p.position())

	// Collect selector until {
	selectorStart := p.pos
	selector := p.collectUntil(parser.TokenLBrace, parser.TokenEOF)

	if !p.match(parser.TokenLBrace) {
		// No brace found - not a declaration
		return nil, nil
	}

	selectorEnd := p.pos - 1
	rawSelector := p.extractRawSource(selectorStart, selectorEnd)
	if rawSelector != "" {
		node.SelectorsRaw = strings.TrimSpace(rawSelector)
	} else {
		node.SelectorsRaw = strings.TrimSpace(selector)
	}

	// Parse children until }
	for !p.check(parser.TokenRBrace) && !p.isAtEnd() {
		p.skipWhitespace()

		if p.check(parser.TokenRBrace) {
			break
		}

		child, err := p.parseDeclarationOrNestedNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.AddChild(child)
		}
	}

	if !p.match(parser.TokenRBrace) {
		return nil, fmt.Errorf("expected '}' in rule '%s' at %v", node.SelectorsRaw, p.peek())
	}

	return node, nil
}

// parseProperty parses property: value;
func (p *Parser) parseProperty() (*Node, error) {
	node := NewNodeWithPosition(NodeProperty, p.position())

	// Collect property name - might be multi-part like "box-shadow" or pseudo-selectors
	propStart := p.pos
	propName := ""
	for !p.isAtEnd() && p.pos-propStart < 5 {
		if p.peek().Type == parser.TokenColon {
			// Check if next token after colon suggests this is a pseudo-selector instead
			if p.pos+1 < len(p.tokens) {
				nextTok := p.tokens[p.pos+1]
				// If next token is {, it's probably a pseudo-selector like :hover {
				if nextTok.Type == parser.TokenLBrace {
					// This is not a property, it's a pseudo-selector
					p.pos = propStart
					p.skipUntilStatementEnd()
					return nil, nil
				}
			}
			// It's a colon in a property
			break
		}
		if p.peek().Type == parser.TokenLBrace {
			// Found { without :, not a property
			p.pos = propStart
			return nil, nil
		}
		if propName != "" && p.needsSpaceBeforeToken(p.peek()) {
			propName += " "
		}
		propName += p.peek().Value
		p.advance()
	}

	node.Name = strings.TrimSpace(propName)

	if !p.match(parser.TokenColon) {
		// No colon found, not a property
		p.pos = propStart
		p.skipUntilStatementEnd()
		return nil, nil
	}

	// Collect value from tokens until ; or }
	valueStart := p.pos
	for !p.isAtEnd() && p.peek().Type != parser.TokenSemicolon && p.peek().Type != parser.TokenRBrace {
		p.advance()
	}
	valueEnd := p.pos

	// Extract raw source
	if valueEnd > valueStart {
		node.Value = strings.TrimSpace(p.extractRawSource(valueStart, valueEnd))
	}

	p.match(parser.TokenSemicolon)

	return node, nil
}

// parseVariable parses @name: value;
func (p *Parser) parseVariable() (*Node, error) {
	node := NewNodeWithPosition(NodeVariable, p.position())

	varTok := p.advance()
	// Token value is just the name part, add @ prefix
	node.Name = "@" + varTok.Value

	if !p.match(parser.TokenColon) {
		// Not a simple variable declaration - might be interpolation like @{prop}
		// Skip until next statement end
		p.skipUntilStatementEnd()
		return nil, nil
	}

	// Collect value until ; or {
	value := p.collectUntil(parser.TokenSemicolon, parser.TokenLBrace, parser.TokenEOF)
	node.Value = strings.TrimSpace(value)

	p.match(parser.TokenSemicolon)

	return node, nil
}

// parseAtRule parses @media, @import, etc
func (p *Parser) parseAtRule() (*Node, error) {
	node := NewNodeWithPosition(NodeAtRule, p.position())

	nameTok := p.advance()
	node.Name = nameTok.Value

	// Collect parameters until { (or ; for @import) or other patterns
	params := ""
	for !p.isAtEnd() {
		if p.peek().Type == parser.TokenLBrace || p.peek().Type == parser.TokenSemicolon {
			break
		}
		if params != "" {
			params += " "
		}
		params += p.peek().Value
		p.advance()
	}
	node.Value = strings.TrimSpace(params)

	// At-rules like @import end with ;
	if p.match(parser.TokenSemicolon) {
		return node, nil
	}

	if !p.match(parser.TokenLBrace) {
		// Not followed by { or ;, skip the statement
		p.skipUntilStatementEnd()
		return nil, nil
	}

	// Parse rules inside at-rule
	for !p.check(parser.TokenRBrace) && !p.isAtEnd() {
		p.skipWhitespace()
		if p.check(parser.TokenRBrace) {
			break
		}

		child, err := p.parseTopLevelNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.AddChild(child)
		}
	}

	if !p.match(parser.TokenRBrace) {
		return nil, fmt.Errorf("expected '}' in @%s at %v", node.Name, p.peek())
	}

	return node, nil
}

// Helper methods

func (p *Parser) peek() parser.Token {
	if p.isAtEnd() {
		return parser.Token{Type: parser.TokenEOF, Value: ""}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() parser.Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) check(tokenType parser.TokenType) bool {
	return p.peek().Type == tokenType
}

func (p *Parser) match(tokenType parser.TokenType) bool {
	if p.check(tokenType) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) isAtEnd() bool {
	if p.pos >= len(p.tokens) {
		return true
	}
	return p.tokens[p.pos].Type == parser.TokenEOF
}

func (p *Parser) position() Position {
	tok := p.peek()
	return Position{
		Line:   tok.Line,
		Column: tok.Column,
		Offset: tok.Offset,
	}
}

// skipWhitespace skips newlines and semicolons
func (p *Parser) skipWhitespace() {
	for p.pos < len(p.tokens) && (p.peek().Type == parser.TokenNewline || p.peek().Type == parser.TokenSemicolon) {
		p.advance()
	}
}

// skipUntilStatementEnd skips until ; or }
func (p *Parser) skipUntilStatementEnd() {
	for !p.isAtEnd() {
		if p.peek().Type == parser.TokenSemicolon || p.peek().Type == parser.TokenRBrace {
			if p.peek().Type == parser.TokenSemicolon {
				p.advance()
			}
			break
		}
		p.advance()
	}
}

// collectUntilForValue collects tokens for a value
func (p *Parser) collectUntilForValue(stopTypes ...parser.TokenType) string {
	var result string
	for !p.isAtEnd() {
		tok := p.peek()

		// Check if we've hit a stop token
		for _, stopType := range stopTypes {
			if tok.Type == stopType {
				return result
			}
		}

		// Just collect tokens as-is
		if result != "" {
			result += " "
		}
		result += tok.Value
		p.advance()
	}
	return result
}

// extractRawSource extracts raw source between token positions
func (p *Parser) extractRawSource(startPos, endPos int) string {
	if startPos >= len(p.tokens) || endPos > len(p.tokens) {
		return ""
	}

	var start, end int

	if startPos < len(p.tokens) {
		start = p.tokens[startPos].Offset
		// For VARIABLE tokens, the offset points to @ but value doesn't include it
		// We need to account for the actual token length
		tok := p.tokens[startPos]
		if tok.Type == parser.TokenVariable && start < len(p.source) && p.source[start] == '@' {
			// Token offset correctly points to @, but length only covers the name
			// So we're good, just use offset as-is
		}
	}

	if endPos > 0 && endPos <= len(p.tokens) {
		endTok := p.tokens[endPos-1]
		end = endTok.Offset
		// For VARIABLE at endPos, add 1 for the @ plus the name length
		if endTok.Type == parser.TokenVariable && end < len(p.source) && p.source[end] == '@' {
			end += 1 + len(endTok.Value) // @name
		} else {
			end += len(endTok.Value)
		}
	}

	if start < end && end <= len(p.source) {
		return p.source[start:end]
	}

	return ""
}

// collectUntil collects tokens until one of the given types is encountered
func (p *Parser) collectUntil(stopTypes ...parser.TokenType) string {
	var result string
	for !p.isAtEnd() {
		tok := p.peek()

		// Check if we've hit a stop token
		for _, stopType := range stopTypes {
			if tok.Type == stopType {
				return result
			}
		}

		// Add token value with appropriate spacing
		if result != "" && p.needsSpaceBeforeToken(tok) && !strings.HasSuffix(result, " ") {
			result += " "
		}
		// Preserve @ prefix for variable references
		if tok.Type == parser.TokenVariable {
			result += "@" + tok.Value
		} else {
			result += tok.Value
		}
		p.advance()
	}
	return result
}

// needsSpaceBeforeToken determines if space is needed before token
func (p *Parser) needsSpaceBeforeToken(tok parser.Token) bool {
	switch tok.Type {
	case parser.TokenComma, parser.TokenSemicolon, parser.TokenRBrace, parser.TokenRParen, parser.TokenRBracket,
		parser.TokenDot, parser.TokenHash, parser.TokenColon, parser.TokenLBracket:
		return false
	}
	return true
}

// looksLikePropertyName checks if tokens look like a property name
func (p *Parser) looksLikePropertyName(tokens []parser.Token) bool {
	if len(tokens) == 0 {
		return false
	}

	// Single identifier or identifier with hyphens (common CSS property names)
	if len(tokens) == 1 {
		return p.isLikelyPropertyName(tokens[0].Value)
	}

	// Multiple tokens - check if it's an identifier sequence with hyphens/minus signs
	first := tokens[0]
	if first.Type != parser.TokenIdent {
		return false
	}

	if len(tokens) == 2 {
		second := tokens[1]
		// identifier - identifier (like "font-size")
		if second.Type == parser.TokenIdent || second.Type == parser.TokenMinus {
			return true
		}
	}

	return false
}

// isLikelyPropertyName checks if an identifier looks like a CSS property name
func (p *Parser) isLikelyPropertyName(name string) bool {
	// Common CSS property patterns
	commonProps := map[string]bool{
		// Layout
		"margin": true, "padding": true, "border": true, "width": true, "height": true,
		"display": true, "position": true, "left": true, "right": true, "top": true, "bottom": true,
		"float": true, "clear": true, "flex": true, "grid": true, "z-index": true,
		// Typography
		"font": true, "font-size": true, "font-weight": true, "color": true, "text-align": true,
		"line-height": true, "letter-spacing": true, "text-decoration": true,
		// Background/Color
		"background": true, "background-color": true, "opacity": true,
		// Shadows/Effects
		"box-shadow": true, "text-shadow": true, "transform": true,
		// Other common
		"content": true, "cursor": true, "overflow": true, "white-space": true,
	}

	if commonProps[name] {
		return true
	}

	// If it contains a hyphen, it's likely a property
	if strings.Contains(name, "-") {
		return true
	}

	return false
}
