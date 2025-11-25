package evaluator

import (
	"fmt"
	"unicode"

	"github.com/titpetric/lessgo/internal/strings"
)

type TokenType int

const (
	TokenIdent TokenType = iota
	TokenOp
	TokenParen
	TokenValue
)

type Token struct {
	Type TokenType
	Text string
}

func IsExpression(tokens []Token) bool {
	for _, tok := range tokens {
		if tok.Type == TokenOp {
			return true
		}
	}
	return false
}

func Tokenize(input string) ([]Token, error) {
	var tokens []Token
	runes := []rune(input)
	i := 0

	var current string
	var open bool
	var space bool
	var parenDepth int // Track nesting depth of parentheses

	for i < len(runes) {
		r := runes[i]

		if open {
			if r == '(' {
				parenDepth++
				current += string(r)
			} else if r == ')' {
				if parenDepth > 0 {
					parenDepth--
					current += string(r)
				} else {
					// This closes our function call
					current += string(r)

					// Append to last token if it exists, otherwise create new token
					if len(tokens) > 0 {
						tokens[len(tokens)-1].Text += current
					} else {
						tokens = append(tokens, Token{Type: TokenValue, Text: current})
					}

					current = ""
					open = false
				}
			} else {
				current += string(r)
			}
			i++
			continue
		}

		// skip spaces
		if unicode.IsSpace(r) {
			space = true
			i++
			continue
		}

		if r == '(' {
			// Only treat as function call if there was a preceding identifier token
			// Otherwise it's grouping parentheses
			if len(tokens) > 0 && tokens[len(tokens)-1].Type == TokenIdent {
				open = true
				current = string(r)
				i++
				continue
			}
			// Otherwise fall through to treat as regular paren
		}

		// parentheses
		if r == '(' || r == ')' {
			tokens = append(tokens, Token{Type: TokenParen, Text: string(r)})
			i++
			continue
		}

		// operators: = > <
		if space && (r == '=' || r == '>' || r == '<' || r == '*' || r == '+' || r == '-' || r == '/') {
			body := string(r)
			if r == '*' {
				//body = "\\*"
			}
			tokens = append(tokens, Token{Type: TokenOp, Text: body})
			space = false
			i++
			continue
		}
		space = false

		// identifiers (@var)
		if r == '@' {
			start := i
			i++
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenIdent, Text: string(runes[start:i])})
			continue
		}

		// bare values (dark)
		if unicode.IsLetter(r) {
			start := i
			i++
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenValue, Text: string(runes[start:i])})
			continue
		}

		if unicode.IsDigit(r) {
			start := i
			i++
			for i < len(runes) && (unicode.IsDigit(runes[i]) || unicode.IsLetter(runes[i])) {
				i++
			}
			tokens = append(tokens, Token{Type: TokenValue, Text: string(runes[start:i])})
			continue
		}

		return []Token{
			Token{Type: TokenValue, Text: string(input)},
		}, nil
	}

	return tokens, nil
}

// ParseExpression converts input like "(@var = dark)" into `(var == "dark")`.
func ParseExpression(input string) (string, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return "", err
	}

	out := make([]string, 0, len(tokens))

	for _, t := range tokens {
		switch t.Type {
		case TokenParen:
			out = append(out, t.Text)

		case TokenIdent:
			// drop '@'
			out = append(out, strings.TrimPrefix(t.Text, "@"))

		case TokenOp:
			// map "=" to "==", keep others
			if t.Text == "=" {
				out = append(out, "==")
			} else {
				out = append(out, t.Text)
			}

		case TokenValue:
			// quote values
			out = append(out, fmt.Sprintf("%q", t.Text))
		}
	}

	return strings.Join(out, " "), nil
}
