package functions

import (
	"fmt"
	"strings"
)

// FuncMap is a map of function name to callable function.
// It works similarly to html/template.FuncMap for registering custom functions.
type FuncMap map[string]interface{}

// DefaultFuncMap returns all built-in LESS functions
func DefaultFuncMap() FuncMap {
	return FuncMap{
		// Math functions
		"ceil":       func(v string) string { return Ceil(v) },
		"floor":      func(v string) string { return Floor(v) },
		"round":      func(v string) string { return Round(v) },
		"abs":        func(v string) string { return Abs(v) },
		"sqrt":       func(v string) string { return Sqrt(v) },
		"pow":        func(a, b string) string { return Pow(a, b) },
		"min":        func(vals ...string) string { return Min(vals...) },
		"max":        func(vals ...string) string { return Max(vals...) },
		"percentage": func(v string) string { return Percentage(v) },
		"mod":        func(a, b string) string { return Mod(a, b) },
		"sin":        func(v string) string { return Sin(v) },
		"cos":        func(v string) string { return Cos(v) },
		"tan":        func(v string) string { return Tan(v) },
		"asin":       func(v string) string { return Asin(v) },
		"acos":       func(v string) string { return Acos(v) },
		"atan":       func(v string) string { return Atan(v) },
		"pi":         func() string { return Pi() },

		// String functions
		"escape": func(s string) string { return Escape(s) },
		"e":      func(s string) string { return E(s) },
		"replace": func(str, pattern, replacement string, args ...string) string {
			if len(args) > 0 {
				return Replace(str, pattern, replacement, args[0])
			}
			return Replace(str, pattern, replacement)
		},
		"format": func(format string, args ...string) string { return Format(format, args...) },

		// List functions
		"length":  func(v string) string { return Length(v) },
		"extract": func(list, idx string) string { return Extract(list, idx) },
		"range": func(start, end string, args ...string) string {
			if len(args) > 0 {
				return Range(start, end, args[0])
			}
			return Range(start, end)
		},

		// Type checking functions
		"isnumber":     func(v string) string { return IsNumberFunction(v) },
		"isstring":     func(v string) string { return IsStringFunction(v) },
		"iscolor":      func(v string) string { return IsColorFunction(v) },
		"iskeyword":    func(v string) string { return IsKeywordFunction(v) },
		"isurl":        func(v string) string { return IsURLFunction(v) },
		"ispixel":      func(v string) string { return IsPixelFunction(v) },
		"isem":         func(v string) string { return IsEmFunction(v) },
		"ispercentage": func(v string) string { return IsPercentageFunction(v) },
		"isunit":       func(v, u string) string { return IsUnitFunction(v, u) },
		"isruleset":    func(v string) string { return IsRulesetFunction(v) },
		"islist":       func(v string) string { return IsListFunction(v) },
		"boolean":      func(v string) string { return Boolean(v) },

		// Color definition functions
		"rgb":  func(r, g, b string) string { return RGB(r, g, b) },
		"rgba": func(r, g, b, a string) string { return RGBA(r, g, b, a) },
		"hsl":  func(h, s, l string) string { return HSL(h, s, l) },
		"hsla": func(h, s, l, a string) string { return HSLA(h, s, l, a) },

		// Color channel extraction functions
		"hue":        func(c string) string { return Hue(c) },
		"saturation": func(c string) string { return Saturation(c) },
		"lightness":  func(c string) string { return Lightness(c) },
		"red":        func(c string) string { return Red(c) },
		"green":      func(c string) string { return Green(c) },
		"blue":       func(c string) string { return Blue(c) },
		"alpha":      func(c string) string { return Alpha(c) },
		"luma":       func(c string) string { return LumaFunction(c) },
		"luminance":  func(c string) string { return Luminance(c) },

		// Color manipulation functions
		"lighten":    func(c, a string) string { return lighten(c, a) },
		"darken":     func(c, a string) string { return darken(c, a) },
		"saturate":   func(c, a string) string { return saturate(c, a) },
		"desaturate": func(c, a string) string { return desaturate(c, a) },
		"spin":       func(c, a string) string { return spin(c, a) },
		"mix": func(c1, c2 string, args ...string) string {
			if len(args) > 0 {
				return mix(c1, c2, args[0])
			}
			return mix(c1, c2, "50%")
		},
		"shade":     func(c, w string) string { return Shade(c, w) },
		"tint":      func(c, w string) string { return Tint(c, w) },
		"greyscale": func(c string) string { return greyscale(c) },
		"fade":      func(c, a string) string { return Fade(c, a) },
		"fadein":    func(c, a string) string { return Fadein(c, a) },
		"fadeout":   func(c, a string) string { return Fadeout(c, a) },
		"contrast":  func(c string, args ...string) string { return Contrast(c, args...) },

		// Color blending functions
		"multiply":   func(c1, c2 string) string { return Multiply(c1, c2) },
		"screen":     func(c1, c2 string) string { return Screen(c1, c2) },
		"overlay":    func(c1, c2 string) string { return Overlay(c1, c2) },
		"softlight":  func(c1, c2 string) string { return Softlight(c1, c2) },
		"hardlight":  func(c1, c2 string) string { return Hardlight(c1, c2) },
		"difference": func(c1, c2 string) string { return Difference(c1, c2) },
		"exclusion":  func(c1, c2 string) string { return Exclusion(c1, c2) },
		"average":    func(c1, c2 string) string { return Average(c1, c2) },
		"negation":   func(c1, c2 string) string { return Negation(c1, c2) },

		// Logical functions
		"if": func(cond, trueVal, falseVal string) string { return If(cond, trueVal, falseVal) },

		// Utility functions
		"color": func(c string) string { return ColorFunction(c) },
		"unit": func(v string, args ...string) string {
			if len(args) > 0 {
				return Unit(v, args[0])
			}
			return Unit(v, "")
		},
		"get-unit": func(v string) string { return GetUnit(v) },
		"convert":  func(v, u string) string { return Convert(v, u) },
	}
}

// Helper functions that wrap Color methods to return strings
func lighten(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Lighten(amountVal)
	return result.ToHex()
}

func darken(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Darken(amountVal)
	return result.ToHex()
}

func saturate(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Saturate(amountVal)
	return result.ToHex()
}

func desaturate(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Desaturate(amountVal)
	return result.ToHex()
}

func spin(colorStr, angle string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	angleVal := parseNumber(angle)
	result := color.Spin(angleVal)
	return result.ToHex()
}

func mix(color1Str, color2Str string, weight ...string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	weightVal := 0.5
	if len(weight) > 0 && weight[0] != "" {
		weightVal = parseNumber(weight[0]) / 100.0
		weightVal = clamp(weightVal, 0, 1)
	}

	result := c1.Mix(c2, weightVal)
	return result.ToHex()
}

func greyscale(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	result := color.Greyscale()

	// Preserve color format
	if strings.HasPrefix(colorStr, "hsl") {
		h, s, l := result.ToHSL()
		if strings.HasPrefix(colorStr, "hsla") {
			return fmt.Sprintf("hsla(%g, %g%%, %g%%, %g)", h, s*100, l*100, result.A)
		}
		return fmt.Sprintf("hsl(%g, %g%%, %g%%)", h, s*100, l*100)
	}
	if strings.HasPrefix(colorStr, "rgb") {
		if strings.HasPrefix(colorStr, "rgba") {
			return fmt.Sprintf("rgba(%d, %d, %d, %g)", uint8(result.R), uint8(result.G), uint8(result.B), result.A)
		}
		return fmt.Sprintf("rgb(%d, %d, %d)", uint8(result.R), uint8(result.G), uint8(result.B))
	}

	return result.ToHex()
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
