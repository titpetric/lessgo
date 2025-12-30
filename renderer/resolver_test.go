package renderer

import (
	"testing"

	"github.com/titpetric/lessgo/evaluator"
)

func TestResolveValue(t *testing.T) {
	resolver := NewResolver(nil) // No file needed for these tests
	stack := NewStack()

	tests := []struct {
		name      string
		value     string
		variables map[string]string
		expected  string
		wantErr   bool
	}{
		{
			name:      "simple variable substitution",
			value:     "@color",
			variables: map[string]string{"color": "red"},
			expected:  "red",
			wantErr:   false,
		},
		{
			name:      "multiplication: 10px * 3",
			value:     "(10px * 3)",
			variables: map[string]string{},
			expected:  "30px",
			wantErr:   false,
		},
		{
			name:      "multiplication with variable: 10px * @n",
			value:     "(10px * @n)",
			variables: map[string]string{"n": "3"},
			expected:  "30px",
			wantErr:   false,
		},
		{
			name:      "addition: 10px + 5px",
			value:     "(10px + 5px)",
			variables: map[string]string{},
			expected:  "15px",
			wantErr:   false,
		},
		{
			name:      "CSS grid minmax passthrough",
			value:     "repeat(auto-fit, minmax(250px, 1fr))",
			variables: map[string]string{},
			expected:  "repeat(auto-fit, minmax(250px, 1fr))",
			wantErr:   false,
		},
		{
			name:      "standalone minmax passthrough",
			value:     "minmax(100px, 1fr)",
			variables: map[string]string{},
			expected:  "minmax(100px, 1fr)",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up stack with variables
			stack = NewStack()
			for k, v := range tt.variables {
				stack.Set(k, v)
			}

			got, err := resolver.ResolveValue(stack, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.expected {
				t.Errorf("ResolveValue() got %q, want %q\nValue: %q\nVariables: %v",
					got, tt.expected, tt.value, tt.variables)
			}
		})
	}
}

func TestResolveValue_ExtractFunctions(t *testing.T) {
	resolver := NewResolver(nil)

	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{
			name:     "minmax should not match min",
			value:    "minmax(250px, 1fr)",
			expected: []string{},
		},
		{
			name:     "repeat with minmax should not match min",
			value:    "repeat(auto-fit, minmax(250px, 1fr))",
			expected: []string{},
		},
		{
			name:     "actual min function should match",
			value:    "min(5px, 10px)",
			expected: []string{"min(5px, 10px)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolver.extractFunctionsFromValue(tt.value)
			if len(got) != len(tt.expected) {
				t.Errorf("extractFunctionsFromValue(%q) got %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string // Just the token text values
	}{
		{
			name:  "simple multiplication",
			input: "(10px * 3)",
			want:  []string{"(", "10px", "*", "3", ")"},
		},
		{
			name:  "addition with spaces",
			input: "10px + 5px",
			want:  []string{"10px", "+", "5px"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := evaluator.Tokenize(tt.input)
			if err != nil {
				t.Fatalf("Tokenize() error = %v", err)
			}

			if len(tokens) != len(tt.want) {
				t.Errorf("Tokenize() got %d tokens, want %d", len(tokens), len(tt.want))
				t.Logf("Tokens: %v", tokens)
				return
			}

			for i, tok := range tokens {
				if tok.Text != tt.want[i] {
					t.Errorf("Token %d: got %q, want %q", i, tok.Text, tt.want[i])
				}
			}
		})
	}
}
