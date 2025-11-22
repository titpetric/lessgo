package renderer

import (
	"testing"

	"github.com/titpetric/lessgo/dst"
)

func TestEvaluateGuard(t *testing.T) {
	r := NewRenderer()
	stack := NewStack()

	tests := []struct {
		name      string
		guard     *dst.Guard
		variables map[string]string
		expected  bool
		wantErr   bool
	}{
		{
			name:      "no guard",
			guard:     &dst.Guard{},
			variables: map[string]string{},
			expected:  true,
			wantErr:   false,
		},
		{
			name:      "simple comparison: 3 > 0",
			guard:     &dst.Guard{Condition: "(@n > 0)"},
			variables: map[string]string{"n": "3"},
			expected:  true,
			wantErr:   false,
		},
		{
			name:      "simple comparison: 0 > 0 should be false",
			guard:     &dst.Guard{Condition: "(@n > 0)"},
			variables: map[string]string{"n": "0"},
			expected:  false,
			wantErr:   false,
		},
		{
			name:      "equality: 5 = 5",
			guard:     &dst.Guard{Condition: "(@val = 5)"},
			variables: map[string]string{"val": "5"},
			expected:  true,
			wantErr:   false,
		},
		{
			name:      "with units: 10px > 5px",
			guard:     &dst.Guard{Condition: "(@width > 5px)"},
			variables: map[string]string{"width": "10px"},
			expected:  true,
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

			got, err := r.evaluateGuard(stack, tt.guard)

			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateGuard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.expected {
				t.Errorf("evaluateGuard() got %v, want %v\nGuard: %+v\nVariables: %+v",
					got, tt.expected, tt.guard, tt.variables)
			}
		})
	}
}
