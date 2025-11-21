package parser

// TokenType represents the type of a lexical token
type TokenType string

const (
	// Special tokens
	TokenEOF     TokenType = "EOF"
	TokenError   TokenType = "ERROR"
	TokenNewline TokenType = "NEWLINE"

	// Literals
	TokenString   TokenType = "STRING"
	TokenNumber   TokenType = "NUMBER"
	TokenColor    TokenType = "COLOR"
	TokenKeyword  TokenType = "KEYWORD"
	TokenIdent    TokenType = "IDENT"
	TokenSelector TokenType = "SELECTOR"
	TokenProperty TokenType = "PROPERTY"

	// Variables and functions
	TokenVariable TokenType = "VARIABLE"
	TokenFunction TokenType = "FUNCTION"

	// Operators
	TokenPlus     TokenType = "PLUS"
	TokenMinus    TokenType = "MINUS"
	TokenStar     TokenType = "STAR"
	TokenSlash    TokenType = "SLASH"
	TokenPercent  TokenType = "PERCENT"
	TokenEq       TokenType = "EQ"
	TokenNe       TokenType = "NE"
	TokenLt       TokenType = "LT"
	TokenLe       TokenType = "LE"
	TokenGt       TokenType = "GT"
	TokenGe       TokenType = "GE"
	TokenAnd      TokenType = "AND"
	TokenOr       TokenType = "OR"
	TokenNot      TokenType = "NOT"
	TokenOperator TokenType = "OPERATOR"

	// Delimiters
	TokenLBrace    TokenType = "LBRACE"
	TokenRBrace    TokenType = "RBRACE"
	TokenLParen    TokenType = "LPAREN"
	TokenRParen    TokenType = "RPAREN"
	TokenLBracket  TokenType = "LBRACKET"
	TokenRBracket  TokenType = "RBRACKET"
	TokenColon     TokenType = "COLON"
	TokenSemicolon TokenType = "SEMICOLON"
	TokenComma     TokenType = "COMMA"
	TokenDot       TokenType = "DOT"
	TokenHash      TokenType = "HASH"
	TokenAt        TokenType = "AT"
	TokenTilde     TokenType = "TILDE"
	TokenGreater   TokenType = "GREATER"
	TokenPlus2     TokenType = "PLUS2" // ++ for adjacent sibling combinator (>)

	// Special
	TokenAmpersand TokenType = "AMPERSAND"
	TokenInterp    TokenType = "INTERP"    // #{ or @{
	TokenInterpEnd TokenType = "INTERPEND" // }
	TokenEscape    TokenType = "ESCAPE"    // ~ escape prefix

	// Comments
	TokenCommentOneline  TokenType = "COMMENT_ONELINE"
	TokenCommentMultline TokenType = "COMMENT_MULTILINE"
)

// Token represents a lexical token
type Token struct {
	Type      TokenType
	Value     string
	Line      int
	Column    int
	Offset    int
	QuoteChar string // For TokenString: " or ' (empty for other types)
}

// Lexer tokenizes LESS source code
type Lexer struct {
	input       string
	pos         int // current position
	line        int // current line
	column      int // current column
	start       int // start of current token
	width       int // width of last rune read
	tokens      []Token
	interpDepth int // tracks nesting depth of interpolation @{...}
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 0,
		start:  0,
		width:  0,
	}
}

// Tokenize returns all tokens from the input
func (l *Lexer) Tokenize() []Token {
	for {
		tok := l.nextToken()
		l.tokens = append(l.tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return l.tokens
}

// nextToken returns the next token from the input
func (l *Lexer) nextToken() Token {
	l.skipWhitespaceAndComments()

	if l.pos >= len(l.input) {
		return l.makeToken(TokenEOF, "")
	}

	ch := l.peek()

	// Check for comments first
	if ch == '/' && l.peekAhead(1) == '/' {
		return l.readLineComment()
	}
	if ch == '/' && l.peekAhead(1) == '*' {
		return l.readBlockComment()
	}

	// Single character tokens
	switch ch {
	case '{':
		l.advance()
		return l.makeToken(TokenLBrace, "{")
	case '}':
		l.advance()
		// Check if we're inside interpolation
		if l.interpDepth > 0 {
			l.interpDepth--
			return l.makeToken(TokenInterpEnd, "}")
		}
		return l.makeToken(TokenRBrace, "}")
	case '(':
		l.advance()
		return l.makeToken(TokenLParen, "(")
	case ')':
		l.advance()
		return l.makeToken(TokenRParen, ")")
	case '[':
		l.advance()
		return l.makeToken(TokenLBracket, "[")
	case ']':
		l.advance()
		return l.makeToken(TokenRBracket, "]")
	case ':':
		l.advance()
		return l.makeToken(TokenColon, ":")
	case ';':
		l.advance()
		return l.makeToken(TokenSemicolon, ";")
	case ',':
		l.advance()
		return l.makeToken(TokenComma, ",")
	case '.':
		l.advance()
		// Check for .. or ... (selectors can have these)
		if l.peek() == '.' {
			l.advance()
		}
		return l.makeToken(TokenDot, ".")
	case '#':
		// Check for #{ for interpolation
		if l.peekAhead(1) == '{' {
			l.advance()
			l.advance()
			return l.makeToken(TokenInterp, "#{")
		}
		// Check for color literal (#fff, #ffffff)
		if isHexDigit(l.peekAhead(1)) {
			return l.readColor()
		}
		l.advance()
		return l.makeToken(TokenHash, "#")
	case '@':
		// Check for @{ for interpolation
		if l.peekAhead(1) == '{' {
			l.advance()
			l.advance()
			l.interpDepth++
			return l.makeToken(TokenInterp, "@{")
		}
		// Check for variable (@var)
		if isLetter(l.peekAhead(1)) || l.peekAhead(1) == '_' {
			return l.readVariable()
		}
		l.advance()
		return l.makeToken(TokenAt, "@")
	case '~':
		l.advance()
		return l.makeToken(TokenTilde, "~")
	case '&':
		l.advance()
		return l.makeToken(TokenAmpersand, "&")
	case '+':
		l.advance()
		if l.peek() == '+' {
			l.advance()
			return l.makeToken(TokenPlus2, "++")
		}
		return l.makeToken(TokenPlus, "+")
	case '-':
		// Could be minus operator or start of number
		if isDigit(l.peekAhead(1)) {
			l.advance() // consume the '-'
			return l.readNumber(true)
		}
		l.advance()
		return l.makeToken(TokenMinus, "-")
	case '*':
		l.advance()
		return l.makeToken(TokenStar, "*")
	case '/':
		l.advance()
		return l.makeToken(TokenSlash, "/")
	case '%':
		// Check if this is a % format function (e.g., %("string", args))
		if l.peekAhead(1) == '(' {
			l.advance() // consume '%'
			return l.makeToken(TokenFunction, "%")
		}
		l.advance()
		return l.makeToken(TokenPercent, "%")
	case '>':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenGe, ">=")
		}
		return l.makeToken(TokenGreater, ">")
	case '<':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenLe, "<=")
		}
		return l.makeToken(TokenLt, "<")
	case '=':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenEq, "==")
		}
		return l.makeToken(TokenEq, "=")
	case '!':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return l.makeToken(TokenNe, "!=")
		}
		return l.makeToken(TokenNot, "!")
	case '"', '\'':
		return l.readString()
	}

	// Numbers
	if isDigit(ch) {
		return l.readNumber(false)
	}

	// Identifiers and keywords
	if isLetter(ch) || ch == '_' {
		return l.readIdentifier()
	}

	// Unknown character
	l.advance()
	return l.makeToken(TokenError, string(ch))
}

// readString reads a string literal
func (l *Lexer) readString() Token {
	quote := l.peek()
	l.advance() // skip opening quote

	startLine := l.line
	startCol := l.column

	value := ""
	for l.pos < len(l.input) && l.peek() != quote {
		if l.peek() == '\\' {
			l.advance()
			if l.pos < len(l.input) {
				// Handle escape sequences
				switch l.peek() {
				case 'n':
					value += "\n"
				case 't':
					value += "\t"
				case 'r':
					value += "\r"
				case '\\':
					value += "\\"
				case '"':
					value += "\""
				case '\'':
					value += "'"
				default:
					// For unknown escapes, include the backslash
					value += "\\" + string(l.peek())
				}
				l.advance()
			}
		} else {
			if l.peek() == '\n' {
				l.line++
				l.column = 0
			}
			value += string(l.peek())
			l.advance()
		}
	}

	if l.pos < len(l.input) {
		l.advance() // skip closing quote
	}

	return Token{
		Type:      TokenString,
		Value:     value,
		Line:      startLine,
		Column:    startCol,
		Offset:    l.start,
		QuoteChar: string(quote),
	}
}

// readNumber reads a number (integer or float) with optional unit
// hasMinusPrefix: true if the '-' has already been consumed before calling this
func (l *Lexer) readNumber(hasMinusPrefix bool) Token {
	var startPos, startCol int
	if hasMinusPrefix {
		startPos = l.pos - 1 // Account for the '-' that was already consumed
		startCol = l.column - 1
	} else {
		startPos = l.pos
		startCol = l.column
	}

	// Read digits before decimal
	for isDigit(l.peek()) {
		l.advance()
	}

	// Read decimal part
	if l.peek() == '.' && isDigit(l.peekAhead(1)) {
		l.advance() // skip dot
		for isDigit(l.peek()) {
			l.advance()
		}
	}

	// Read unit (px, em, %, etc.)
	for isLetter(l.peek()) || l.peek() == '%' {
		l.advance()
	}

	value := l.input[startPos:l.pos]

	return Token{
		Type:   TokenNumber,
		Value:  value,
		Line:   l.line,
		Column: startCol,
		Offset: l.start,
	}
}

// readColor reads a color literal (#fff or #ffffff)
func (l *Lexer) readColor() Token {
	startCol := l.column
	l.advance() // skip #

	value := "#"
	hexCount := 0
	for isHexDigit(l.peek()) && hexCount < 8 {
		value += string(l.peek())
		hexCount++
		l.advance()
	}

	// Validate hex color (3, 4, 6, or 8 digits)
	if hexCount == 3 || hexCount == 4 || hexCount == 6 || hexCount == 8 {
		return Token{
			Type:   TokenColor,
			Value:  value,
			Line:   l.line,
			Column: startCol,
			Offset: l.start,
		}
	}

	// Not a valid color, treat as hash
	return Token{
		Type:   TokenHash,
		Value:  value,
		Line:   l.line,
		Column: startCol,
		Offset: l.start,
	}
}

// readVariable reads a variable reference (@variable)
func (l *Lexer) readVariable() Token {
	startCol := l.column
	l.advance() // skip @

	name := ""
	for isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' || l.peek() == '-' {
		name += string(l.peek())
		l.advance()
	}

	return Token{
		Type:   TokenVariable,
		Value:  name,
		Line:   l.line,
		Column: startCol,
		Offset: l.start,
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() Token {
	startCol := l.column
	startPos := l.pos

	for isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' || l.peek() == '-' {
		l.advance()
	}

	value := l.input[startPos:l.pos]

	// Check for keywords
	tokenType := TokenIdent

	// Map specific keywords
	keywords := map[string]TokenType{
		"and": TokenAnd,
		"or":  TokenOr,
		"not": TokenNot,
	}

	if tt, ok := keywords[value]; ok {
		tokenType = tt
	}

	// Check if followed by ( for function
	if l.peek() == '(' {
		tokenType = TokenFunction
	}

	return Token{
		Type:   tokenType,
		Value:  value,
		Line:   l.line,
		Column: startCol,
		Offset: l.start,
	}
}

// readLineComment reads a single-line comment (//)
func (l *Lexer) readLineComment() Token {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance() // skip first /
	l.advance() // skip second /

	value := "//"
	for l.pos < len(l.input) && l.peek() != '\n' {
		value += string(l.peek())
		l.advance()
	}

	return Token{
		Type:   TokenCommentOneline,
		Value:  value,
		Line:   startLine,
		Column: startCol,
		Offset: startPos,
	}
}

// readBlockComment reads a multi-line comment (/* */)
func (l *Lexer) readBlockComment() Token {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance() // skip /
	l.advance() // skip *

	value := "/*"
	for l.pos+1 < len(l.input) {
		if l.input[l.pos] == '*' && l.input[l.pos+1] == '/' {
			value += "*/"
			l.advance()
			l.advance()
			break
		}
		if l.peek() == '\n' {
			l.line++
			l.column = 0
		} else {
			l.column++
		}
		value += string(l.peek())
		l.pos++
	}

	return Token{
		Type:   TokenCommentMultline,
		Value:  value,
		Line:   startLine,
		Column: startCol,
		Offset: startPos,
	}
}

// skipWhitespaceAndComments skips over whitespace (not comments anymore)
func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.input) {
		ch := l.peek()

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			if ch == '\n' {
				l.line++
				l.column = 0
			} else {
				l.column++
			}
			l.pos++
			continue
		}

		break
	}

	l.start = l.pos
}

// peek returns the current rune without advancing
func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

// peekAhead returns the rune at offset n ahead
func (l *Lexer) peekAhead(n int) byte {
	if l.pos+n >= len(l.input) {
		return 0
	}
	return l.input[l.pos+n]
}

// advance moves to the next position
func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		l.width = 1
		l.column++
		l.pos++
	}
}

// makeToken creates a token with current position
func (l *Lexer) makeToken(typ TokenType, value string) Token {
	return Token{
		Type:   typ,
		Value:  value,
		Line:   l.line,
		Column: l.column - len(value),
		Offset: l.start,
	}
}

// Helper functions

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
