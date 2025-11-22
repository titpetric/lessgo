package expression

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		wantNum  float64
		wantUnit string
		wantErr  bool
	}{
		{"10px", 10, "px", false},
		{"50%", 0.5, "", false},
		{"-5em", -5, "em", false},
		{"1.5", 1.5, "", false},
		{"3", 3, "", false},
		{"0.25rem", 0.25, "rem", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%s) err = %v, want %v", tt.input, err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if v.Number != tt.wantNum || v.Unit != tt.wantUnit {
				t.Errorf("Parse(%s) = %g%s, want %g%s", tt.input, v.Number, v.Unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		wantNum  float64
		wantUnit string
		wantErr  bool
	}{
		{"10px", "5px", 15, "px", false},
		{"10px", "5", 15, "px", false},
		{"10", "5px", 15, "px", false},
		{"10px", "5em", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.left+"+"+tt.right, func(t *testing.T) {
			left, _ := Parse(tt.left)
			right, _ := Parse(tt.right)
			v, err := left.Add(right)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Add err = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if v.Number != tt.wantNum || v.Unit != tt.wantUnit {
				t.Errorf("Add = %g%s, want %g%s", v.Number, v.Unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		wantNum  float64
		wantUnit string
		wantErr  bool
	}{
		{"10px", "5", 50, "px", false},
		{"5", "10px", 50, "px", false},
		{"15px", "2", 30, "px", false},
		{"10px", "5px", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.left+"*"+tt.right, func(t *testing.T) {
			left, _ := Parse(tt.left)
			right, _ := Parse(tt.right)
			v, err := left.Multiply(right)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Multiply err = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if v.Number != tt.wantNum || v.Unit != tt.wantUnit {
				t.Errorf("Multiply = %g%s, want %g%s", v.Number, v.Unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		wantNum  float64
		wantUnit string
		wantErr  bool
	}{
		{"50px", "5", 10, "px", false},
		{"50px", "10px", 5, "", false},
		{"24px", "2", 12, "px", false},
		{"10px", "0", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.left+"/"+tt.right, func(t *testing.T) {
			left, _ := Parse(tt.left)
			right, _ := Parse(tt.right)
			v, err := left.Divide(right)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Divide err = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if v.Number != tt.wantNum || v.Unit != tt.wantUnit {
				t.Errorf("Divide = %g%s, want %g%s", v.Number, v.Unit, tt.wantNum, tt.wantUnit)
			}
		})
	}
}

func TestPercentToDecimal(t *testing.T) {
	v, _ := Parse("50%")
	if v.Number != 0.5 || v.Unit != "" {
		t.Errorf("50%% = %g%s, want 0.5", v.Number, v.Unit)
	}
}
