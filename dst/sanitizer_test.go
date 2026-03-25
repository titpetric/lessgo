package dst

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/lessgo/internal/strings"
)

func TestSanitizeBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "minified CSS with nested block",
			input: `.classname { font-weight: bold; color: #f70; a { text-decoration: underline } }`,
			// Trailing whitespace is trimmed before closing braces
			expected: ".classname {\n font-weight: bold;\n color: #f70;\n a {\n text-decoration: underline;\n}\n}",
		},
		{
			name:     "already formatted",
			input:    ".foo {\n  color: red;\n}\n",
			expected: ".foo {\n  color: red;\n}\n",
		},
		{
			name:  "quoted strings preserved",
			input: `.foo { content: "{ ; }"; }`,
			// Content inside quotes is preserved
			expected: ".foo {\n content: \"{ ; }\";\n}",
		},
		{
			name:  "declaration without trailing semicolon",
			input: `.foo { color: red }`,
			// Semicolon is injected before }, trailing whitespace trimmed
			expected: ".foo {\n color: red;\n}",
		},
		{
			name:  "interpolation preserved",
			input: `.@{prefix} { color: red; }`,
			// @{...} interpolation blocks are not broken by newlines
			expected: ".@{prefix} {\n color: red;\n}",
		},
		{
			name:  "comment only block",
			input: ".foo { // comment\n}",
			// No semicolon added when block only contains comments
			expected: ".foo {\n // comment\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(SanitizeBytes([]byte(tt.input)))
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeReader(t *testing.T) {
	input := `.foo { color: red }`
	expected := ".foo {\n color: red;\n}"

	reader := SanitizeReader(strings.NewReader(input))
	result, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, expected, string(result))
}
