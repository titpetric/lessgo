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

// ParseColor parses a color from a hex string, rgb(a) function, hsl(a) function, or CSS keyword
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

	// Handle hsl() and hsla()
	if strings.HasPrefix(s, "hsl") {
		return ParseHSL(s)
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

// ParseHSL parses hsl() or hsla() format
func ParseHSL(input string) (*Color, error) {
	var hsla bool
	s := input
	if strings.HasPrefix(s, "hsla") {
		hsla = true
		s = s[5 : len(s)-1] // remove "hsla(" and ")"
	} else if strings.HasPrefix(s, "hsl") {
		s = s[4 : len(s)-1] // remove "hsl(" and ")"
	} else {
		return nil, fmt.Errorf("invalid hsl color: %s", input)
	}

	parts := strings.Split(s, ",")
	if hsla && len(parts) != 4 {
		if !hsla && len(parts) != 3 {
			return nil, fmt.Errorf("invalid hsl color format")
		}
	}

	// Trim spaces and remove % from parts
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		parts[i] = strings.TrimSuffix(parts[i], "%")
	}

	h, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, err
	}
	sat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, err
	}
	l, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, err
	}

	a := 1.0
	if hsla && len(parts) > 3 {
		a, err = strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return nil, err
		}
	}

	// Convert HSL percentages to 0-1 range
	sat = sat / 100.0
	l = l / 100.0

	return HSLToColor(h, sat, l, a), nil
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

// HSVToColor converts HSV to RGB Color
func HSVToColor(h, s, v float64, a float64) *Color {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	// Normalize s and v to 0-1 if they're not already
	if s > 1 {
		s = 1
	}
	if v > 1 {
		v = 1
	}

	c := v * s
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

	m := v - c
	return &Color{
		R: (r1 + m) * 255,
		G: (g1 + m) * 255,
		B: (b1 + m) * 255,
		A: a,
	}
}

// ToHSV converts RGB to HSV
func (c *Color) ToHSV() (h, s, v float64) {
	r := c.R / 255.0
	g := c.G / 255.0
	b := c.B / 255.0

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	d := max - min

	// Value is the maximum component
	v = max

	// Saturation
	if max == 0 {
		s = 0
	} else {
		s = d / max
	}

	// Hue
	if max == min {
		h = 0
	} else {
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

	return h, s, v
}

// Greyscale returns the greyscale version of the color
func (c *Color) Greyscale() *Color {
	_, _, l := c.ToHSL()
	return HSLToColor(0, 0, l, c.A)
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

// RGB creates a color from RGB components (0-255)
func RGB(r, g, b string) string {
	rNum := parseChannelNumber(r)
	gNum := parseChannelNumber(g)
	bNum := parseChannelNumber(b)

	color := &Color{rNum, gNum, bNum, 1.0}
	return color.ToHex()
}

// RGBA creates a color from RGBA components (0-255, 0-1)
func RGBA(r, g, b, a string) string {
	rNum := parseChannelNumber(r)
	gNum := parseChannelNumber(g)
	bNum := parseChannelNumber(b)
	aNum := parseAlpha(a)

	color := &Color{rNum, gNum, bNum, aNum}
	return color.ToRGB()
}

// HSL creates a color from HSL components (hue 0-360, saturation 0-100, lightness 0-100)
// Returns in hsl() format, not hex
func HSL(h, s, l string) string {
	hNum := parseNumber(h)
	sNum := parseNumber(s) / 100.0 // Convert from percentage to 0-1
	lNum := parseNumber(l) / 100.0 // Convert from percentage to 0-1

	// Clamp values to valid ranges
	hNum = math.Mod(hNum, 360)
	if hNum < 0 {
		hNum += 360
	}
	sNum = math.Max(0, math.Min(1, sNum))
	lNum = math.Max(0, math.Min(1, lNum))

	// Return in hsl() format
	return fmt.Sprintf("hsl(%g, %g%%, %g%%)", hNum, sNum*100, lNum*100)
}

// HSLA creates a color from HSLA components (hue 0-360, saturation 0-100, lightness 0-100, alpha 0-1)
// Returns in hsla() format
func HSLA(h, s, l, a string) string {
	hNum := parseNumber(h)
	sNum := parseNumber(s) / 100.0 // Convert from percentage to 0-1
	lNum := parseNumber(l) / 100.0 // Convert from percentage to 0-1
	aNum := parseAlpha(a)

	// Clamp values to valid ranges
	hNum = math.Mod(hNum, 360)
	if hNum < 0 {
		hNum += 360
	}
	sNum = math.Max(0, math.Min(1, sNum))
	lNum = math.Max(0, math.Min(1, lNum))
	aNum = math.Max(0, math.Min(1, aNum))

	// Return in hsla() format
	return fmt.Sprintf("hsla(%g, %g%%, %g%%, %g)", hNum, sNum*100, lNum*100, aNum)
}

// Hue extracts the hue component (0-360) from a color
func Hue(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0"
	}
	h, _, _ := color.ToHSL()
	return strconv.FormatFloat(h, 'f', -1, 64)
}

// Saturation extracts the saturation component (0-100) from a color
func Saturation(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0%"
	}
	_, s, _ := color.ToHSL()
	return strconv.FormatFloat(s*100, 'f', -1, 64) + "%"
}

// Lightness extracts the lightness component (0-100) from a color
func Lightness(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0%"
	}
	_, _, l := color.ToHSL()
	return strconv.FormatFloat(l*100, 'f', -1, 64) + "%"
}

// Red extracts the red channel (0-255) from a color
func Red(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(int64(math.Round(color.R)), 10)
}

// Green extracts the green channel (0-255) from a color
func Green(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(int64(math.Round(color.G)), 10)
}

// Blue extracts the blue channel (0-255) from a color
func Blue(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(int64(math.Round(color.B)), 10)
}

// Alpha extracts the alpha channel (0-1) from a color
func Alpha(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "1"
	}
	return strconv.FormatFloat(color.A, 'f', -1, 64)
}

// LumaFunction returns the perceived brightness (luminance) of a color as a string
func LumaFunction(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0"
	}

	lum := color.Luma()
	// Round to 8 decimal places to match LESS output
	return strconv.FormatFloat(math.Round(lum*100000000)/100000000, 'f', -1, 64) + "%"
}

// Luminance calculates the luminance of a color (without gamma correction)
func Luminance(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return "0"
	}

	// Normalize RGB values to 0-1
	r := color.R / 255.0
	g := color.G / 255.0
	b := color.B / 255.0

	// ITU-R BT.709 luminance without gamma correction
	lum := 0.2126*r + 0.7152*g + 0.0722*b
	lumPercent := lum * 100
	// Round to 8 decimal places to match LESS output
	return strconv.FormatFloat(math.Round(lumPercent*100000000)/100000000, 'f', -1, 64) + "%"
}

// Fade sets the opacity of a color (0-100%)
func Fade(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}

	amountNum := parseNumber(amount) / 100.0 // Convert percentage to 0-1
	amountNum = math.Max(0, math.Min(1, amountNum))

	color.A = amountNum
	return formatColor(colorStr, color)
}

// Fadein increases opacity
func Fadein(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}

	amountNum := parseNumber(amount) / 100.0 // Convert percentage
	color.A = math.Min(1, color.A+amountNum)

	return formatColor(colorStr, color)
}

// Fadeout decreases opacity
func Fadeout(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}

	amountNum := parseNumber(amount) / 100.0 // Convert percentage
	color.A = math.Max(0, color.A-amountNum)

	return formatColor(colorStr, color)
}

// Tint mixes a color with white
func Tint(colorStr, weight string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}

	weightNum := parseNumber(weight) / 100.0 // Convert percentage
	white := &Color{255, 255, 255, 1}
	mixed := color.Mix(white, weightNum)

	return mixed.ToHex()
}

// Shade mixes a color with black
func Shade(colorStr, weight string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}

	weightNum := parseNumber(weight) / 100.0 // Convert percentage
	black := &Color{0, 0, 0, 1}
	mixed := color.Mix(black, weightNum)

	return mixed.ToHex()
}

// Contrast returns the dark or light color with greatest contrast
func Contrast(colorStr string, args ...string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}

	dark := &Color{0, 0, 0, 1}
	light := &Color{255, 255, 255, 1}

	// Parse optional arguments
	if len(args) > 0 && args[0] != "" {
		d, err := ParseColor(args[0])
		if err == nil {
			dark = d
		}
	}
	if len(args) > 1 && args[1] != "" {
		l, err := ParseColor(args[1])
		if err == nil {
			light = l
		}
	}

	// Calculate luma of all colors to determine contrast
	colorLuma := color.Luma()
	darkLuma := dark.Luma()
	lightLuma := light.Luma()

	// Return the color with greater contrast
	if math.Abs(darkLuma-colorLuma) > math.Abs(lightLuma-colorLuma) {
		return dark.ToHex()
	}
	return light.ToHex()
}

// Multiply blends two colors using multiply mode
func Multiply(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	result := &Color{
		R: (c1.R / 255.0) * (c2.R / 255.0) * 255,
		G: (c1.G / 255.0) * (c2.G / 255.0) * 255,
		B: (c1.B / 255.0) * (c2.B / 255.0) * 255,
		A: c1.A,
	}
	return result.ToHex()
}

// Screen blends two colors using screen mode
func Screen(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	r := 1.0 - (1.0-(c1.R/255.0))*(1.0-(c2.R/255.0))
	g := 1.0 - (1.0-(c1.G/255.0))*(1.0-(c2.G/255.0))
	b := 1.0 - (1.0-(c1.B/255.0))*(1.0-(c2.B/255.0))

	result := &Color{
		R: r * 255,
		G: g * 255,
		B: b * 255,
		A: c1.A,
	}
	return result.ToHex()
}

// Overlay blends two colors using overlay mode
func Overlay(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	blendChannel := func(a, b float64) float64 {
		a = a / 255.0
		b = b / 255.0
		if a < 0.5 {
			return 2 * a * b
		}
		return 1.0 - 2*(1.0-a)*(1.0-b)
	}

	result := &Color{
		R: blendChannel(c1.R, c2.R) * 255,
		G: blendChannel(c1.G, c2.G) * 255,
		B: blendChannel(c1.B, c2.B) * 255,
		A: c1.A,
	}
	return result.ToHex()
}

// Softlight blends two colors using soft light mode
func Softlight(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	blendChannel := func(a, b float64) float64 {
		a = a / 255.0
		b = b / 255.0
		if b < 0.5 {
			return a - (1-2*b)*a*(1-a)
		}
		var g float64
		if a < 0.25 {
			g = ((16*a-12)*a + 4) * a
		} else {
			g = math.Sqrt(a)
		}
		return a + (2*b-1)*(g-a)
	}

	result := &Color{
		R: blendChannel(c1.R, c2.R) * 255,
		G: blendChannel(c1.G, c2.G) * 255,
		B: blendChannel(c1.B, c2.B) * 255,
		A: c1.A,
	}
	return result.ToHex()
}

// Hardlight blends two colors using hard light mode
func Hardlight(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	blendChannel := func(a, b float64) float64 {
		a = a / 255.0
		b = b / 255.0
		if b < 0.5 {
			return 2 * a * b
		}
		return 1.0 - 2*(1.0-a)*(1.0-b)
	}

	result := &Color{
		R: blendChannel(c1.R, c2.R) * 255,
		G: blendChannel(c1.G, c2.G) * 255,
		B: blendChannel(c1.B, c2.B) * 255,
		A: c1.A,
	}
	return result.ToHex()
}

// Difference blends two colors using difference mode
func Difference(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	result := &Color{
		R: math.Abs(c1.R - c2.R),
		G: math.Abs(c1.G - c2.G),
		B: math.Abs(c1.B - c2.B),
		A: c1.A,
	}
	return result.ToHex()
}

// Exclusion blends two colors using exclusion mode
func Exclusion(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	blendChannel := func(a, b float64) float64 {
		a = a / 255.0
		b = b / 255.0
		return a + b - 2*a*b
	}

	result := &Color{
		R: blendChannel(c1.R, c2.R) * 255,
		G: blendChannel(c1.G, c2.G) * 255,
		B: blendChannel(c1.B, c2.B) * 255,
		A: c1.A,
	}
	return result.ToHex()
}

// Average blends two colors using average mode
func Average(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	result := &Color{
		R: (c1.R + c2.R) / 2,
		G: (c1.G + c2.G) / 2,
		B: (c1.B + c2.B) / 2,
		A: c1.A,
	}
	return result.ToHex()
}

// Negation blends two colors using negation mode
func Negation(color1Str, color2Str string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	result := &Color{
		R: 255 - math.Abs(255-c1.R-c2.R),
		G: 255 - math.Abs(255-c1.G-c2.G),
		B: 255 - math.Abs(255-c1.B-c2.B),
		A: c1.A,
	}
	return result.ToHex()
}

// ColorFunction parses a string as a color
func ColorFunction(colorStr string) string {
	// Remove quotes if present
	colorStr = strings.TrimSpace(colorStr)
	if len(colorStr) >= 2 && ((colorStr[0] == '"' && colorStr[len(colorStr)-1] == '"') ||
		(colorStr[0] == '\'' && colorStr[len(colorStr)-1] == '\'')) {
		colorStr = colorStr[1 : len(colorStr)-1]
	}

	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	return color.ToHex()
}

// Unit removes or changes the unit of a dimension
// unit(value) returns dimensionless number
// unit(value, newUnit) returns the number part of value with the new unit
func Unit(value string, newUnit string) string {
	value = strings.TrimSpace(value)
	newUnit = strings.TrimSpace(newUnit)

	// Extract the numeric part first
	num := parseNumber(value)

	// If no new unit specified, return dimensionless number
	if newUnit == "" {
		// Return just the number without any unit
		if num == math.Floor(num) && num >= -1e15 && num <= 1e15 {
			return strconv.FormatInt(int64(num), 10)
		}

		result := strconv.FormatFloat(num, 'f', -1, 64)
		if strings.Contains(result, ".") {
			result = strings.TrimRight(result, "0")
			result = strings.TrimRight(result, ".")
		}
		return result
	}

	// Remove quotes from newUnit if present
	if len(newUnit) >= 2 && ((newUnit[0] == '"' && newUnit[len(newUnit)-1] == '"') ||
		(newUnit[0] == '\'' && newUnit[len(newUnit)-1] == '\'')) {
		newUnit = newUnit[1 : len(newUnit)-1]
	}

	// Format with new unit
	if num == math.Floor(num) && num >= -1e15 && num <= 1e15 {
		return strconv.FormatInt(int64(num), 10) + newUnit
	}

	result := strconv.FormatFloat(num, 'f', -1, 64)
	if strings.Contains(result, ".") {
		result = strings.TrimRight(result, "0")
		result = strings.TrimRight(result, ".")
	}
	return result + newUnit
}

// GetUnit returns the unit of a dimension as a string (without quotes)
func GetUnit(value string) string {
	value = strings.TrimSpace(value)
	unit := extractUnit(value)
	return unit
}

// Convert converts a number to a different unit
// Supports: px, cm, mm, in, pt, pc, em, ex, ch, rem, vw, vh, vmin, vmax, %
func Convert(value string, targetUnit string) string {
	value = strings.TrimSpace(value)
	targetUnit = strings.TrimSpace(targetUnit)

	// Remove quotes from targetUnit if present
	if len(targetUnit) >= 2 && ((targetUnit[0] == '"' && targetUnit[len(targetUnit)-1] == '"') ||
		(targetUnit[0] == '\'' && targetUnit[len(targetUnit)-1] == '\'')) {
		targetUnit = targetUnit[1 : len(targetUnit)-1]
	}

	// Parse the input number and unit
	num := parseNumber(value)
	sourceUnit := extractUnit(value)

	// Define conversion factors to mm (as base unit)
	conversionToMM := map[string]float64{
		"mm": 1,
		"cm": 10,
		"in": 25.4,
		"pt": 25.4 / 72,
		"pc": 25.4 / 6,
		"px": 0.264583, // Standard web conversion
	}

	// Get conversion factors
	sourceFactor, ok1 := conversionToMM[sourceUnit]
	targetFactor, ok2 := conversionToMM[targetUnit]

	if !ok1 || !ok2 {
		// If we can't convert, just use the Unit function
		return Unit(value, targetUnit)
	}

	// Convert through mm
	valueInMM := num * sourceFactor
	targetValue := valueInMM / targetFactor

	// Format the result
	if targetValue == math.Floor(targetValue) && targetValue >= -1e15 && targetValue <= 1e15 {
		return strconv.FormatInt(int64(targetValue), 10) + targetUnit
	}

	result := strconv.FormatFloat(targetValue, 'f', -1, 64)
	if strings.Contains(result, ".") {
		result = strings.TrimRight(result, "0")
		result = strings.TrimRight(result, ".")
	}
	return result + targetUnit
}

// parseChannelNumber parses a number in the range 0-255
func parseChannelNumber(s string) float64 {
	s = strings.TrimSpace(s)
	num, _ := strconv.ParseFloat(s, 64)
	return math.Max(0, math.Min(255, num))
}

// parseAlpha parses an alpha value (0-1)
func parseAlpha(s string) float64 {
	s = strings.TrimSpace(s)
	num, _ := strconv.ParseFloat(s, 64)
	return math.Max(0, math.Min(1, num))
}

// Public wrapper functions for use by renderer

// roundHSLValue rounds HSL component values to 8 decimal places to avoid floating point artifacts
// This prevents output like "89.99999999999999%" instead of "90%"
func roundHSLValue(val float64) float64 {
	return math.Round(val*100000000) / 100000000
}

// formatColor returns the color in the same format as the input string
func formatColor(colorStr string, result *Color) string {
	switch {
	case strings.HasPrefix(colorStr, "hsla"):
		h, s, l := result.ToHSL()
		s = roundHSLValue(s * 100)
		l = roundHSLValue(l * 100)
		return fmt.Sprintf("hsla(%g, %g%%, %g%%, %g)", h, s, l, result.A)
	case strings.HasPrefix(colorStr, "hsl"):
		h, s, l := result.ToHSL()
		s = roundHSLValue(s * 100)
		l = roundHSLValue(l * 100)
		return fmt.Sprintf("hsl(%g, %g%%, %g%%)", h, s, l)
	case strings.HasPrefix(colorStr, "rgba"):
		return fmt.Sprintf("rgba(%d, %d, %d, %g)", uint8(result.R), uint8(result.G), uint8(result.B), result.A)
	case strings.HasPrefix(colorStr, "rgb"):
		return fmt.Sprintf("rgb(%d, %d, %d)", uint8(result.R), uint8(result.G), uint8(result.B))
	default:
		return result.ToHex()
	}
}

// Lighten lightens a color by a percentage
func Lighten(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Lighten(amountVal)
	return formatColor(colorStr, result)
}

// Darken darkens a color by a percentage
func Darken(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Darken(amountVal)
	return formatColor(colorStr, result)
}

// Saturate increases saturation
func Saturate(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Saturate(amountVal)
	return formatColor(colorStr, result)
}

// Desaturate decreases saturation
func Desaturate(colorStr, amount string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	amountVal := parseNumber(amount) / 100.0
	result := color.Desaturate(amountVal)
	return formatColor(colorStr, result)
}

// Spin rotates the hue
func Spin(colorStr, degrees string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	angleVal := parseNumber(degrees)
	result := color.Spin(angleVal)
	return formatColor(colorStr, result)
}

// Mix mixes two colors
func Mix(color1Str, color2Str string, args ...string) string {
	c1, err1 := ParseColor(color1Str)
	c2, err2 := ParseColor(color2Str)
	if err1 != nil || err2 != nil {
		return color1Str
	}

	weightVal := 0.5
	if len(args) > 0 && args[0] != "" {
		weightVal = parseNumber(args[0]) / 100.0
		if weightVal < 0 {
			weightVal = 0
		}
		if weightVal > 1 {
			weightVal = 1
		}
	}

	result := c1.Mix(c2, weightVal)
	return result.ToHex()
}

// Greyscale returns the greyscale version of the color
func Greyscale(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	result := color.Greyscale()
	return formatColor(colorStr, result)
}

// HSV creates a color from HSV values, returning it as hex
func HSV(h, s, v string) string {
	hVal := parseNumber(h)
	sVal := parseNumber(s) / 100.0
	vVal := parseNumber(v) / 100.0
	result := HSVToColor(hVal, sVal, vVal, 1.0)
	return result.ToHex()
}

// HSVA creates a color from HSVA values, returning it as rgba
func HSVA(h, s, v, a string) string {
	hVal := parseNumber(h)
	sVal := parseNumber(s) / 100.0
	vVal := parseNumber(v) / 100.0
	aVal := parseNumber(a)
	result := HSVToColor(hVal, sVal, vVal, aVal)
	return fmt.Sprintf("rgba(%d, %d, %d, %g)", uint8(math.Round(result.R)), uint8(math.Round(result.G)), uint8(math.Round(result.B)), result.A)
}

// ARGB returns a color in #ARGB format (alpha in first position)
func ARGB(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	a := uint8(math.Round(color.A * 255))
	r := uint8(math.Round(color.R))
	g := uint8(math.Round(color.G))
	b := uint8(math.Round(color.B))
	return fmt.Sprintf("#%02x%02x%02x%02x", a, r, g, b)
}

// HSVHue extracts the hue from a color
func HSVHue(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	h, _, _ := color.ToHSV()
	// Round hue to nearest integer to avoid floating point artifacts
	h = math.Round(h)
	return fmt.Sprintf("%g", h)
}

// HSVSaturation extracts the saturation from a color (HSV saturation)
func HSVSaturation(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	_, s, _ := color.ToHSV()
	return fmt.Sprintf("%g%%", roundHSLValue(s*100))
}

// HSVValue extracts the value (brightness) from a color (HSV value)
func HSVValue(colorStr string) string {
	color, err := ParseColor(colorStr)
	if err != nil {
		return colorStr
	}
	_, _, v := color.ToHSV()
	return fmt.Sprintf("%g%%", roundHSLValue(v*100))
}
