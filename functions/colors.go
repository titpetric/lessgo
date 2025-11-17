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

// ParseColor parses a color from a hex string, rgb(a) function, or CSS keyword
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

	// Handle CSS color keywords
	if hex, ok := cssColorKeywords[strings.ToLower(s)]; ok {
		return ParseHex(hex)
	}

	return nil, fmt.Errorf("invalid color: %s", s)
}

// cssColorKeywords maps CSS color names to hex values
var cssColorKeywords = map[string]string{
	"aliceblue":            "#f0f8ff",
	"antiquewhite":         "#faebd7",
	"aqua":                 "#00ffff",
	"aquamarine":           "#7fffd4",
	"azure":                "#f0ffff",
	"beige":                "#f5f5dc",
	"bisque":               "#ffe4c4",
	"black":                "#000000",
	"blanchedalmond":       "#ffebcd",
	"blue":                 "#0000ff",
	"blueviolet":           "#8a2be2",
	"brown":                "#a52a2a",
	"burlywood":            "#deb887",
	"cadetblue":            "#5f9ea0",
	"chartreuse":           "#7fff00",
	"chocolate":            "#d2691e",
	"coral":                "#ff7f50",
	"cornflowerblue":       "#6495ed",
	"cornsilk":             "#fff8dc",
	"crimson":              "#dc143c",
	"cyan":                 "#00ffff",
	"darkblue":             "#00008b",
	"darkcyan":             "#008b8b",
	"darkgoldenrod":        "#b8860b",
	"darkgray":             "#a9a9a9",
	"darkgrey":             "#a9a9a9",
	"darkgreen":            "#006400",
	"darkkhaki":            "#bdb76b",
	"darkmagenta":          "#8b008b",
	"darkolivegreen":       "#556b2f",
	"darkorange":           "#ff8c00",
	"darkorchid":           "#9932cc",
	"darkred":              "#8b0000",
	"darksalmon":           "#e9967a",
	"darkseagreen":         "#8fbc8f",
	"darkslateblue":        "#483d8b",
	"darkslategray":        "#2f4f4f",
	"darkslategrey":        "#2f4f4f",
	"darkturquoise":        "#00ced1",
	"darkviolet":           "#9400d3",
	"deeppink":             "#ff1493",
	"deepskyblue":          "#00bfff",
	"dimgray":              "#696969",
	"dimgrey":              "#696969",
	"dodgerblue":           "#1e90ff",
	"firebrick":            "#b22222",
	"floralwhite":          "#fffaf0",
	"forestgreen":          "#228b22",
	"fuchsia":              "#ff00ff",
	"gainsboro":            "#dcdcdc",
	"ghostwhite":           "#f8f8ff",
	"gold":                 "#ffd700",
	"goldenrod":            "#daa520",
	"gray":                 "#808080",
	"grey":                 "#808080",
	"green":                "#008000",
	"greenyellow":          "#adff2f",
	"honeydew":             "#f0fff0",
	"hotpink":              "#ff69b4",
	"indianred":            "#cd5c5c",
	"indigo":               "#4b0082",
	"ivory":                "#fffff0",
	"khaki":                "#f0e68c",
	"lavender":             "#e6e6fa",
	"lavenderblush":        "#fff0f5",
	"lawngreen":            "#7cfc00",
	"lemonchiffon":         "#fffacd",
	"lightblue":            "#add8e6",
	"lightcoral":           "#f08080",
	"lightcyan":            "#e0ffff",
	"lightgoldenrodyellow": "#fafad2",
	"lightgray":            "#d3d3d3",
	"lightgrey":            "#d3d3d3",
	"lightgreen":           "#90ee90",
	"lightpink":            "#ffb6c1",
	"lightsalmon":          "#ffa07a",
	"lightseagreen":        "#20b2aa",
	"lightskyblue":         "#87cefa",
	"lightslategray":       "#778899",
	"lightslategrey":       "#778899",
	"lightsteelblue":       "#b0c4de",
	"lightyellow":          "#ffffe0",
	"lime":                 "#00ff00",
	"limegreen":            "#32cd32",
	"linen":                "#faf0e6",
	"magenta":              "#ff00ff",
	"maroon":               "#800000",
	"mediumaquamarine":     "#66cdaa",
	"mediumblue":           "#0000cd",
	"mediumorchid":         "#ba55d3",
	"mediumpurple":         "#9370db",
	"mediumseagreen":       "#3cb371",
	"mediumslateblue":      "#7b68ee",
	"mediumspringgreen":    "#00fa9a",
	"mediumturquoise":      "#48d1cc",
	"mediumvioletred":      "#c71585",
	"midnightblue":         "#191970",
	"mintcream":            "#f5fffa",
	"mistyrose":            "#ffe4e1",
	"moccasin":             "#ffe4b5",
	"navajowhite":          "#ffdead",
	"navy":                 "#000080",
	"oldlace":              "#fdf5e6",
	"olive":                "#808000",
	"olivedrab":            "#6b8e23",
	"orange":               "#ffa500",
	"orangered":            "#ff4500",
	"orchid":               "#da70d6",
	"palegoldenrod":        "#eee8aa",
	"palegreen":            "#98fb98",
	"paleturquoise":        "#afeeee",
	"palevioletred":        "#db7093",
	"papayawhip":           "#ffefd5",
	"peachpuff":            "#ffdab9",
	"peru":                 "#cd853f",
	"pink":                 "#ffc0cb",
	"plum":                 "#dda0dd",
	"powderblue":           "#b0e0e6",
	"purple":               "#800080",
	"red":                  "#ff0000",
	"rosybrown":            "#bc8f8f",
	"royalblue":            "#4169e1",
	"saddlebrown":          "#8b4513",
	"salmon":               "#fa8072",
	"sandybrown":           "#f4a460",
	"seagreen":             "#2e8b57",
	"seashell":             "#fff5ee",
	"sienna":               "#a0522d",
	"silver":               "#c0c0c0",
	"skyblue":              "#87ceeb",
	"slateblue":            "#6a5acd",
	"slategray":            "#708090",
	"slategrey":            "#708090",
	"snow":                 "#fffafa",
	"springgreen":          "#00ff7f",
	"steelblue":            "#4682b4",
	"tan":                  "#d2b48c",
	"teal":                 "#008080",
	"thistle":              "#d8bfd8",
	"tomato":               "#ff6347",
	"turquoise":            "#40e0d0",
	"violet":               "#ee82ee",
	"wheat":                "#f5deb3",
	"white":                "#ffffff",
	"whitesmoke":           "#f5f5f5",
	"yellow":               "#ffff00",
	"yellowgreen":          "#9acd32",
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

// Luma returns the perceived brightness (luminance) of the color as a percentage
// Uses the formula: luma = 0.2126*R + 0.7152*G + 0.0722*B (relative luminance)
func (c *Color) Luma() float64 {
	// Normalize RGB values to 0-1
	r := c.R / 255.0
	g := c.G / 255.0
	b := c.B / 255.0

	// Apply gamma correction (sRGB)
	r = gammaCorrect(r)
	g = gammaCorrect(g)
	b = gammaCorrect(b)

	// ITU-R BT.709 luminance
	lum := 0.2126*r + 0.7152*g + 0.0722*b
	return lum * 100 // Return as percentage
}

// gammaCorrect applies gamma correction for sRGB
func gammaCorrect(value float64) float64 {
	if value <= 0.03928 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}
