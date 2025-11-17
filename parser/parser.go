package parser

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
)

// Parser parses LESS tokens into an AST
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new parser
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

// Parse parses the tokens into an AST
func (p *Parser) Parse() (*ast.Stylesheet, error) {
	stylesheet := ast.NewStylesheet()

	for !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stylesheet.AddRule(stmt)
		}
	}

	return stylesheet, nil
}

// parseStatement parses a top-level statement
func (p *Parser) parseStatement() (ast.Statement, error) {
	// Skip any newlines
	for p.match(TokenSemicolon) {
		// empty statement
	}

	if p.isAtEnd() {
		return nil, nil
	}

	tok := p.peek()

	// Variable declaration
	if tok.Type == TokenVariable {
		return p.parseVariableDeclaration()
	}

	// At-rule
	if tok.Type == TokenAt {
		return p.parseAtRule()
	}

	// Rule (selector + block)
	return p.parseRule()
}

// parseRule parses a CSS rule
func (p *Parser) parseRule() (ast.Statement, error) {
	selector, err := p.parseSelector()
	if err != nil {
		return nil, err
	}

	rule := ast.NewRule(selector)

	if !p.match(TokenLBrace) {
		return nil, fmt.Errorf("expected '{' at %v", p.peek())
	}

	// Parse declarations and nested rules
	for !p.check(TokenRBrace) && !p.isAtEnd() {
		if p.match(TokenSemicolon) {
			continue
		}

		// Check if it's a nested rule
		if p.check(TokenDot) || p.check(TokenHash) || p.check(TokenLBracket) ||
			p.check(TokenAmpersand) || p.check(TokenGreater) {
			nestedStmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if nestedStmt != nil {
				rule.AddNestedRule(nestedStmt)
			}
		} else if p.check(TokenVariable) || p.check(TokenAt) {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				rule.AddNestedRule(stmt)
			}
		} else {
			// Parse declaration
			decl, err := p.parseDeclaration()
			if err != nil {
				return nil, err
			}
			if decl != nil {
				rule.AddDeclaration(*decl)
			}
			p.match(TokenSemicolon)
		}
	}

	if !p.match(TokenRBrace) {
		return nil, fmt.Errorf("expected '}' at %v", p.peek())
	}

	return rule, nil
}

// parseSelector parses a CSS selector
func (p *Parser) parseSelector() (ast.Selector, error) {
	parts := []string{}

	for {
		part := ""

		// Collect selector tokens until comma or brace
		for !p.check(TokenLBrace) && !p.check(TokenComma) && !p.isAtEnd() {
			if p.check(TokenSemicolon) {
				break
			}

			tok := p.peek()
			part += tok.Value

			// Handle whitespace between tokens in selectors
			p.advance()

			// Look ahead for space
			if !p.check(TokenLBrace) && !p.check(TokenComma) &&
				!p.check(TokenSemicolon) && !p.isAtEnd() {
				nextTok := p.peek()
				// Add space between tokens if needed
				if tok.Type != TokenGreater && nextTok.Type != TokenGreater &&
					tok.Type != TokenPlus && nextTok.Type != TokenPlus &&
					tok.Type != TokenTilde && nextTok.Type != TokenTilde {
					if needsSpaceBetween(tok, nextTok) {
						part += " "
					}
				}
			}
		}

		if part != "" {
			parts = append(parts, strings.TrimSpace(part))
		}

		if !p.match(TokenComma) {
			break
		}
	}

	if len(parts) == 0 {
		return ast.Selector{}, fmt.Errorf("empty selector at %v", p.peek())
	}

	return ast.Selector{Parts: parts}, nil
}

// parseDeclaration parses a CSS declaration
func (p *Parser) parseDeclaration() (*ast.Declaration, error) {
	if !p.checkIdent() {
		return nil, nil
	}

	property := p.advance().Value

	if !p.match(TokenColon) {
		return nil, fmt.Errorf("expected ':' after property at %v", p.peek())
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	return &ast.Declaration{
		Property: property,
		Value:    value,
	}, nil
}

// parseValue parses a CSS value (handles operators and comma-separated lists)
func (p *Parser) parseValue() (ast.Value, error) {
	return p.parseCommaList()
}

// parseCommaList parses comma-separated values
func (p *Parser) parseCommaList() (ast.Value, error) {
	values := []ast.Value{}

	for !p.check(TokenSemicolon) && !p.check(TokenRBrace) && !p.isAtEnd() {
		val, err := p.parseBinaryOp()
		if err != nil {
			return nil, err
		}
		if val == nil {
			break
		}
		values = append(values, val)

		if !p.match(TokenComma) {
			break
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("expected value at %v", p.peek())
	}

	if len(values) == 1 {
		return values[0], nil
	}

	return &ast.List{
		Values:    values,
		Separator: ", ",
	}, nil
}

// parseBinaryOp parses binary operations (+ - * /)
func (p *Parser) parseBinaryOp() (ast.Value, error) {
	left, err := p.parseSimpleValue()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, nil
	}

	// Check for operators
	for p.checkOperator() && !p.check(TokenComma) {
		opToken := p.advance()
		operator := opToken.Value

		right, err := p.parseSimpleValue()
		if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, fmt.Errorf("expected value after operator %s at %v", operator, p.peek())
		}

		left = &ast.BinaryOp{
			Left:     left,
			Operator: operator,
			Right:    right,
		}
	}

	return left, nil
}

// parseSimpleValue parses a single value token
func (p *Parser) parseSimpleValue() (ast.Value, error) {
	tok := p.peek()

	if tok.Type == TokenEOF || tok.Type == TokenSemicolon || tok.Type == TokenRBrace {
		return nil, nil
	}

	switch tok.Type {
	case TokenString:
		p.advance()
		return &ast.Literal{Type: ast.StringLiteral, Value: tok.Value}, nil

	case TokenNumber:
		p.advance()
		return &ast.Literal{Type: ast.UnitLiteral, Value: tok.Value}, nil

	case TokenColor:
		p.advance()
		return &ast.Literal{Type: ast.ColorLiteral, Value: tok.Value}, nil

	case TokenVariable:
		name := tok.Value
		p.advance()
		return &ast.Variable{Name: name}, nil

	case TokenFunction:
		return p.parseFunctionCall()

	case TokenIdent:
		p.advance()
		return &ast.Literal{Type: ast.KeywordLiteral, Value: tok.Value}, nil

	case TokenOperator:
		// Binary operators will be handled at a higher level
		return nil, nil

	default:
		if p.checkOperator() {
			// Return nil to signal end of value
			return nil, nil
		}
		p.advance()
		return &ast.Literal{Type: ast.KeywordLiteral, Value: tok.Value}, nil
	}
}

// parseFunctionCall parses a function call
func (p *Parser) parseFunctionCall() (*ast.FunctionCall, error) {
	name := p.advance().Value

	if !p.match(TokenLParen) {
		return nil, fmt.Errorf("expected '(' after function name at %v", p.peek())
	}

	args := []ast.Value{}

	for !p.check(TokenRParen) && !p.isAtEnd() {
		arg, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if arg != nil {
			args = append(args, arg)
		}

		if !p.match(TokenComma) {
			break
		}
	}

	if !p.match(TokenRParen) {
		return nil, fmt.Errorf("expected ')' at %v", p.peek())
	}

	return &ast.FunctionCall{
		Name:      name,
		Arguments: args,
	}, nil
}

// parseVariableDeclaration parses a variable declaration
func (p *Parser) parseVariableDeclaration() (*ast.VariableDeclaration, error) {
	name := p.advance().Value

	if !p.match(TokenColon) {
		return nil, fmt.Errorf("expected ':' after variable name at %v", p.peek())
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	p.match(TokenSemicolon)

	return &ast.VariableDeclaration{
		Name:  name,
		Value: value,
	}, nil
}

// parseAtRule parses an at-rule (@media, @import, etc.)
func (p *Parser) parseAtRule() (*ast.AtRule, error) {
	if !p.match(TokenAt) {
		return nil, fmt.Errorf("expected '@' at %v", p.peek())
	}

	name := p.advance().Value
	params := ""

	// Collect parameters until { or ;
	for !p.check(TokenLBrace) && !p.check(TokenSemicolon) && !p.isAtEnd() {
		params += p.advance().Value
	}

	rule := &ast.AtRule{
		Name:       name,
		Parameters: strings.TrimSpace(params),
	}

	if p.match(TokenLBrace) {
		// Parse block
		var stmts []ast.Statement
		for !p.check(TokenRBrace) && !p.isAtEnd() {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		}
		if !p.match(TokenRBrace) {
			return nil, fmt.Errorf("expected '}' at %v", p.peek())
		}
		rule.Block = stmts
	} else if p.match(TokenSemicolon) {
		// Import or similar - no block
	}

	return rule, nil
}

// Helper methods

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) advance() Token {
	tok := p.peek()
	if !p.isAtEnd() {
		p.pos++
	}
	return tok
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) checkIdent() bool {
	return p.check(TokenIdent) || p.check(TokenProperty)
}

func (p *Parser) checkOperator() bool {
	return p.check(TokenPlus) || p.check(TokenMinus) ||
		p.check(TokenStar) || p.check(TokenSlash) ||
		p.check(TokenPercent) || p.check(TokenEq) ||
		p.check(TokenNe) || p.check(TokenLt) ||
		p.check(TokenLe) || p.check(TokenGt) ||
		p.check(TokenGe)
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.peek().Type == TokenEOF
}

// needsSpaceBetween determines if a space is needed between two tokens
func needsSpaceBetween(prev, next Token) bool {
	// These token types don't need spaces
	noSpace := map[TokenType]bool{
		TokenDot:       true,
		TokenHash:      true,
		TokenLBracket:  true,
		TokenLParen:    true,
		TokenColon:     true,
		TokenSemicolon: true,
		TokenComma:     true,
		TokenGreater:   true,
		TokenPlus:      true,
		TokenTilde:     true,
	}

	if noSpace[next.Type] || noSpace[prev.Type] {
		return false
	}

	return true
}
