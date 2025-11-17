# LESS Features Implemented

## Core Language Features

### Basics
- [ ] CSS Passthrough - [CSS is valid LESS](docs/feat-css-passthrough.md)
- [ ] Comments - `/* */` and `//` style comments

### Variables
- [ ] Variable Declaration (@var: value) - [Variable Basics](docs/feat-variables.md)
- [ ] Variable Interpolation (@{var}, @var in selectors)
- [ ] Variable Variables (@var: @other-var)
- [ ] Lazy Evaluation
- [ ] Properties as Variables ($prop syntax)
- [ ] Default Variables (@var: default-value)

### Nesting
- [ ] Nested Selectors - [Nesting Basics](docs/feat-nesting.md)
- [ ] Parent Selector (&) - [Parent Selectors](docs/feat-parent-selector.md)
- [ ] Multiple & - [Multiple Parent References](docs/feat-multiple-ampersand.md)
- [ ] Nested At-Rules (@media, @supports) - [At-Rule Nesting](docs/feat-nested-at-rules.md)
- [ ] Bubbling - [Selector Bubbling](docs/feat-bubbling.md)

### Mixins
- [ ] Simple Mixins - [Basic Mixins](docs/feat-mixins-basic.md)
- [ ] Mixins with Parentheses (.mixin())
- [ ] Parametric Mixins - [Mixins with Parameters](docs/feat-mixins-parametric.md)
- [ ] Mixin Guards - [Guard Conditions](docs/feat-mixin-guards.md)
- [ ] Pattern Matching - [Pattern Matching in Mixins](docs/feat-pattern-matching.md)
- [ ] Recursive Mixins
- [ ] Namespace Mixins - [Namespaced Mixins](docs/feat-namespaces.md)
- [ ] !important Keyword

### Operations & Math
- [ ] Arithmetic Operations (+, -, *, /) - [Math Operations](docs/feat-operations.md)
- [ ] Unit Conversion
- [ ] Color Operations (color math)
- [ ] calc() Exception

### Functions

#### Color Functions
- [ ] rgb() - [Color Definition](docs/feat-color-rgb.md)
- [ ] rgba()
- [ ] hsl()
- [ ] hsla()
- [ ] hsv()
- [ ] hsva()
- [ ] color() function
- [ ] lighten()
- [ ] darken()
- [ ] saturate()
- [ ] desaturate()
- [ ] spin()
- [ ] mix()
- [ ] tint() / shade()
- [ ] greyscale()
- [ ] contrast()
- [ ] Color blending (multiply, screen, overlay, softlight, hardlight, difference, exclusion, average, negation)

#### String Functions
- [x] escape() - [String Functions](docs/feat-strings.md)
- [x] e() - Quote removal / string escaping
- [x] % (format string)
- [ ] replace()

#### Math Functions
- [ ] ceil() - [Math Functions](docs/feat-math.md)
- [ ] floor()
- [ ] round()
- [ ] sqrt()
- [ ] abs()
- [ ] sin()
- [ ] cos()
- [ ] tan()
- [ ] asin()
- [ ] acos()
- [ ] atan()
- [ ] pi()
- [ ] pow()
- [ ] mod()
- [ ] min()
- [ ] max()
- [ ] percentage()

#### List Functions
- [x] length()
- [x] extract()
- [x] range()
- [ ] each()

#### Type Functions
- [x] isnumber()
- [x] isstring()
- [x] iscolor()
- [x] iskeyword()
- [x] isurl()
- [x] ispixel()
- [x] isem()
- [x] ispercentage()
- [x] isunit()
- [x] isruleset()
- [x] islist()
- [ ] isdefined() (requires scope tracking)

#### Logical Functions
- [ ] if()
- [x] boolean()

### Advanced Features
- [ ] @import - [Import System](docs/feat-imports.md)
  - [ ] .less files
  - [ ] .css files
  - [ ] Import options (reference, inline, less, css, once, multiple, optional)
- [ ] Extend - [Selector Extension](docs/feat-extend.md)
- [ ] Maps - [Using Maps](docs/feat-maps.md)
- [ ] Detached Rulesets
- [ ] CSS Guards - [Conditional Selectors](docs/feat-css-guards.md)
- [ ] Scope & Visibility - [Variable Scope](docs/feat-scope.md)
- [ ] @plugin - [Plugin System](docs/feat-plugins.md) (if implemented)

## Syntax Support

- [ ] CSS Selectors (all CSS 3 selectors)
- [ ] @media queries
- [ ] @supports queries
- [ ] @keyframes
- [ ] @font-face
- [ ] @namespace
- [ ] All pseudo-classes and pseudo-elements
- [ ] Attribute selectors

## Known Limitations / Not Implemented

- [ ] File I/O and import path resolution (integration responsibility)
- [ ] @plugin system (advanced feature, may defer)
- [ ] Some edge cases in variable interpolation
- [ ] Real-time browser compilation (JS only)
- [ ] Source maps (can be added later)

## Testing Status

- Integration tests pending (requires lessc comparison)
- See PROGRESS.md for implementation timeline
