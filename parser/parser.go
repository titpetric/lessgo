package parser

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
)

// Parser parses LESS tokens into an AST
type Parser struct {
	tokens   []Token
	pos      int
	source   string                // Original source for comment extraction
	comments map[int][]CommentInfo // Comments mapped by line number
}

// NewParser creates a new parser
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

// NewParserWithSource creates a new parser with source code for comment preservation
func NewParserWithSource(tokens []Token, source string) *Parser {
	p := &Parser{
		tokens:   tokens,
		pos:      0,
		source:   source,
		comments: ExtractComments(source),
	}
	return p
}

// getCommentsForLine gets all comments associated with the given line
// This function collects comments on or immediately before the given line
func (p *Parser) getCommentsForLine(lineNum int) []*ast.Comment {
	var comments []*ast.Comment

	// Look for comments on the line itself and the previous line
	// (comments are typically on the line before or same line after code)
	for line := lineNum - 5; line <= lineNum; line++ {
		if line >= 0 && line < len(strings.Split(p.source, "\n")) {
			if comms, ok := p.comments[line]; ok {
				for _, comm := range comms {
					// Convert CommentInfo to ast.Comment
					comment := &ast.Comment{
						Text:    comm.Text,
						IsBlock: comm.IsBlock,
					}
					comments = append(comments, comment)
				}
			}
		}
	}

	return comments
}

// getAndClearCommentsBeforeLine extracts comments that appear before a line and removes them from the map
// This prevents the same comment from being attached to multiple statements
// lineNum is 1-indexed (from tokens), but comment map is 0-indexed
func (p *Parser) getAndClearCommentsBeforeLine(lineNum int) []*ast.Comment {
	var comments []*ast.Comment

	// Convert token's 1-indexed line to 0-indexed for comment map
	// Collect comments on lines before this statement
	for line := 0; line < lineNum-1; line++ {
		if comms, ok := p.comments[line]; ok {
			// Only include if no code on this line follows the comment
			for _, comm := range comms {
				comment := &ast.Comment{
					Text:    comm.Text,
					IsBlock: comm.IsBlock,
				}
				comments = append(comments, comment)
			}
			delete(p.comments, line) // Remove to avoid reattaching
		}
	}

	return comments
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

	// Capture leading comments before parsing this statement
	// Get comments on lines before this statement
	leadingComments := p.getAndClearCommentsBeforeLine(tok.Line)

	// At-rule (@import, @media, @keyframes, etc.)
	// These are tokenized as TokenVariable with names like "import", "media", etc.
	if tok.Type == TokenVariable {
		// Check if it's an at-rule keyword
		if isAtRuleKeyword(tok.Value) {
			return p.parseAtRule()
		}
		// Otherwise it's a variable declaration
		varDecl, err := p.parseVariableDeclaration()
		// Attach comments to variable declaration
		if varDecl != nil {
			varDecl.Comments = leadingComments
		}
		return varDecl, err
	}

	// At-rule (in case @ is tokenized separately)
	if tok.Type == TokenAt {
		return p.parseAtRule()
	}

	// Rule (selector + block)
	stmt, err := p.parseRule()
	// Attach comments to rule
	if rule, ok := stmt.(*ast.Rule); ok && rule != nil {
		rule.Comments = leadingComments
	}
	return stmt, err
}

// parseRule parses a CSS rule
func (p *Parser) parseRule() (ast.Statement, error) {
	selector, err := p.parseSelector()
	if err != nil {
		return nil, err
	}

	rule := ast.NewRule(selector)

	// Check for mixin parameters: .mixin(@param1; @param2)
	if p.check(TokenLParen) {
		params, err := p.parseMixinParameters()
		if err != nil {
			return nil, err
		}
		rule.Parameters = params
	}

	// Check for guard condition: when (...) or unless (...)
	if p.checkIdentValue("when") || p.checkIdentValue("unless") {
		isWhen := p.peekIdentValue() == "when"
		p.advance()
		guard, err := p.parseGuard(isWhen)
		if err != nil {
			return nil, err
		}
		rule.Guard = guard
	}

	if !p.match(TokenLBrace) {
		return nil, fmt.Errorf("expected '{' at %v", p.peek())
	}

	// Parse declarations and nested rules
	for !p.check(TokenRBrace) && !p.isAtEnd() {
		if p.match(TokenSemicolon) {
			continue
		}

		// Check for &:extend(...) syntax
		if p.check(TokenAmpersand) {
			savedPos := p.pos
			p.advance() // consume &
			if p.check(TokenColon) {
				p.advance() // consume :
				// The lexer tokenizes 'extend(...)' as FUNCTION token
				if p.check(TokenFunction) && p.peek().Value == "extend" {
					// This is an extend declaration
					p.pos = savedPos
					extend, err := p.parseExtend()
					if err != nil {
						return nil, err
					}
					rule.Extends = append(rule.Extends, extend...)
					p.match(TokenSemicolon)
					continue
				} else if p.check(TokenIdent) && p.peek().Value == "extend" {
					// In case of separate tokenization
					p.advance() // consume extend
					if p.check(TokenLParen) {
						// This is an extend declaration
						p.pos = savedPos
						extend, err := p.parseExtend()
						if err != nil {
							return nil, err
						}
						rule.Extends = append(rule.Extends, extend...)
						p.match(TokenSemicolon)
						continue
					}
				}
			}
			// Not an extend, reset and handle as nested rule
			p.pos = savedPos
		}

		// Try to detect mixin call (.classname() or #namespace.classname())
		if (p.check(TokenDot) || p.check(TokenHash)) && p.isMixinCall() {
			mixin, err := p.parseMixinCall()
			if err != nil {
				return nil, err
			}
			if mixin != nil {
				rule.AddNestedRule(mixin)
			}
			p.match(TokenSemicolon)
		} else if p.check(TokenDot) || p.check(TokenHash) || p.check(TokenLBracket) ||
			p.check(TokenAmpersand) || p.check(TokenGreater) {
			// It's a nested rule
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

		// Collect selector tokens until comma, brace, or LPAREN (for parameters)
		for !p.check(TokenLBrace) && !p.check(TokenComma) && !p.check(TokenLParen) && !p.isAtEnd() {
			if p.check(TokenSemicolon) {
				break
			}

			tok := p.peek()

			// Handle interpolation in selectors
			if tok.Type == TokenInterp {
				// Store interpolation marker and the variable/expression that follows
				p.advance()
				// Peek at the next token(s) to get the variable name
				if p.check(TokenVariable) {
					varTok := p.advance()
					part += "@{" + varTok.Value + "}"
				} else if p.check(TokenIdent) {
					// Could be a property name or keyword in interpolation
					identTok := p.advance()
					part += "@{" + identTok.Value + "}"
				}
				if !p.match(TokenInterpEnd) {
					return ast.Selector{}, fmt.Errorf("expected '}' in selector interpolation at %v", p.peek())
				}
				// Don't add spaces around interpolation - continue directly to next token
				continue
			}

			part += tok.Value

			// Handle whitespace between tokens in selectors
			p.advance()

			// Look ahead for space
			if !p.check(TokenLBrace) && !p.check(TokenComma) && !p.check(TokenLParen) &&
				!p.check(TokenSemicolon) && !p.isAtEnd() {
				nextTok := p.peek()
				// Add space between tokens if needed
				// Skip space before operators like +, -, ~, > and after punctuation
				if tok.Type != TokenGreater && nextTok.Type != TokenGreater &&
					tok.Type != TokenPlus && nextTok.Type != TokenPlus &&
					tok.Type != TokenTilde && nextTok.Type != TokenTilde &&
					tok.Type != TokenMinus && nextTok.Type != TokenMinus &&
					needsSpaceBetween(tok, nextTok) {
					part += " "
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
	property := ""

	// Parse property name (may include interpolation like @{prop})
	// CSS3 custom properties start with -- (two hyphens)
	for {
		if p.check(TokenInterp) {
			// Handle interpolation in property name
			p.advance()
			if p.check(TokenVariable) {
				varTok := p.advance()
				property += "@{" + varTok.Value + "}"
			} else if p.check(TokenIdent) {
				identTok := p.advance()
				property += "@{" + identTok.Value + "}"
			}
			if !p.match(TokenInterpEnd) {
				return nil, fmt.Errorf("expected '}' in property interpolation at %v", p.peek())
			}
		} else if p.check(TokenMinus) {
			// Handle CSS3 custom properties (--name) and property names with hyphens
			property += p.advance().Value
		} else if p.checkIdent() {
			property += p.advance().Value
		} else {
			break
		}

		// Check if next is colon (end of property) or continue with more property name
		if p.check(TokenColon) {
			break
		}
	}

	if property == "" {
		return nil, nil
	}

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

// parseMixinParameters parses mixin parameters: @param1; @param2
func (p *Parser) parseMixinParameters() ([]string, error) {
	if !p.match(TokenLParen) {
		return nil, fmt.Errorf("expected '(' in mixin parameters at %v", p.peek())
	}

	params := []string{}

	for !p.check(TokenRParen) && !p.isAtEnd() {
		if p.check(TokenVariable) {
			paramName := p.advance().Value
			params = append(params, paramName)

			// Skip optional colon and default value
			if p.match(TokenColon) {
				// Parse default value until ; or )
				for !p.check(TokenSemicolon) && !p.check(TokenRParen) && !p.isAtEnd() {
					p.advance()
				}
			}

			// Check for separator (; or ,)
			if p.match(TokenSemicolon) || p.match(TokenComma) {
				// Continue to next parameter
				continue
			}
		} else {
			// Skip non-variable tokens (for malformed input)
			p.advance()
		}
	}

	if !p.match(TokenRParen) {
		return nil, fmt.Errorf("expected ')' in mixin parameters at %v", p.peek())
	}

	return params, nil
}

// isMixinCall checks if the current position starts a mixin call
func (p *Parser) isMixinCall() bool {
	savedPos := p.pos

	// Try to parse as a mixin call
	// Pattern: .classname() or #namespace.classname() or .namespace > .classname()
	for !p.isAtEnd() {
		if p.match(TokenDot) || p.match(TokenHash) {
			// Skip the class/id name (could be IDENT or FUNCTION)
			if p.check(TokenIdent) {
				p.advance()
			} else if p.check(TokenFunction) {
				// mixin() is tokenized as FUNCTION
				p.advance()
				p.pos = savedPos
				return true // Found .functionname() - this is a mixin call
			} else {
				p.pos = savedPos
				return false
			}

			// Check for > (descendent selector)
			if p.check(TokenGreater) {
				p.advance()
				continue
			}

			// Check for ()
			if p.check(TokenLParen) {
				p.pos = savedPos
				return true
			}

			// Check for another . or #
			if p.check(TokenDot) || p.check(TokenHash) {
				continue
			}

			// No () found, not a mixin call
			p.pos = savedPos
			return false
		} else {
			p.pos = savedPos
			return false
		}
	}

	p.pos = savedPos
	return false
}

// parseMixinCall parses a mixin call like .classname() or #namespace.classname()
func (p *Parser) parseMixinCall() (*ast.MixinCall, error) {
	path := []string{}

	// Parse the namespace/path (.classname or #namespace.classname)
	for {
		if p.match(TokenDot) {
			if p.check(TokenIdent) {
				path = append(path, p.advance().Value)
			} else if p.check(TokenFunction) {
				// mixin() is tokenized as FUNCTION, extract just the name
				funcToken := p.advance()
				path = append(path, funcToken.Value)
				// Skip the () which are already consumed as part of FUNCTION token
				// The FUNCTION token only gives us the name
				if !p.match(TokenLParen) {
					return nil, fmt.Errorf("expected '(' after function name in mixin call at %v", p.peek())
				}
				// Parse arguments
				args := []ast.Value{}
				for !p.check(TokenRParen) && !p.isAtEnd() {
					arg, err := p.parseCommaListNoSpaces()
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
					return nil, fmt.Errorf("expected ')' in mixin call at %v", p.peek())
				}
				return &ast.MixinCall{
					Path:      path,
					Arguments: args,
				}, nil
			} else {
				return nil, fmt.Errorf("expected identifier after '.' in mixin call at %v", p.peek())
			}
		} else if p.match(TokenHash) {
			if !p.check(TokenIdent) {
				return nil, fmt.Errorf("expected identifier after '#' in mixin call at %v", p.peek())
			}
			path = append(path, p.advance().Value)
		} else {
			break
		}

		// Check for > (descendent selector)
		if p.match(TokenGreater) {
			// Continue to next segment
			continue
		} else {
			// End of path
			break
		}
	}

	if len(path) == 0 {
		return nil, fmt.Errorf("expected mixin name at %v", p.peek())
	}

	// Expect (
	if !p.match(TokenLParen) {
		return nil, fmt.Errorf("expected '(' in mixin call at %v", p.peek())
	}

	// Parse arguments (optional)
	args := []ast.Value{}
	for !p.check(TokenRParen) && !p.isAtEnd() {
		arg, err := p.parseCommaListNoSpaces()
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

	// Expect )
	if !p.match(TokenRParen) {
		return nil, fmt.Errorf("expected ')' in mixin call at %v", p.peek())
	}

	return &ast.MixinCall{
		Path:      path,
		Arguments: args,
	}, nil
}

// parseValue parses a CSS value (handles operators and comma-separated lists)
func (p *Parser) parseValue() (ast.Value, error) {
	return p.parseCommaListWithSpaces(true)
}

// parseCommaListNoSpaces parses comma-separated values without space-separated values
func (p *Parser) parseCommaListNoSpaces() (ast.Value, error) {
	return p.parseCommaListWithSpaces(false)
}

// parseCommaListWithSpaces parses comma-separated or optionally space-separated values
func (p *Parser) parseCommaListWithSpaces(allowSpaces bool) (ast.Value, error) {
	values := []ast.Value{}
	separator := " " // default to space-separated

	for !p.check(TokenSemicolon) && !p.check(TokenRBrace) && !p.isAtEnd() {
		val, err := p.parseBinaryOp()
		if err != nil {
			return nil, err
		}
		if val == nil {
			break
		}
		values = append(values, val)

		// Check for comma separator - commas are only allowed when allowSpaces=true
		// (in property values, not in function arguments)
		if allowSpaces && p.match(TokenComma) {
			separator = ", " // use comma separator
			continue
		}

		// Check if there's another value following (space-separated)
		// Only for allowSpaces=true (property values, not function arguments)
		if !allowSpaces {
			break
		}

		// Stop if we hit end-of-value markers
		if p.check(TokenSemicolon) || p.check(TokenRBrace) || p.isAtEnd() {
			break
		}

		// Look ahead to see if the next token could start a value
		nextTok := p.peek()

		// Check if this looks like a new property (IDENT followed by COLON)
		// This handles missing semicolons between declarations
		if nextTok.Type == TokenIdent && p.peekAhead(1).Type == TokenColon {
			// This is a new property, stop parsing this value
			break
		}

		// Check for at-rules like @media that should not be part of property values
		if nextTok.Type == TokenVariable && isAtRuleKeyword(nextTok.Value) {
			// This is an at-rule, not a value
			break
		}

		if isValueStart(nextTok.Type) {
			// Continue to parse space-separated values
			separator = " " // ensure we use space separator for space-separated values
			continue
		}
		break
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("expected value at %v", p.peek())
	}

	if len(values) == 1 {
		return values[0], nil
	}

	// Return list with appropriate separator
	return &ast.List{
		Values:    values,
		Separator: separator,
	}, nil
}

// parseCommaList is for backwards compatibility
func (p *Parser) parseCommaList() (ast.Value, error) {
	return p.parseCommaListWithSpaces(true)
}

// canStartValue checks if a token type can start a value
func canStartValue(tokType TokenType) bool {
	switch tokType {
	case TokenString, TokenNumber, TokenColor, TokenVariable,
		TokenInterp, TokenFunction, TokenIdent, TokenMinus, TokenPlus,
		TokenLParen:
		return true
	}
	return false
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
		// Don't treat @media, @import, etc. as variable references in values
		if isAtRuleKeyword(name) {
			return nil, nil
		}
		p.advance()
		return &ast.Variable{Name: name}, nil

	case TokenInterp:
		return p.parseInterpolation()

	case TokenFunction:
		return p.parseFunctionCall()

	case TokenIdent:
		p.advance()
		return &ast.Literal{Type: ast.KeywordLiteral, Value: tok.Value}, nil

	case TokenMinus:
		// Handle CSS3 custom properties (--name) and negative keywords
		// Collect all consecutive minus and ident tokens
		value := ""
		for p.check(TokenMinus) || p.checkIdent() {
			value += p.advance().Value
		}
		return &ast.Literal{Type: ast.KeywordLiteral, Value: value}, nil

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
		// Parse function argument - allows space-separated and comma-separated values
		arg, err := p.parseFunctionArg()
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

// parseFunctionArg parses a function argument, allowing space-separated values
// but stopping at comma or closing parenthesis
func (p *Parser) parseFunctionArg() (ast.Value, error) {
	values := []ast.Value{}

	for !p.check(TokenComma) && !p.check(TokenRParen) && !p.isAtEnd() {
		val, err := p.parseBinaryOp()
		if err != nil {
			return nil, err
		}
		if val == nil {
			break
		}
		values = append(values, val)
		// Debug: uncomment to see what we parsed
		// fmt.Fprintf(os.Stderr, "DEBUG parseFunctionArg: parsed value %v, now at %v\n", val, p.peek())

		// Check if there's another value (space-separated)
		if p.check(TokenComma) || p.check(TokenRParen) || p.isAtEnd() {
			break
		}

		// Look ahead to see if the next token could start a value
		nextTok := p.peek()

		// Check if this looks like the end of the argument
		// (new property with COLON, semicolon, brace, etc)
		if nextTok.Type == TokenIdent && p.peekAhead(1).Type == TokenColon {
			break
		}

		// If the next token can't start a value, stop
		// Special case: TokenMinus only starts a value if followed by a number
		if nextTok.Type == TokenMinus {
			if p.peekAhead(1).Type != TokenNumber {
				break
			}
		} else if !canStartValue(nextTok.Type) {
			break
		}
	}

	if len(values) == 0 {
		return nil, nil
	}

	if len(values) == 1 {
		return values[0], nil
	}

	// Multiple values - return as a space-separated list
	return &ast.List{
		Values:    values,
		Separator: " ",
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
	var name string

	// Handle both TokenAt followed by identifier, and TokenVariable (like "import", "media")
	if p.check(TokenVariable) {
		// @import, @media, etc. are tokenized as TokenVariable
		name = p.advance().Value
	} else if p.match(TokenAt) {
		// Fallback for @ followed by identifier
		if p.checkIdent() {
			name = p.advance().Value
		} else {
			return nil, fmt.Errorf("expected rule name after '@' at %v", p.peek())
		}
	} else {
		return nil, fmt.Errorf("expected '@' or at-rule keyword at %v", p.peek())
	}

	params := ""

	// Collect parameters until { or ;
	for !p.check(TokenLBrace) && !p.check(TokenSemicolon) && !p.isAtEnd() {
		tok := p.advance()
		if params != "" {
			// Add space unless:
			// - current token is ), ], :, ;, . or
			// - previous char is (, [, or .
			lastChar := params[len(params)-1]
			if tok.Value != ")" && tok.Value != "]" && tok.Value != ":" &&
				tok.Value != ";" && tok.Value != "." &&
				lastChar != '(' && lastChar != '[' && lastChar != '.' {
				params += " "
			}
		}
		params += tok.Value
	}

	rule := &ast.AtRule{
		Name:       name,
		Parameters: strings.TrimSpace(params),
	}

	if p.match(TokenLBrace) {
		// Parse block - can contain rules and/or declarations
		var stmts []ast.Statement
		for !p.check(TokenRBrace) && !p.isAtEnd() {
			if p.match(TokenSemicolon) {
				continue
			}

			// Check if this looks like a rule (starts with selector) or a declaration
			// Selectors start with: . # [ & > or identifiers that could be pseudo-selectors
			if p.check(TokenDot) || p.check(TokenHash) || p.check(TokenLBracket) ||
				p.check(TokenAmpersand) || p.check(TokenGreater) {
				// It's a nested rule
				nestedStmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				if nestedStmt != nil {
					stmts = append(stmts, nestedStmt)
				}
			} else if p.check(TokenVariable) || p.check(TokenAt) {
				// Could be a nested at-rule or variable declaration
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				if stmt != nil {
					stmts = append(stmts, stmt)
				}
			} else if p.checkIdent() {
				// Try to parse as a declaration first
				decl, err := p.parseDeclaration()
				if err != nil {
					// Not a declaration, try as a rule (pseudo-selectors, etc.)
					nestedStmt, err := p.parseStatement()
					if err != nil {
						return nil, err
					}
					if nestedStmt != nil {
						stmts = append(stmts, nestedStmt)
					}
				} else if decl != nil {
					stmts = append(stmts, &ast.DeclarationStmt{Declaration: *decl})
					p.match(TokenSemicolon)
				}
			} else {
				// Try to parse as statement (covers other cases)
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				if stmt != nil {
					stmts = append(stmts, stmt)
				}
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

// isAtRuleKeyword checks if a keyword is an at-rule directive
func isAtRuleKeyword(name string) bool {
	switch strings.ToLower(name) {
	case "import", "media", "keyframes", "supports", "namespace",
		"document", "page", "font-face", "charset", "when", "unless":
		return true
	default:
		return false
	}
}

// parseInterpolation parses @{ ... } or #{ ... } interpolation
func (p *Parser) parseInterpolation() (*ast.Interpolation, error) {
	if !p.match(TokenInterp) {
		return nil, fmt.Errorf("expected '@{' or '#{' at %v", p.peek())
	}

	// Parse the expression inside the braces
	expr, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	if !p.match(TokenInterpEnd) {
		return nil, fmt.Errorf("expected '}' to close interpolation at %v", p.peek())
	}

	return &ast.Interpolation{
		Expression: expr,
	}, nil
}

// parseGuard parses a guard condition: when (condition) or unless (condition)
func (p *Parser) parseGuard(isWhen bool) (*ast.Guard, error) {
	guard := &ast.Guard{
		IsWhen:     isWhen,
		Conditions: []*ast.GuardCondition{},
	}

	if !p.match(TokenLParen) {
		return nil, fmt.Errorf("expected '(' after when/unless at %v", p.peek())
	}

	for {
		// Parse left side of comparison - just a simple value
		left, err := p.parseSimpleValue()
		if err != nil {
			return nil, err
		}

		// Parse comparison operator
		var operator string
		if p.check(TokenEq) {
			operator = "="
			p.advance()
		} else if p.check(TokenNe) {
			operator = "!="
			p.advance()
		} else if p.check(TokenLt) {
			operator = "<"
			p.advance()
		} else if p.check(TokenLe) {
			operator = "<="
			p.advance()
		} else if p.check(TokenGt) {
			operator = ">"
			p.advance()
		} else if p.check(TokenGe) {
			operator = ">="
			p.advance()
		} else {
			return nil, fmt.Errorf("expected comparison operator in guard at %v", p.peek())
		}

		// Parse right side of comparison - just a simple value
		right, err := p.parseSimpleValue()
		if err != nil {
			return nil, err
		}

		guard.Conditions = append(guard.Conditions, &ast.GuardCondition{
			Left:     left,
			Operator: operator,
			Right:    right,
		})

		// Check for more conditions (and/or)
		if p.checkIdentValue("and") {
			p.advance()
			continue
		} else if p.checkIdentValue("or") {
			p.advance()
			continue
		} else {
			break
		}
	}

	if !p.match(TokenRParen) {
		return nil, fmt.Errorf("expected ')' at end of guard at %v", p.peek())
	}

	return guard, nil
}

// parseExtend parses &:extend(.selector) or &:extend(.selector1, .selector2) syntax
// Returns a slice of Extend nodes, one for each selector in the extend list
func (p *Parser) parseExtend() ([]ast.Extend, error) {
	if !p.match(TokenAmpersand) {
		return nil, fmt.Errorf("expected '&' at start of extend at %v", p.peek())
	}

	if !p.match(TokenColon) {
		return nil, fmt.Errorf("expected ':' after '&' in extend at %v", p.peek())
	}

	// Handle FUNCTION token (lexer produces this for "extend(...)")
	if p.check(TokenFunction) && p.peek().Value == "extend" {
		// The FUNCTION token just contains "extend", the ( is a separate token
		p.advance() // consume 'extend' token
	} else if p.check(TokenIdent) && p.peek().Value == "extend" {
		// Alternative: separate IDENT and LPAREN tokens
		p.advance() // consume 'extend'
	} else {
		return nil, fmt.Errorf("expected 'extend' keyword at %v", p.peek())
	}

	// Now expect opening paren
	if !p.match(TokenLParen) {
		return nil, fmt.Errorf("expected '(' in extend at %v", p.peek())
	}

	extends := []ast.Extend{}

	// Parse selectors (can be comma-separated)
	for !p.check(TokenRParen) && !p.isAtEnd() {
		selector := ""

		// Parse selector tokens until comma or )
		for !p.check(TokenComma) && !p.check(TokenRParen) && !p.isAtEnd() {
			tok := p.peek()
			selector += tok.Value
			p.advance()
			// Add space if needed
			if !p.check(TokenComma) && !p.check(TokenRParen) && !p.isAtEnd() {
				nextTok := p.peek()
				if tok.Type != TokenGreater && nextTok.Type != TokenGreater &&
					needsSpaceBetween(tok, nextTok) {
					selector += " "
				}
			}
		}

		selector = strings.TrimSpace(selector)

		// Check for 'all' keyword
		all := false
		if strings.HasSuffix(selector, " all") {
			selector = strings.TrimSpace(strings.TrimSuffix(selector, "all"))
			all = true
		}

		if selector != "" {
			extends = append(extends, ast.Extend{
				Selector: selector,
				All:      all,
			})
		}

		if !p.match(TokenComma) {
			break
		}
	}

	if !p.match(TokenRParen) {
		return nil, fmt.Errorf("expected ')' in extend at %v", p.peek())
	}

	return extends, nil
}

// Helper methods

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) peekAhead(n int) Token {
	if p.pos+n < len(p.tokens) {
		return p.tokens[p.pos+n]
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

// checkIdentValue checks if current token is a specific identifier
func (p *Parser) checkIdentValue(value string) bool {
	tok := p.peek()
	return (tok.Type == TokenIdent || tok.Type == TokenProperty) && strings.ToLower(tok.Value) == strings.ToLower(value)
}

// peekIdentValue returns the value of current token if it's an identifier
func (p *Parser) peekIdentValue() string {
	tok := p.peek()
	if tok.Type == TokenIdent || tok.Type == TokenProperty {
		return tok.Value
	}
	return ""
}

// isValueStart checks if a token type can start a value
// isValueStartWithContext checks if a token can start a value
// Special handling needed to avoid treating @media, @import etc. as values
func isValueStartWithContext(tt TokenType, nextValue string) bool {
	switch tt {
	case TokenString, TokenNumber, TokenColor,
		TokenFunction, TokenIdent, TokenLParen, TokenInterp:
		return true
	case TokenVariable:
		// @variable is a value, but @media, @import, etc. are at-rules
		// We need to check if this is an at-rule keyword
		if isAtRuleKeyword(nextValue) {
			return false // It's an at-rule, not a value
		}
		return true // It's a variable reference
	default:
		return false
	}
}

// Fallback for cases where we don't have context
func isValueStart(tt TokenType) bool {
	switch tt {
	case TokenString, TokenNumber, TokenColor, TokenVariable,
		TokenFunction, TokenIdent, TokenLParen, TokenInterp:
		return true
	default:
		return false
	}
}

func (p *Parser) checkOperator() bool {
	return p.check(TokenPlus) || p.check(TokenMinus) ||
		p.check(TokenStar) || p.check(TokenSlash) ||
		p.check(TokenPercent) || p.check(TokenEq) ||
		p.check(TokenNe) || p.check(TokenLt) ||
		p.check(TokenLe) || p.check(TokenGt) ||
		p.check(TokenGe) || p.check(TokenGreater)
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
