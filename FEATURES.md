# lessgo Feature Roadmap

## Done (100% Complete - 67/67 Fixtures Passing)

### CLI Commands
- [x] `lessgo generate` - Compile LESS to CSS with full feature support
- [x] `lessgo fmt` - Format LESS files with consistent indentation
- [x] `lessgo ast` - Introspect AST structure for debugging

### Core Features

- [x] 001 - Basic CSS parsing and formatting
- [x] 002 - Variables (`@var: value;`) with block-level scoping via Stack
- [x] 002-scoped - Block-scoped variable overrides
- [x] 003 - Nesting (nested selectors with indentation)
- [x] 004 - Operations (arithmetic: `+`, `-`, `*`, `/` with units)
- [x] 005 - Parent selector (`&`)
- [x] 006 - Color functions (`rgb()`, `rgba()`, hex colors, `hsl()`, `hsla()`)
- [x] 007 - Color manipulation (`lighten`, `darken`)
- [x] 008 - Math functions (`ceil`, `floor`, `round`, `abs`, `sqrt`, `pow`, `min`, `max`)
- [x] 009 - Basic mixins (mixin definitions and invocations)
- [x] 010 - Parametric mixins (mixin parameters and argument passing)
- [x] 011 - Import statements (@import "file.less")
- [x] 011-mixin-guards - Mixin guards with conditions
- [x] 012 - Type functions (isnumber, isstring, iscolor, iskeyword, isurl, ispixel, isem, ispercentage, isunit)
- [x] 014 - Nested media queries (@media blocks bubble to top level)
- [x] 015 - Extend (basic: .class { &:extend(.parent); })
- [x] 016 - Extend (multiple selectors and extends)
- [x] 040 - List functions (`length()`)
- [x] 042 - List functions (`range()`, `extract()`)
- [x] 090 - Color operations (`saturate`, `desaturate`)
- [x] 031 - Logical functions (`boolean()`)
- [x] 032 - String functions (`escape()`)
- [x] 200 - Mixin pattern matching (arity-based overloading)
- [x] 202 - Mixin namespace (#namespace > .mixin() calls)
- [x] 203 - Detached rulesets (@var: { ... } and @var() calls)
- [x] 203 - Mixin override (mixins can override parent declarations)
- [x] 204 - Maps (namespace blocks with variables only)
- [x] 130 - Image functions (`image-width()`, `image-height()`, `image-size()` for local files)

## In Progress


## Planned

### Core Expressions
- [x] 050 - Math functions (basic)
- [x] 051 - Math functions (advanced)
- [x] 052 - Math functions (trigonometric)

### Variables & Imports
- [ ] 011 - Import statements
- [ ] 017 - CSS3 variables

### Functions
- [x] 012 - Type functions
- [x] 060 - Type functions (number)
- [x] 061 - Type functions (color)
- [x] 062 - Type functions (other)
- [x] 063 - Type functions (defined)

### String Operations
- [ ] 013 - Interpolation (basic, needed for each and recursive mixins)
- [x] 032 - String functions (escape)
- [x] 033 - String functions (e)
- [x] 034 - String functions (format)
- [x] 035 - String functions (replace)

### List Functions
- [x] 040 - List functions (length)
- [x] 041 - List functions (extract)
- [x] 042 - List functions (range)
- [x] 043 - List functions (each)

### Logical Functions
- [x] 020 - Luma and if
- [x] 031 - Logical functions (boolean)

### Mixins
- [x] 009 - Basic mixins
- [x] 010 - Parametric mixins
- [x] 011 - Mixin guards
- [x] 200 - Mixin pattern matching
- [ ] 201 - Mixin recursive (requires @{n} interpolation)
- [x] 202 - Mixin namespace

### Color Operations
- [x] 070 - Color definition (RGB)
- [x] 071 - Color definition (HSL)
- [x] 072 - Color definition (HSV)
- [x] 073 - Color definition (ARGB)
- [x] 080 - Color channels (HSL)
- [x] 081 - Color channels (HSV)
- [x] 082 - Color channels (RGB)
- [x] 083 - Color channels (luma)
- [x] 090 - Color operations (saturate)
- [x] 091 - Color operations (lighten)
- [x] 092 - Color operations (fade)
- [x] 093 - Color operations (spin)
- [x] 094 - Color operations (mix)
- [x] 095 - Color operations (greyscale)

### Color Blending
- [x] 100 - Color blending (multiply)
- [x] 101 - Color blending (overlay)
- [x] 102 - Color blending (difference)

### Misc Functions
- [x] 110 - Misc functions (unit)
- [x] 111 - Misc functions (color)

### Advanced
- [x] 018 - Edge cases
- [x] 019 - Comments
- [ ] 203 - Detached rulesets (requires block variable feature)
- [x] 204 - Maps (namespace blocks with variables only)

## Test Statistics

Total fixtures: 67
Passing: 67 (100% pass rate)
Failing: 0
