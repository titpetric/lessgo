package functions

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Color represents an RGBA color
type Color struct {
	R, G, B, A float64 // values 0-255 for RGB, 0-1 for A
}

// ParseColor parses a color from a hex string or rgb(a) function
func ParseColor(s string) (*Color, error) {
	s = strings.TrimSpace(s)

	// Handle hex colors
	if strings.HasPrefix(s, "#") {
		return ParseHex(s)
	}

	// Handle rgb() and rgba()
	if strings.HasPrefix(s, "rgb") {
		return ParseRGB(s)
	}

	return nil, fmt.Errorf("invalid color: %s", s)
}

// ParseHex parses a hex color string (#fff, #ffffff, #rrggbbaa)
func ParseHex(hex string) (*Color, error) {
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b, a float64
	a = 1.0

	switch len(hex) {
	case 3: // #rgb
		r = parseHexDigit(string(hex[0])) * 17 / 255.0
		g = parseHexDigit(string(hex[1])) * 17 / 255.0
		b = parseHexDigit(string(hex[2])) * 17 / 255.0
	case 4: // #rgba
		r = parseHexDigit(string(hex[0])) * 17 / 255.0
		g = parseHexDigit(string(hex[1])) * 17 / 255.0
		b = parseHexDigit(string(hex[2])) * 17 / 255.0
		a = parseHexDigit(string(hex[3])) / 15.0
	case 6: // #rrggbb
		r = parseHexByte(hex[0:2]) / 255.0
		g = parseHexByte(hex[2:4]) / 255.0
		b = parseHexByte(hex[4:6]) / 255.0
	case 8: // #rrggbbaa
		r = parseHexByte(hex[0:2]) / 255.0
		g = parseHexByte(hex[2:4]) / 255.0
		b = parseHexByte(hex[4:6]) / 255.0
		a = parseHexByte(hex[6:8]) / 255.0
	default:
		return nil, fmt.Errorf("invalid hex color: #%s", hex)
	}

	return &Color{r * 255, g * 255, b * 255, a}, nil
}

// ParseRGB parses rgb() or rgba() format
func ParseRGB(s string) (*Color, error) {
	var rgba bool
	if strings.HasPrefix(s, "rgba") {
		rgba = true
		s = s[5 : len(s)-1] // remove "rgba(" and ")"
	} else if strings.HasPrefix(s, "rgb") {
		s = s[4 : len(s)-1] // remove "rgb(" and ")"
	} else {
		return nil, fmt.Errorf("invalid rgb color: %s", s)
	}

	parts := strings.Split(s, ",")
	if rgba && len(parts) != 4 {
		if !rgba && len(parts) != 3 {
			return nil, fmt.Errorf("invalid rgb color format")
		}
	}

	// Trim spaces from parts
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	r, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, err
	}

	a := 1.0
	if rgba && len(parts) > 3 {
		a, err = strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return nil, err
		}
	}

	return &Color{r, g, b, a}, nil
}

func parseHexDigit(h string) float64 {
	n, _ := strconv.ParseInt(h, 16, 64)
	return float64(n)
}

func parseHexByte(h string) float64 {
	n, _ := strconv.ParseInt(h, 16, 64)
	return float64(n)
}

// ToHex returns the color as a hex string
func (c *Color) ToHex() string {
	r := uint8(math.Round(c.R))
	g := uint8(math.Round(c.G))
	b := uint8(math.Round(c.B))

	if c.A < 1.0 {
		a := uint8(math.Round(c.A * 255))
		return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// ToRGB returns the color as rgb() or rgba() format
func (c *Color) ToRGB() string {
	r := uint8(math.Round(c.R))
	g := uint8(math.Round(c.G))
	b := uint8(math.Round(c.B))

	if c.A < 1.0 {
		return fmt.Sprintf("rgba(%d, %d, %d, %g)", r, g, b, c.A)
	}

	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

// Lighten lightens a color by a percentage
func (c *Color) Lighten(amount float64) *Color {
	h, s, l := c.ToHSL()
	l = math.Min(1.0, l+amount)
	return HSLToColor(h, s, l, c.A)
}

// Darken darkens a color by a percentage
func (c *Color) Darken(amount float64) *Color {
	h, s, l := c.ToHSL()
	l = math.Max(0.0, l-amount)
	return HSLToColor(h, s, l, c.A)
}

// Saturate increases saturation
func (c *Color) Saturate(amount float64) *Color {
	h, s, l := c.ToHSL()
	s = math.Min(1.0, s+amount)
	return HSLToColor(h, s, l, c.A)
}

// Desaturate decreases saturation
func (c *Color) Desaturate(amount float64) *Color {
	h, s, l := c.ToHSL()
	s = math.Max(0.0, s-amount)
	return HSLToColor(h, s, l, c.A)
}

// Spin rotates the hue
func (c *Color) Spin(degrees float64) *Color {
	h, s, l := c.ToHSL()
	h = math.Mod(h+degrees, 360)
	if h < 0 {
		h += 360
	}
	return HSLToColor(h, s, l, c.A)
}

// Mix mixes two colors
func (c *Color) Mix(other *Color, weight float64) *Color {
	weight = math.Max(0, math.Min(1, weight))
	return &Color{
		R: c.R*(1-weight) + other.R*weight,
		G: c.G*(1-weight) + other.G*weight,
		B: c.B*(1-weight) + other.B*weight,
		A: c.A*(1-weight) + other.A*weight,
	}
}

// ToHSL converts RGB to HSL
func (c *Color) ToHSL() (h, s, l float64) {
	r := c.R / 255.0
	g := c.G / 255.0
	b := c.B / 255.0

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = math.Mod((g-b)/d+6, 6)
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h *= 60
	}

	return h, s, l
}

// HSLToColor converts HSL to RGB Color
func HSLToColor(h, s, l float64, a float64) *Color {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	// Normalize s and l to 0-1 if they're not already
	if s > 1 {
		s = 1
	}
	if l > 1 {
		l = 1
	}

	c := (1 - math.Abs(2*l-1)) * s
	hp := h / 60
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))

	var r1, g1, b1 float64
	switch {
	case hp >= 0 && hp < 1:
		r1, g1, b1 = c, x, 0
	case hp >= 1 && hp < 2:
		r1, g1, b1 = x, c, 0
	case hp >= 2 && hp < 3:
		r1, g1, b1 = 0, c, x
	case hp >= 3 && hp < 4:
		r1, g1, b1 = 0, x, c
	case hp >= 4 && hp < 5:
		r1, g1, b1 = x, 0, c
	case hp >= 5 && hp < 6:
		r1, g1, b1 = c, 0, x
	}

	m := l - c/2
	return &Color{
		R: (r1 + m) * 255,
		G: (g1 + m) * 255,
		B: (b1 + m) * 255,
		A: a,
	}
}

// Greyscale returns the greyscale version of the color
func (c *Color) Greyscale() *Color {
	h, _, l := c.ToHSL()
	return HSLToColor(h, 0, l, c.A)
}
