package expression

import (
	"testing"
)

func TestEvalBasic(t *testing.T) {
	e, _ := NewEvaluator(nil)

	tests := []struct {
		expr     string
		wantNum  float64
		wantUnit string
	}{
		{"10px", 10, "px"},
		{"10px * 2", 20, "px"},
		{"15px * 10", 150, "px"},
		{"10px + 5px", 15, "px"},
		{"20px - 5px", 15, "px"},
		{"24px / 2", 12, "px"},
		{"50px / 10px", 5, ""},
		{"50%", 0.5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			v, err := e.Eval(tt.expr)
			if err != nil {
				t.Fatalf("Eval(%s) err = %v", tt.expr, err)
			}
			if v.Number != tt.wantNum || v.Unit != tt.wantUnit {
				t.Errorf("Eval(%s) = %g%s, want %g%s", tt.expr, v.Number, v.Unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}

func TestEvalWithVariables(t *testing.T) {
	e, _ := NewEvaluator(nil)
	e.SetVariable("base", NewValue(10, "px"))

	tests := []struct {
		expr     string
		wantNum  float64
		wantUnit string
	}{
		{"@base", 10, "px"},
		{"@base * 2", 20, "px"},
		{"@base + 5px", 15, "px"},
		{"@base * 10", 100, "px"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			v, err := e.Eval(tt.expr)
			if err != nil {
				t.Fatalf("Eval(%s) err = %v", tt.expr, err)
			}
			if v.Number != tt.wantNum || v.Unit != tt.wantUnit {
				t.Errorf("Eval(%s) = %g%s, want %g%s", tt.expr, v.Number, v.Unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}

func TestEvalComplexExpressions(t *testing.T) {
	e, _ := NewEvaluator(nil)
	e.SetVariable("size", NewValue(10, "px"))
	e.SetVariable("multiplier", NewValue(2, ""))

	tests := []struct {
		expr     string
		wantNum  float64
		wantUnit string
	}{
		{"@size * @multiplier", 20, "px"},
		{"@size + @size", 20, "px"},
		{"@size * 3 + 5px", 35, "px"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			v, err := e.Eval(tt.expr)
			if err != nil {
				t.Fatalf("Eval(%s) err = %v", tt.expr, err)
			}
			if v.Number != tt.wantNum || v.Unit != tt.wantUnit {
				t.Errorf("Eval(%s) = %g%s, want %g%s", tt.expr, v.Number, v.Unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}
