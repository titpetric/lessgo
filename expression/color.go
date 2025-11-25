package expression

import (
	"fmt"
	"strconv"

	"github.com/titpetric/lessgo/internal/strings"
)

// Color represents a color value (RGB, HSL, named colors, etc)
type Color struct {
	R   uint8   // Red (0-255)
	G   uint8   // Green (0-255)
	B   uint8   // Blue (0-255)
	A   float64 // Alpha (0-1)
	H   float64 // Hue (0-360) for HSL colors
	S   float64 // Saturation (0-100) for HSL colors
	L   float64 // Lightness (0-100) for HSL colors
	HSL bool    // True if color is stored in HSL format
	Raw string  // original raw string
}

// ParseColor parses a color string into a Color
// Supports: #RRGGBB, #RGB, rgb(r, g, b), rgba(r, g, b, a), hsl(...), etc
func ParseColor(s string) (*Color, error) {
	s = strings.TrimSpace(s)

	// Hex color: #RRGGBB or #RGB
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s)
	}

	// rgb/rgba function
	if strings.HasPrefix(s, "rgb(") || strings.HasPrefix(s, "rgba(") {
		return parseRGBColor(s)
	}

	// hsl/hsla function
	if strings.HasPrefix(s, "hsl(") || strings.HasPrefix(s, "hsla(") {
		return parseHSLColor(s)
	}

	return nil, fmt.Errorf("invalid color: %s", s)
}

// parseHexColor parses hex color codes
func parseHexColor(s string) (*Color, error) {
	s = strings.TrimPrefix(s, "#")

	var r, g, b uint8
	var a float64 = 1.0

	if len(s) == 6 {
		// #RRGGBB
		rv, err := strconv.ParseUint(s[0:2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", s)
		}
		r = uint8(rv)

		gv, err := strconv.ParseUint(s[2:4], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", s)
		}
		g = uint8(gv)

		bv, err := strconv.ParseUint(s[4:6], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", s)
		}
		b = uint8(bv)
	} else if len(s) == 3 {
		// #RGB -> #RRGGBB
		rv, err := strconv.ParseUint(string(s[0])+string(s[0]), 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", s)
		}
		r = uint8(rv)

		gv, err := strconv.ParseUint(string(s[1])+string(s[1]), 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", s)
		}
		g = uint8(gv)

		bv, err := strconv.ParseUint(string(s[2])+string(s[2]), 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %s", s)
		}
		b = uint8(bv)
	} else {
		return nil, fmt.Errorf("invalid hex color: %s", s)
	}

	return &Color{R: r, G: g, B: b, A: a, Raw: "#" + s}, nil
}

// parseRGBColor parses rgb(r, g, b) or rgba(r, g, b, a)
func parseRGBColor(s string) (*Color, error) {
	var isAlpha bool
	var content string

	if strings.HasPrefix(s, "rgba(") {
		isAlpha = true
		content = strings.TrimPrefix(s, "rgba(")
	} else {
		content = strings.TrimPrefix(s, "rgb(")
	}

	content = strings.TrimSuffix(content, ")")

	// Split by comma
	parts := strings.Split(content, ",")
	if isAlpha && len(parts) != 4 {
		return nil, fmt.Errorf("rgba expects 4 arguments, got %d", len(parts))
	}
	if !isAlpha && len(parts) != 3 {
		return nil, fmt.Errorf("rgb expects 3 arguments, got %d", len(parts))
	}

	// Trim spaces from parts in-place
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Parse R, G, B
	r, err := strconv.ParseInt(parts[0], 10, 16)
	if err != nil || r < 0 || r > 255 {
		return nil, fmt.Errorf("invalid red value: %s", parts[0])
	}

	g, err := strconv.ParseInt(parts[1], 10, 16)
	if err != nil || g < 0 || g > 255 {
		return nil, fmt.Errorf("invalid green value: %s", parts[1])
	}

	b, err := strconv.ParseInt(parts[2], 10, 16)
	if err != nil || b < 0 || b > 255 {
		return nil, fmt.Errorf("invalid blue value: %s", parts[2])
	}

	var a float64 = 1.0
	if isAlpha {
		av, err := strconv.ParseFloat(parts[3], 64)
		if err != nil || av < 0 || av > 1 {
			return nil, fmt.Errorf("invalid alpha value: %s", parts[3])
		}
		a = av
	}

	return &Color{R: uint8(r), G: uint8(g), B: uint8(b), A: a, Raw: s}, nil
}

// parseHSLColor parses hsl(h, s, l) or hsla(h, s, l, a)
func parseHSLColor(s string) (*Color, error) {
	var isAlpha bool
	var content string

	if strings.HasPrefix(s, "hsla(") {
		isAlpha = true
		content = strings.TrimPrefix(s, "hsla(")
	} else {
		content = strings.TrimPrefix(s, "hsl(")
	}

	content = strings.TrimSuffix(content, ")")

	// Split by comma
	parts := strings.Split(content, ",")
	if isAlpha && len(parts) != 4 {
		return nil, fmt.Errorf("hsla expects 4 arguments, got %d", len(parts))
	}
	if !isAlpha && len(parts) != 3 {
		return nil, fmt.Errorf("hsl expects 3 arguments, got %d", len(parts))
	}

	// Trim spaces from parts in-place
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Parse H (0-360)
	h, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || h < 0 || h > 360 {
		return nil, fmt.Errorf("invalid hue value: %s", parts[0])
	}

	// Parse S (0-100%)
	sStr := strings.TrimSuffix(parts[1], "%")
	s_val, err := strconv.ParseFloat(sStr, 64)
	if err != nil || s_val < 0 || s_val > 100 {
		return nil, fmt.Errorf("invalid saturation value: %s", parts[1])
	}

	// Parse L (0-100%)
	lStr := strings.TrimSuffix(parts[2], "%")
	l_val, err := strconv.ParseFloat(lStr, 64)
	if err != nil || l_val < 0 || l_val > 100 {
		return nil, fmt.Errorf("invalid lightness value: %s", parts[2])
	}

	var a float64 = 1.0
	if isAlpha {
		av, err := strconv.ParseFloat(parts[3], 64)
		if err != nil || av < 0 || av > 1 {
			return nil, fmt.Errorf("invalid alpha value: %s", parts[3])
		}
		a = av
	}

	r, g, b := hslToRGB(h, s_val, l_val)

	return &Color{
		R: r, G: g, B: b, A: a,
		H: h, S: s_val, L: l_val,
		HSL: true,
		Raw: s,
	}, nil
}

// String returns the representation of the color
func (c *Color) String() string {
	if c.HSL {
		// Output in HSL format
		if c.A < 1.0 {
			return fmt.Sprintf("hsla(%g, %g%%, %g%%, %g)", c.H, c.S, c.L, c.A)
		}
		return fmt.Sprintf("hsl(%g, %g%%, %g%%)", c.H, c.S, c.L)
	}

	if c.A < 1.0 {
		// Return rgba format
		return fmt.Sprintf("rgba(%d, %d, %d, %g)", c.R, c.G, c.B, c.A)
	}

	// Prefer raw hex format if it was provided (preserves shorthand #333 vs #333333)
	if c.Raw != "" && strings.HasPrefix(c.Raw, "#") {
		return c.Raw
	}

	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// Lighten lightens the color by increasing lightness
func (c *Color) Lighten(amount float64) *Color {
	var h, s, l float64
	if c.HSL {
		h, s, l = c.H, c.S, c.L
	} else {
		h, s, l = rgbToHSL(c.R, c.G, c.B)
	}
	l = min(100, l+amount)
	r, g, b := hslToRGB(h, s, l)
	return &Color{
		R: r, G: g, B: b, A: c.A,
		H: h, S: s, L: l,
		HSL: c.HSL,
		Raw: "",
	}
}

// Darken darkens the color by decreasing lightness
func (c *Color) Darken(amount float64) *Color {
	var h, s, l float64
	if c.HSL {
		h, s, l = c.H, c.S, c.L
	} else {
		h, s, l = rgbToHSL(c.R, c.G, c.B)
	}
	l = max(0, l-amount)
	r, g, b := hslToRGB(h, s, l)
	return &Color{
		R: r, G: g, B: b, A: c.A,
		H: h, S: s, L: l,
		HSL: c.HSL,
		Raw: "",
	}
}

// Saturate increases the saturation
func (c *Color) Saturate(amount float64) *Color {
	var h, s, l float64
	if c.HSL {
		h, s, l = c.H, c.S, c.L
	} else {
		h, s, l = rgbToHSL(c.R, c.G, c.B)
	}
	s = min(100, s+amount)
	r, g, b := hslToRGB(h, s, l)
	return &Color{
		R: r, G: g, B: b, A: c.A,
		H: h, S: s, L: l,
		HSL: c.HSL,
		Raw: "",
	}
}

// Desaturate decreases the saturation
func (c *Color) Desaturate(amount float64) *Color {
	var h, s, l float64
	if c.HSL {
		h, s, l = c.H, c.S, c.L
	} else {
		h, s, l = rgbToHSL(c.R, c.G, c.B)
	}
	s = max(0, s-amount)
	r, g, b := hslToRGB(h, s, l)
	return &Color{
		R: r, G: g, B: b, A: c.A,
		H: h, S: s, L: l,
		HSL: c.HSL,
		Raw: "",
	}
}

// Spin rotates the hue by degrees
func (c *Color) Spin(degrees float64) *Color {
	var h, s, l float64
	if c.HSL {
		h, s, l = c.H, c.S, c.L
	} else {
		h, s, l = rgbToHSL(c.R, c.G, c.B)
	}
	h = h + degrees
	// Normalize hue to 0-360
	for h < 0 {
		h += 360
	}
	for h >= 360 {
		h -= 360
	}
	r, g, b := hslToRGB(h, s, l)
	return &Color{
		R: r, G: g, B: b, A: c.A,
		H: h, S: s, L: l,
		HSL: c.HSL,
		Raw: "",
	}
}

// rgbToHSL converts RGB to HSL (H: 0-360, S: 0-100, L: 0-100)
func rgbToHSL(r, g, b uint8) (float64, float64, float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := maxFloat(rf, gf, bf)
	min := minFloat(rf, gf, bf)
	l := (max + min) / 2.0

	if max == min {
		return 0, 0, l * 100
	}

	var h, s float64
	d := max - min

	if l > 0.5 {
		s = d / (2.0 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case rf:
		h = fmod((gf-bf)/d, 6.0)
	case gf:
		h = (bf-rf)/d + 2.0
	case bf:
		h = (rf-gf)/d + 4.0
	}

	h = h * 60.0
	if h < 0 {
		h += 360
	}

	return h, s * 100, l * 100
}

// hslToRGB converts HSL to RGB (H: 0-360, S: 0-100, L: 0-100)
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	h = fmod(h, 360.0)
	if h < 0 {
		h += 360
	}
	s = s / 100.0
	l = l / 100.0

	var c float64
	if l < 0.5 {
		c = 2.0 * l * s
	} else {
		c = (2.0 - 2.0*l) * s
	}

	hPrime := h / 60.0
	x := c * (1.0 - absFloat(fmod(hPrime, 2.0)-1.0))

	var r, g, b float64

	switch {
	case hPrime < 1.0:
		r, g, b = c, x, 0
	case hPrime < 2.0:
		r, g, b = x, c, 0
	case hPrime < 3.0:
		r, g, b = 0, c, x
	case hPrime < 4.0:
		r, g, b = 0, x, c
	case hPrime < 5.0:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	m := l - c/2.0

	// Round to nearest integer
	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255

	// Add 0.5 and truncate for proper rounding
	return uint8(r + 0.5), uint8(g + 0.5), uint8(b + 0.5)
}

// Utility functions
func minFloat(a, b, c float64) float64 {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func maxFloat(a, b, c float64) float64 {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}

func absFloat(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}

func fmod(a, b float64) float64 {
	return a - float64(int(a/b))*b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
