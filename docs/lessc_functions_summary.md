# LESS Functions Reference & Implementation Status

Complete list of LESS.js built-in functions from https://lesscss.org/functions/

## Logical Functions

| Function | Parameters                | Returns | Description                                      | Fixture |
|----------|---------------------------|---------|--------------------------------------------------|---------|
| if       | condition, value1, value2 | any     | Returns value1 if condition is true, else value2 | 030     |
| boolean  | condition                 | boolean | Evaluates a boolean expression                   | 031     |

## String Functions

| Function | Parameters                          | Returns | Description                                               | Fixture |
|----------|-------------------------------------|---------|-----------------------------------------------------------|---------|
| escape   | string                              | string  | URL-encodes special characters                            | 032     |
| e        | string                              | string  | Returns unquoted string (without quotes)                  | 033     |
| % format | format, args...                     | string  | Formats string with placeholders (%s, %d, %a, %S, %D, %A) | 034     |
| replace  | string, pattern, replacement, flags | string  | Replaces pattern in string with replacement               | 035     |

## List Functions

| Function | Parameters       | Returns | Description                                            | Fixture |
|----------|------------------|---------|--------------------------------------------------------|---------|
| length   | list             | integer | Returns number of elements in a list                   | 040     |
| extract  | list, index      | any     | Returns element at specified index (1-indexed)         | 041     |
| range    | start, end, step | list    | Generates list of values from start to end by step     | 042     |
| each     | list, ruleset    | -       | Iterates over list and applies ruleset to each element | 043     |

## Math Functions

| Function   | Parameters       | Returns    | Description                                 | Fixture |
|------------|------------------|------------|---------------------------------------------|---------|
| ceil       | number           | integer    | Rounds up to next highest integer           | 050     |
| floor      | number           | integer    | Rounds down to next lowest integer          | 050     |
| round      | number, decimals | number     | Rounds to nearest integer or decimal places | 050     |
| abs        | number           | number     | Returns absolute value                      | 050     |
| sqrt       | number           | number     | Returns square root                         | 051     |
| pow        | base, exponent   | number     | Returns base raised to exponent             | 051     |
| min        | values...        | number     | Returns minimum value                       | 051     |
| max        | values...        | number     | Returns maximum value                       | 051     |
| percentage | number           | percentage | Converts decimal to percentage              | 051     |
| sin        | number           | number     | Sine function (radians)                     | 052     |
| asin       | number           | number     | Arcsine function                            | 052     |
| cos        | number           | number     | Cosine function (radians)                   | 052     |
| acos       | number           | number     | Arccosine function                          | 052     |
| tan        | number           | number     | Tangent function (radians)                  | 052     |
| atan       | number           | number     | Arctangent function                         | 052     |
| pi         | -                | number     | Returns π (pi) constant                     | 052     |
| mod        | number, divisor  | number     | Returns remainder of division (modulo)      | 052     |

## Type Functions

| Function     | Parameters  | Returns | Description                             | Fixture |
|--------------|-------------|---------|-----------------------------------------|---------|
| isnumber     | value       | boolean | Checks if value is a number             | 060     |
| isstring     | value       | boolean | Checks if value is a string             | 060     |
| iscolor      | value       | boolean | Checks if value is a color              | 061     |
| iskeyword    | value       | boolean | Checks if value is a keyword            | 062     |
| isurl        | value       | boolean | Checks if value is a URL                | 062     |
| ispixel      | value       | boolean | Checks if value has px unit             | 061     |
| isem         | value       | boolean | Checks if value has em unit             | 062     |
| ispercentage | value       | boolean | Checks if value is percentage           | 061     |
| isunit       | value, unit | boolean | Checks if value has specified unit      | 062     |
| isruleset    | value       | boolean | Checks if value is a ruleset            | 063     |
| isdefined    | variable    | boolean | Checks if variable is defined (v4.0.0+) | 063     |

## Color Definition Functions

| Function | Parameters                        | Returns | Description                                          | Fixture |
|----------|-----------------------------------|---------|------------------------------------------------------|---------|
| rgb      | red, green, blue                  | color   | Creates RGB color from 0-255 values                  | 070     |
| rgba     | red, green, blue, alpha           | color   | Creates RGBA color with alpha 0-1                    | 070     |
| argb     | color                             | string  | Returns ARGB hex format (#AARRGGBB)                  | 073     |
| hsl      | hue, saturation, lightness        | color   | Creates color from HSL (hue 0-360, sat/light 0-100%) | 071     |
| hsla     | hue, saturation, lightness, alpha | color   | Creates HSLA color with alpha                        | 071     |
| hsv      | hue, saturation, value            | color   | Creates color from HSV space                         | 072     |
| hsva     | hue, saturation, value, alpha     | color   | Creates HSVA color with alpha                        | 072     |

## Color Channel Functions

| Function      | Parameters | Returns    | Description                                   | Fixture |
|---------------|------------|------------|-----------------------------------------------|---------|
| hue           | color      | integer    | Extracts hue (0-360) from HSL color           | 080     |
| saturation    | color      | percentage | Extracts saturation from HSL color            | 080     |
| lightness     | color      | percentage | Extracts lightness from HSL color             | 080     |
| hsvhue        | color      | integer    | Extracts hue from HSV color                   | 081     |
| hsvsaturation | color      | percentage | Extracts saturation from HSV color            | 081     |
| hsvvalue      | color      | percentage | Extracts value from HSV color                 | 081     |
| red           | color      | integer    | Extracts red channel (0-255)                  | 082     |
| green         | color      | integer    | Extracts green channel (0-255)                | 082     |
| blue          | color      | integer    | Extracts blue channel (0-255)                 | 082     |
| alpha         | color      | number     | Extracts alpha channel (0-1)                  | 082     |
| luma          | color      | percentage | Calculates luma with gamma correction         | 083     |
| luminance     | color      | percentage | Calculates luminance without gamma correction | 083     |

## Color Operation Functions

| Function   | Parameters                    | Returns | Description                                           | Fixture |
|------------|-------------------------------|---------|-------------------------------------------------------|---------|
| saturate   | color, amount, method         | color   | Increases saturation (absolute or relative)           | 090     |
| desaturate | color, amount, method         | color   | Decreases saturation                                  | 090     |
| lighten    | color, amount, method         | color   | Increases lightness                                   | 091     |
| darken     | color, amount, method         | color   | Decreases lightness                                   | 091     |
| fadein     | color, amount, method         | color   | Increases opacity (decreases transparency)            | 092     |
| fadeout    | color, amount, method         | color   | Decreases opacity (increases transparency)            | 092     |
| fade       | color, amount                 | color   | Sets opacity to amount (0-100%)                       | 092     |
| spin       | color, angle                  | color   | Rotates hue angle by degrees                          | 093     |
| mix        | color1, color2, weight        | color   | Mixes two colors with optional weight (0-100%)        | 094     |
| tint       | color, weight                 | color   | Mixes color with white (equivalent to mix with white) | 094     |
| shade      | color, weight                 | color   | Mixes color with black (equivalent to mix with black) | 094     |
| greyscale  | color                         | color   | Removes saturation (desaturate by 100%)               | 095     |
| contrast   | color, dark, light, threshold | color   | Returns dark or light color with greatest contrast    | 095     |

## Color Blending Functions

| Function   | Parameters     | Returns | Description           | Fixture |
|------------|----------------|---------|-----------------------|---------|
| multiply   | color1, color2 | color   | Multiply blend mode   | 100     |
| screen     | color1, color2 | color   | Screen blend mode     | 100     |
| overlay    | color1, color2 | color   | Overlay blend mode    | 101     |
| softlight  | color1, color2 | color   | Soft light blend mode | 101     |
| hardlight  | color1, color2 | color   | Hard light blend mode | 101     |
| difference | color1, color2 | color   | Difference blend mode | 102     |
| exclusion  | color1, color2 | color   | Exclusion blend mode  | 102     |
| average    | color1, color2 | color   | Average blend mode    | 102     |
| negation   | color1, color2 | color   | Negation blend mode   | 102     |

## Misc Functions

| Function     | Parameters                 | Returns   | Description                                         | Fixture |
|--------------|----------------------------|-----------|-----------------------------------------------------|---------|
| color        | string                     | color     | Parses string as color                              | 111     |
| image-size   | string                     | list      | Gets image dimensions (requires file access)        | -       |
| image-width  | string                     | number    | Gets image width (requires file access)             | -       |
| image-height | string                     | number    | Gets image height (requires file access)            | -       |
| convert      | number, unit               | number    | Converts number to different unit                   | 110     |
| data-uri     | mime, string               | string    | Encodes data as URI                                 | -       |
| default      | -                          | boolean   | Returns true if no other mixin matches (guard only) | -       |
| unit         | dimension, unit            | dimension | Removes or changes unit of dimension                | 110     |
| get-unit     | dimension                  | string    | Returns unit of dimension as string                 | 110     |
| svg-gradient | direction, color, color... | gradient  | Creates SVG gradient (complex, file-based)          | -       |

## Test Fixtures Summary

Total fixtures created: **41 pairs** (82 files)

### Fixture Organization by Number Ranges
- **030-031**: Logical Functions (if, boolean)
- **032-035**: String Functions (escape, e, format, replace)
- **040-043**: List Functions (length, extract, range, each)
- **050-052**: Math Functions (basic, advanced, trigonometric)
- **060-063**: Type Functions (number, color, other, defined)
- **070-073**: Color Definition Functions (rgb, hsl, hsv, argb)
- **080-083**: Color Channel Functions (hsl, hsv, rgb, luma)
- **090-095**: Color Operation Functions (saturate, lighten, fade, spin, mix, greyscale)
- **100-102**: Color Blending Functions (multiply, overlay, difference)
- **110-111**: Misc Functions (unit, color)

## Notes

- Fixtures marked with "-" for file access features (image-*, data-uri, svg-gradient) require filesystem/URI handling not typical in unit tests
- All examples follow official LESS documentation at https://lesscss.org/functions/
- Color output formats may vary (hex vs rgba) based on compiler implementation
- Trigonometric functions work in radians; results shown are approximate
- Type check functions return true/false values
- Several functions support optional parameters (method, threshold, etc.)
