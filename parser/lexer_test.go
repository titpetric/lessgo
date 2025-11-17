package parser_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sourcegraph/lessgo/parser"
)

func TestLexerBasics(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []parser.TokenType
	}{
		{
			name:  "empty input",
			input: "",
			expected: []parser.TokenType{
				parser.TokenEOF,
			},
		},
		{
			name:  "simple rule",
			input: "body { color: red; }",
			expected: []parser.TokenType{
				parser.TokenIdent,     // body
				parser.TokenLBrace,    // {
				parser.TokenIdent,     // color
				parser.TokenColon,     // :
				parser.TokenIdent,     // red
				parser.TokenSemicolon, // ;
				parser.TokenRBrace,    // }
				parser.TokenEOF,
			},
		},
		{
			name:  "variable",
			input: "@primary: #fff;",
			expected: []parser.TokenType{
				parser.TokenVariable,  // primary
				parser.TokenColon,     // :
				parser.TokenColor,     // #fff
				parser.TokenSemicolon, // ;
				parser.TokenEOF,
			},
		},
		{
			name:  "comment removal",
			input: "/* comment */ body { /* inline */ color: red; }",
			expected: []parser.TokenType{
				parser.TokenIdent,     // body
				parser.TokenLBrace,    // {
				parser.TokenIdent,     // color
				parser.TokenColon,     // :
				parser.TokenIdent,     // red
				parser.TokenSemicolon, // ;
				parser.TokenRBrace,    // }
				parser.TokenEOF,
			},
		},
		{
			name:  "line comment",
			input: "body { // comment\ncolor: red; }",
			expected: []parser.TokenType{
				parser.TokenIdent,     // body
				parser.TokenLBrace,    // {
				parser.TokenIdent,     // color
				parser.TokenColon,     // :
				parser.TokenIdent,     // red
				parser.TokenSemicolon, // ;
				parser.TokenRBrace,    // }
				parser.TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := parser.NewLexer(tt.input)
			tokens := lexer.Tokenize()

			require.Equal(t, len(tt.expected), len(tokens), "token count mismatch")

			for i, tok := range tokens {
				require.Equal(t, tt.expected[i], tok.Type,
					"token %d type mismatch: expected %v, got %v", i, tt.expected[i], tok.Type)
			}
		})
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		value string
	}{
		{
			name:  "double quoted string",
			input: `"hello world"`,
			value: "hello world",
		},
		{
			name:  "single quoted string",
			input: `'hello world'`,
			value: "hello world",
		},
		{
			name:  "string with escapes",
			input: `"hello\nworld"`,
			value: "hello\nworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := parser.NewLexer(tt.input)
			tokens := lexer.Tokenize()

			require.True(t, len(tokens) > 0, "no tokens")
			require.Equal(t, parser.TokenString, tokens[0].Type)
			require.Equal(t, tt.value, tokens[0].Value)
		})
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		value string
	}{
		{
			name:  "integer",
			input: "42",
			value: "42",
		},
		{
			name:  "float",
			input: "3.14",
			value: "3.14",
		},
		{
			name:  "with unit",
			input: "16px",
			value: "16px",
		},
		{
			name:  "percentage",
			input: "50%",
			value: "50%",
		},
		{
			name:  "negative number",
			input: "-10",
			value: "-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := parser.NewLexer(tt.input)
			tokens := lexer.Tokenize()

			require.True(t, len(tokens) > 0)
			require.Equal(t, parser.TokenNumber, tokens[0].Type)
			require.Equal(t, tt.value, tokens[0].Value)
		})
	}
}

func TestLexerColors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		value string
	}{
		{
			name:  "hex 3-digit",
			input: "#fff",
			value: "#fff",
		},
		{
			name:  "hex 6-digit",
			input: "#ffffff",
			value: "#ffffff",
		},
		{
			name:  "hex 8-digit (rgba)",
			input: "#ffffff80",
			value: "#ffffff80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := parser.NewLexer(tt.input)
			tokens := lexer.Tokenize()

			require.True(t, len(tokens) > 0)
			require.Equal(t, parser.TokenColor, tokens[0].Type)
			require.Equal(t, tt.value, tokens[0].Value)
		})
	}
}
