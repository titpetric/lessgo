# Complete LESS.js Functions Reference (v4.1.3)

Based on official documentation: https://lesscss.org/functions/

## Quick Reference by Category

### Logical Functions (2 total)

```less
if(condition, value1, value2)           // Returns value1 if condition true, else value2
boolean(condition)                      // Evaluates a boolean expression
```

### String Functions (4 total)

```less
escape(string)                          // URL-encodes special characters
e(string)                               // Returns string without quotes
%(format, arg1, arg2, ...)              // Formats string with placeholders
replace(string, pattern, replacement)   // Replaces pattern in string
```

### List Functions (4 total)

```less
length(list)                            // Returns number of elements
extract(list, index)                    // Gets element at index (1-indexed)
range(start, end, step)                 // Generates list of values
each(list, ruleset)                     // Iterates over list with ruleset
```

### Math Functions (18 total)

```less
// Basic
ceil(number)                            // Round up to next integer
floor(number)                           // Round down to next integer
round(number, decimals)                 // Round to nearest integer
abs(number)                             // Absolute value
percentage(decimal)                     // Convert to percentage

// Advanced
sqrt(number)                            // Square root
pow(base, exponent)                     // Base to exponent power
min(value1, value2, ...)                // Minimum value
max(value1, value2, ...)                // Maximum value
mod(number, divisor)                    // Modulo (remainder)

// Trigonometric (in radians)
sin(radians)                            // Sine
cos(radians)                            // Cosine
tan(radians)                            // Tangent
asin(number)                            // Arcsine
acos(number)                            // Arccosine
atan(number)                            // Arctangent
pi()                                    // π constant (3.14159...)
```

### Type Functions (11 total)

```less
// Value type checks
isnumber(value)                         // Is number?
isstring(value)                         // Is string?
iscolor(value)                          // Is color?
iskeyword(value)                        // Is keyword?
isurl(value)                            // Is URL?

// Unit type checks
ispixel(value)                          // Has px unit?
isem(value)                             // Has em unit?
ispercentage(value)                     // Is percentage?
isunit(value, unit)                     // Has specified unit?

// Advanced checks
isruleset(value)                        // Is ruleset/block?
isdefined(variable)                     // Is variable defined? (v4.0.0+)
```

### Color Definition Functions (7 total)

```less
// RGB color space
rgb(red, green, blue)                   // Create color from RGB (0-255)
rgba(red, green, blue, alpha)           // Create color with alpha (0-1)

// HSL color space
hsl(hue, saturation, lightness)         // Create from HSL
hsla(hue, saturation, lightness, alpha) // Create from HSLA

// HSV color space
hsv(hue, saturation, value)             // Create from HSV
hsva(hue, saturation, value, alpha)     // Create from HSVA

// Format conversion
argb(color)                             // Convert to ARGB hex (#AARRGGBB)
```

### Color Channel Functions (12 total)

```less
// HSL channels
hue(color)                              // Extract hue (0-360)
saturation(color)                       // Extract saturation (0-100%)
lightness(color)                        // Extract lightness (0-100%)

// HSV channels
hsvhue(color)                           // Extract HSV hue (0-360)
hsvsaturation(color)                    // Extract HSV saturation (0-100%)
hsvvalue(color)                         // Extract HSV value (0-100%)

// RGB channels
red(color)                              // Extract red (0-255)
green(color)                            // Extract green (0-255)
blue(color)                             // Extract blue (0-255)
alpha(color)                            // Extract alpha (0-1)

// Luminance
luma(color)                             // Luma with gamma correction
luminance(color)                        // Luminance without gamma
```

### Color Operation Functions (13 total)

```less
// Saturation operations
saturate(color, amount, method)         // Increase saturation
desaturate(color, amount, method)       // Decrease saturation

// Lightness operations
lighten(color, amount, method)          // Increase lightness
darken(color, amount, method)           // Decrease lightness

// Opacity operations
fadein(color, amount, method)           // Increase opacity
fadeout(color, amount, method)          // Decrease opacity
fade(color, amount)                     // Set opacity to amount

// Hue and mixing operations
spin(color, angle)                      // Rotate hue angle
mix(color1, color2, weight)             // Mix two colors
tint(color, weight)                     // Mix with white
shade(color, weight)                    // Mix with black

// Grayscale and contrast
greyscale(color)                        // Remove all saturation
contrast(color, dark, light, threshold) // Return most contrasting color
```

### Color Blending Functions (9 total)

```less
multiply(color1, color2)                // Multiply blend mode
screen(color1, color2)                  // Screen blend mode
overlay(color1, color2)                 // Overlay blend mode
softlight(color1, color2)               // Soft light blend mode
hardlight(color1, color2)               // Hard light blend mode
difference(color1, color2)              // Difference blend mode
exclusion(color1, color2)               // Exclusion blend mode
average(color1, color2)                 // Average blend mode
negation(color1, color2)                // Negation blend mode
```

### Misc Functions (9 total)

```less
// Unit/dimension manipulation
unit(dimension, unit)                   // Remove or change unit
get-unit(dimension)                     // Get unit as string
convert(number, unit)                   // Convert to different unit

// Color parsing
color(string)                           // Parse string as color

// File-based functions (require file access)
image-size(path)                        // Get image dimensions [w, h]
image-width(path)                       // Get image width
image-height(path)                      // Get image height

// Advanced
data-uri(mime, path)                    // Encode file as data URI
svg-gradient(direction, colors...)      // Create SVG gradient

// Guard conditions only
default()                               // True if no other mixin matches
```

---

## Function Parameters & Returns

### Parameter Type Indicators
- `number` - Numeric value (5, 1.5, 10px, 50%)
- `string` - Text value ("hello", 'text')
- `color` - Color object (#fff, rgb(255,0,0), hsl(90,100%,50%))
- `dimension` - Value with unit (10px, 5em, 3%)
- `list` - Comma or space-separated values
- `condition` - Boolean expression (true/false)
- `ruleset` - Block of CSS properties
- `variable` - @variable reference

### Return Type Indicators
- `integer` - Whole number (3, -5, 42)
- `number` - Decimal or whole (3.14, 2, 50.5)
- `color` - Color object
- `string` - Text output
- `percentage` - Percentage value (50%, 100%)
- `boolean` - true or false
- `list` - Multiple values
- `dimension` - Value with unit

---

## Usage Examples by Category

### Logical Functions

```less
@width: 100px;
div { width: if(@width > 50px, @width, 50px); }  // 100px

@light: boolean(luma(#fff) > 50%);  // true
```

### String Functions

```less
@url: "hello world";
div { content: escape(@url); }  // hello%20world

@filter: "alpha(opacity=50)";
div { filter: e(@filter); }  // alpha(opacity=50)

@fmt: %("Value: %s", "test");  // "Value: test"
@replaced: replace("hello", "l", "r");  // "herro"
```

### List Functions

```less
@colors: red, green, blue;
@n: length(@colors);  // 3
@first: extract(@colors, 1);  // red

@range: range(3);  // 1, 2, 3
@range5: range(1px, 5px);  // 1px, 2px, 3px, 4px, 5px

each(@range, { .col-@{value} { width: @value * 10%; } })
```

### Math Functions

```less
@a: ceil(2.4);      // 3
@b: floor(2.6);     // 2
@c: round(2.5);     // 3 or 2 (banker's rounding)
@d: abs(-5);        // 5
@e: percentage(0.5);  // 50%
@f: sqrt(16);       // 4
@g: pow(2, 3);      // 8
@h: min(5, 10, 3);  // 3
@i: max(5, 10, 3);  // 10
@j: mod(5, 2);      // 1
@k: pi();           // 3.14159265359
```

### Type Functions

```less
@n: 42;
@s: "hello";
@c: #ff0000;
@p: 50%;
@e: 1.5em;

isnumber(@n);       // true
isstring(@s);       // true
iscolor(@c);        // true
ispercentage(@p);   // true
isem(@e);           // true
iskeyword(bold);    // true
```

### Color Functions

```less
// Definition
@c1: rgb(255, 128, 0);              // #ff8000
@c2: hsl(90, 100%, 50%);            // #80ff00
@c3: rgba(255, 128, 0, 0.5);        // rgba with transparency

// Channels
@h: hue(@c2);                       // 90
@s: saturation(@c2);                // 100%
@l: lightness(@c2);                 // 50%

// Operations
@lighter: lighten(@c2, 20%);        // #99ff33
@darker: darken(@c2, 20%);          // #669900
@more-sat: saturate(@c2, 10%);      // #99ff00
@less-sat: desaturate(@c2, 10%);    // #7fee00
@mixed: mix(red, blue);             // #800080
@contrast: contrast(#333);          // #ffffff
```

---

## Color Parameters Guide

### HSL Values
- **Hue**: 0-360 degrees (0=red, 120=green, 240=blue)
- **Saturation**: 0-100% (0=gray, 100=pure color)
- **Lightness**: 0-100% (0=black, 50=normal, 100=white)

### HSV Values
- **Hue**: 0-360 degrees
- **Saturation**: 0-100% (0=white, 100=pure color)
- **Value**: 0-100% (0=black, 100=bright)

### RGB Values
- **Red, Green, Blue**: 0-255 each
- **Alpha**: 0-1 (0=transparent, 1=opaque)

### Optional Parameters
- `method`: "relative" for percentage-based changes instead of absolute
- `threshold`: 0-100% bias in contrast() function
- `weight`: 0-100% for mixing proportion

---

## Deprecated or Removed Functions

None in v4.1.3, but note:
- `luminance()` was previously called `luma()` before v1.7.0
- `boolean()` requires v3.0.0+
- `isdefined()` added in v4.0.0
- Trigonometric functions added in v3.7.0

---

## Performance Notes

1. **Color operations** are computed at compile time
2. **Trigonometric functions** use radians, not degrees
3. **Lazy evaluation** means variables in functions are computed when needed
4. **Type functions** return boolean true/false as CSS values
5. **Math functions** preserve units and handle unit conversion

---

## Common Patterns

### Responsive sizing

```less
@base: 16px;
div { font-size: @base; }
h1 { font-size: @base * 1.5; }  // 24px
```

### Color theming

```less
@primary: #0066cc;
@light: lighten(@primary, 20%);
@dark: darken(@primary, 20%);
@hover: saturate(@primary, 10%);
```

### Grid system with ranges

```less
@cols: 12;
@gutter: 20px;
each(range(@cols), {
  .col-@{value} { width: (100% / @cols) * @value - @gutter; }
})
```

### Guard conditions

```less
.mixin(@value) when (isnumber(@value)) {
  width: @value * 1px;
}
.mixin(@value) when (isstring(@value)) {
  content: @value;
}
```
