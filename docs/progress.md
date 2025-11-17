# Implementation Progress

## Phase 1: Core Infrastructure (✅ COMPLETE)

### ✅ Lexer & Parser Foundation
- [x] Project structure and go.mod
- [x] AST type definitions
- [x] Complete lexer with token recognition
  - [x] Whitespace/comment handling (// and /* */)
  - [x] String literals with escape sequences (\n, \t, \r, \\, \", \')
  - [x] Number literals with units (10px, 1.5em, -5)
  - [x] Variable references (@var, @primary-color)
  - [x] Operators (+, -, *, /, =, >, <, etc.)
  - [x] Selectors and CSS keywords
  - [x] Color literals (#fff, #ffffff, #rrggbbaa)
  
- [x] Core parser structure
  - [x] Rule parsing (selector + declarations)
  - [x] Variable declarations
  - [x] Nesting support (nested selectors with parent context)
  - [x] Mixin calls (partially - parsed but not rendered)
  - [x] Binary operations (+, -, *, /) with unit support

### ✅ AST Definitions
- [x] Node types for all LESS constructs
- [x] Basic utility functions

### ✅ Renderer Foundation
- [x] CSS output from basic AST
- [x] Proper formatting and indentation
- [x] Arithmetic operations evaluation (10px + 5px = 15px)

### ✅ Test Infrastructure
- [x] Fixture test harness (5s timeout)
- [x] Taskfile.yml with build/test/fmt targets
- [x] 4 passing fixture tests (basic-css, variables, nesting, operations)

## Phase 2: Core Features (Mostly Complete)

### ✅ Variables
- [x] Variable declaration (@var: value)
- [x] Variable resolution in values
- [x] Variable interpolation in selectors (@{var})
- [x] Variable interpolation in property names (@{prop-name})
- [ ] Scoped variables (lazy evaluation)

### ✅ Nesting
- [x] Basic nesting (child selectors)
- [x] Parent selector (&) - implemented and tested
- [x] Nested at-rules (@media with bare declarations) - LESS-style bubbling up

### ✅ Mixins (Parametric Mixins Complete)
- [x] Simple mixins (classname mixin calls) - .mixin() calls now apply mixin declarations
- [x] Parametric mixins (with arguments) - .mixin(@param) definitions and .mixin(value) calls
- [ ] Mixin guards
- [ ] Namespace mixins (#namespace > .mixin)
- **Note**: Parser detects parameters in mixin definitions and binds arguments when called. Parametric mixins are not output to CSS (only regular, non-parametric mixins are). Renderer creates temporary variable scopes for parameter binding.

### ✅ Operations
- [x] Arithmetic operations (+, -, *, /) with unit support
- [x] Color operations (lighten, darken, saturate, desaturate, spin, greyscale)
- [ ] Unit conversions

## Phase 3: Advanced Features

### Color Functions
- [ ] Basic color functions (rgb, rgba, hsl, hsla)
- [ ] Color manipulation (lighten, darken, saturate, desaturate)
- [ ] Color blending functions

### String Functions
- [ ] String manipulation (concatenation, replace)
- [ ] Escaping

### Math Functions
- [ ] Basic math (ceil, floor, round, sqrt, abs)
- [ ] Trigonometric functions (sin, cos, tan)
- [ ] min, max, pow, mod

### Type Functions
- [ ] Type checking (isnumber, isstring, iscolor, etc.)
- [ ] Type conversion functions

### Other Features
- [x] @import (basic - file resolution with fs.FS, error on missing imports, optional imports)
- [ ] @media and other at-rules
- [ ] Extend/::extend
- [ ] Maps
- [ ] Detached rulesets
- [ ] Plugins (if time permits)

## Known Issues & Blockers

### ✅ LEXER FIXED - All 4 Critical Bugs Resolved
1. **Color Detection** - FIXED ✓
   - Moved color check to switch statement for `#` before it returns HASH
   - Now correctly detects 3, 4, 6, and 8 digit hex colors

2. **Negative Numbers** - FIXED ✓
   - Updated readNumber() to accept hasMinusPrefix parameter
   - Properly captures `-` when followed by digit

3. **Variable Token** - FIXED ✓
   - Moved variable check to switch statement for `@` 
   - Now correctly captures full @variable-name tokens

4. **String Escape Sequences** - FIXED ✓
   - Added proper escape sequence handling: \n, \t, \r, \\, \", \'
   - Strings now correctly interpret escape sequences

### Test Status
- ✅ All lexer tests passing (6/6 test groups pass)
- ✅ All fixture tests passing (18/18)
  - 001-basic-css
  - 002-variables
  - 003-nesting
  - 004-operations
  - 005-parent-selector
  - 006-color-functions
  - 007-color-manipulation
  - 008-math-functions
  - 009-basic-mixins
  - 010-parametric-mixins
  - 011-import
  - 011-mixin-guards
  - 012-type-functions
  - 013-interpolation
  - 014-nested-media
  - 015-extend-basic
  - 016-extend-multiple
  - ✅ lessc integration tests: 15/15 passing (100%)
- ✅ Parser handles space-separated and comma-separated values correctly
- ✅ Color manipulation functions working (lighten, darken, etc.)
- ✅ Basic mixin support - declarations from .mixin() calls applied to calling rules
- ✅ Parametric mixin support - arguments bound to parameters in separate variable scope

## Features Added This Session (Current Session - Comment Preservation)

### ✅ Comment Preservation and Rendering
- [x] Created comment extraction system (`parser/comments.go`)
  - Extracts comments from source preserving line/column positions
  - Only extracts top-level comments (not indented/inside blocks)
  - Handles both `//` and `/* */` style comments
  - Only captures comments on their own lines (not mixed with code)
- [x] Extended AST to track comments
  - Added `Comments []*Comment` field to Rule nodes
  - Added `Comments []*Comment` field to VariableDeclaration nodes
- [x] Updated Parser to attach comments to statements
  - Created `NewParserWithSource(tokens, source)` for comment preservation
  - Implemented comment-to-statement attachment logic
  - Comments attached to first statement following them
- [x] Updated Renderer to output comments as CSS
  - Converts `//` comments to `/* */` format
  - Renders comments before associated statements
  - Ensures proper spacing (1 blank line before comment+declaration)
- [x] Updated Formatter to preserve comments in formatted output
  - Comments preserved in formatted LESS files
  - Proper indentation for comments
  - Consistent spacing maintained
- [x] Updated test harness to use comment-aware parser
  - All fixture tests now preserve and render comments
  - Updated fixture CSS files to include rendered comments (4 fixtures changed)
  - **All 22 fixture tests passing with comment support**

## Features Added Previous Session (CSS3 Variables)

### ✅ CSS3 Variables (Custom Properties) Support
- [x] Fixed parser to handle CSS3 `--property-name` syntax (double-dash)
- [x] Updated `parseDeclaration()` to consume TokenMinus for property names
- [x] Fixed `parseFunctionArg()` to prevent infinite loops with `--identifier` patterns
- [x] Added special handling in `parseSimpleValue()` for minus tokens followed by identifiers
- [x] CSS3 variables like `var(--primary-color)` now compile correctly
- [x] Fallback values in var() work: `var(--color, #333)`
- [x] Test fixture for CSS3 variables (017) now passes
- [x] **All 20 fixture tests pass**

### ✅ Developer Quality of Life Improvements
- [x] Fixed timeout issue - parser was hanging on CSS3 variable syntax
- [x] Updated AGENTS.md to enforce `-timeout 5s` on all go test commands
- [x] Updated Taskfile.yml to include testdata tests in main test task
- [x] Fixed extend test fixture (015) to have actual extend syntax
- [x] All tests now complete in <1s with proper timeout handling
- [x] Fixed string literal rendering to preserve quotes
- [x] Added edge case test fixture (018) with attribute selectors, pseudo-classes, nested operators
- [x] **All 21 fixture tests passing**

### Known Issues to Address
- [ ] Comment preservation in formatter (comments dropped during parsing - design issue)
- [ ] Escaped strings in double quotes need lexer fix (`\"` handling in strings)
- [ ] Attribute selector rendering adds spaces around `=` (renderer issue)
- [ ] Selector spacing with combinators (>, +) inconsistent (renderer issue)

## Features Added Previous Session

### ✅ Extend/Inheritance Feature
- [x] Added Extend AST node type for &:extend(.selector) declarations
- [x] Parser support for extend syntax within rules
- [x] Handle both FUNCTION and IDENT tokenization of 'extend'
- [x] Renderer applies extends by merging selectors
- [x] Track all rules and extends for proper selector composition
- [x] Test fixtures for basic and multiple extends (015, 016)
- [x] **All 18 fixture tests pass including new extend tests**
- [x] Extends work with multiple selectors: &:extend(.class1, .class2)

## Features Added Previous Session

### ✅ Nested @media Rules with Bare Declarations Support
- [x] Added DeclarationStmt AST node type to wrap declarations as statements
- [x] Updated parser to handle bare declarations (property: value;) inside @media blocks
- [x] Implemented LESS-style @media bubbling - nested @media queries with bare declarations are hoisted
- [x] Fixed parameter parsing to preserve spacing (e.g., "@media (max-width: 600px)")
- [x] Updated renderer to bubble up @media rules and wrap declarations in parent selectors
- [x] **All 16 fixture tests pass (including new 014-nested-media)**
- [x] **Nested media queries now compile correctly with parent selector wrapping**

### ✅ Type Checking Functions Fix & Integration Test Pass (Previous Session)
- [x] Fixed `iscolor()` to recognize named CSS color keywords (red, blue, green, etc.)
- [x] Fixed `iskeyword()` to treat any unquoted literal/identifier as a keyword
- [x] Fixed `isstring()` to only recognize quoted strings (not unquoted identifiers)
- [x] Fixed `boolean()` to return true ONLY for the literal keyword `true`
- [x] Fixed `length()` to return 1 for quoted strings (single value, not char count)
- [x] Fixed `escape()` function to not escape exclamation marks (matches lessc behavior)
- [x] Added proper AST-based type checking for type functions
- [x] Implemented variable expansion detection for function arguments
- [x] Fixed issue where list variables in function calls weren't expanded properly
- [x] **All 15 fixture tests pass and match lessc output exactly (100%)**
- [x] **All 15 integration tests pass against actual lessc compiler**

### ✅ Stack-Based Variable Scoping Implementation (Previous Session)
- [x] Created `parser/stack.go` - adapted from vuego's stack implementation
- [x] Stack provides LIFO variable scope management (Push/Pop operations)
- [x] Updated Renderer to use Stack instead of flat map for variables
- [x] Implemented proper scope push/pop for mixin parameter binding
- [x] All mixin parameters now live in separate scope layer
- [x] Variable lookups search from top scope to root (following LESS semantics)
- [x] Enables foundation for lazy evaluation and advanced scoping
- [x] Backward compatible - all 13 existing fixture tests still passing

### ✅ Type Checking Functions Implementation (Previous Session)
- [x] Created `functions/types.go` with all type checking functions
- [x] Implemented: isnumber, isstring, iscolor, iskeyword, isurl, ispixel, isem, ispercentage, isunit, isruleset, islist
- [x] Implemented list/string functions: length, extract, range, escape, e, boolean
- [x] Renderer evaluates type checking functions on AST values (preserves type info)
- [x] Variable resolution for type checking (checks type of variable value)
- [x] Parser enhancement: `parseFunctionArg()` handles space-separated values in function arguments
- [x] Fixed parser to allow both space-separated and comma-separated function arguments
- [x] Fixture tests for type checking functions (011 and 012)

## Features Added Previous Session

### ✅ Variable Interpolation Implementation
- [x] Lexer tracks interpolation depth for @{ ... } syntax
- [x] Parser handles interpolation in selectors: `.@{var}`
- [x] Parser handles interpolation in property names: `@{prop-name}`
- [x] Parser handles interpolation in values via Interpolation AST node
- [x] Renderer resolves interpolation by variable lookup
- [x] Both selector and property interpolation working
- [x] Compatible with lessc output
- [x] Fixture test for interpolation

## Features Added Previous Session

### ✅ @import Implementation
- [x] Created `importer` package with fs.FS support
- [x] File resolution relative to importing file
- [x] Error on missing imports (unless optional)
- [x] Support for @import "file.less" syntax
- [x] Support for @import url("file.less") syntax  
- [x] Optional imports: @import "file.less" (optional)
- [x] Nested import resolution (imports within imports)
- [x] Parser recognizes @import, @media, etc. as at-rules
- [x] Both `fmt` and `compile` commands resolve imports
- [x] Fixture test for @import functionality

### ✅ Formatter Implementation (`cmd/lessgo fmt`)
- [x] Created `cmd/lessgo` binary with `fmt` command
- [x] Formatter parses LESS, adds missing semicolons, fixes indentation
- [x] Uses 2-space indentation as standard
- [x] Supports glob patterns for multiple files
- [x] Handles missing semicolons between declarations (intelligent lookahead)
- [x] **FIXED**: Preserves variable references in formatted output (@primary stays @primary)
- [x] **FIXED**: Properly handles nested rule indentation
- [x] **FIXED**: Improved blank line handling between declarations and nested rules
- [x] **FIXED**: Outputs mixin parameters in formatted output
- [x] Now errors on missing imports during formatting

### ✅ Parametric Mixins Implementation
- [x] Added Parameters field to Rule AST node
- [x] Parser detects and parses mixin parameters (.mixin(@param))
- [x] Parser stops selector parsing at LPAREN to avoid consuming parameters
- [x] Renderer binds mixin arguments to parameters in temporary variable scope
- [x] Renderer skips parametric mixin definitions in CSS output
- [x] Created compile command in CLI for testing

### ✅ Parser Improvements
- [x] Made semicolons optional at end of declarations
- [x] Added lookahead for detecting property boundaries (IDENT + COLON pattern)
- [x] Prevents infinite loops when parsing declarations without semicolons
- [x] Added parameter parsing for mixin definitions
- [x] Fixed selector parsing to stop at LPAREN (for mixin parameters)
- [x] Added at-rule keyword detection to distinguish @import/@media from @variables

## Features Added Current Session (Fixture Verification & Fixes)

### ✅ Test Verification Infrastructure
- [x] Created `verify_fixtures.sh` - verifies all fixture .css files against lessc output
- [x] Created `verify_lessgo.sh` - verifies lessgo compilation against fixtures
- [x] All 58 fixtures verified against official lessc compiler (100% match)
- [x] Whitespace normalization in tests (tolerates extra blank lines and trailing spaces)

### ✅ Parser Fixes
- [x] Fixed attribute selector formatting: `[data-test="value"]` (no spaces around =)
- [x] Fixed STRING token handling in selectors - adds quotes back properly
- [x] Fixed spacing rules in needsSpaceBetween() - no space after brackets or operators
- [x] Added support for % format function syntax: `%("format", args)`

### ✅ Renderer Fixes
- [x] Fixed media query indentation - nested rules inside @media now properly indented
- [x] Fixed multi-selector rendering - each selector on separate line with comma-newline separator
- [x] Fixed % format function evaluation - quotes properly handled in result
- [x] Improved Format() function - strips quotes from arguments before substitution

### ✅ Current Test Status
- [x] 32/58 fixtures passing (55% pass rate)
- [x] All basic features working (variables, nesting, operations, mixins, etc.)
- [x] Media query nesting fixed
- [x] Attribute selectors fixed
- [x] % format function working

### 🔴 Remaining Issues (26 failures, 1 timeout)
- [ ] 030-logical-functions-if: Timeout (2s) - infinite loop likely in conditional evaluation
- [ ] 035-string-functions-replace: replace() function not evaluated in variable declarations
- [ ] 040-043: List functions (length, extract, range, each) not evaluated
- [ ] 043-list-functions-each: Parse error on each() syntax
- [ ] 051-052: Advanced math functions (trig) returning literal function calls
- [ ] 063-type-functions-defined: Parse/evaluation error on defined() function
- [ ] 072-073, 080-083: Color definition/channel functions need implementation
- [ ] 090-095: Color blending operations returning literals instead of evaluated results
- [ ] 100-102: Color blending functions (multiply, overlay, difference) not evaluated
- [ ] 110-111: Misc functions (unit, get-unit, convert, color) not evaluated

## Next Session Action Plan

### Priority 1: Function Evaluation Architecture
- [ ] Implement function evaluation in variable declarations (currently only in property values)
- [ ] Add support for all string functions (replace, substring, etc.)
- [ ] Add support for list functions (length, extract, range, each)
- [ ] Fix infinite loop in logical-functions-if test (030 timeout)

### Priority 2: Color Functions
- [ ] Implement color definition functions (rgb, rgba, hsl, hsla, hsv, argb)
- [ ] Implement color channel functions (red, green, blue, hue, saturation, lightness, alpha, luma)
- [ ] Implement color blending functions (multiply, overlay, difference, etc.)

### Priority 3: Math Functions
- [ ] Complete trigonometric functions (sin, cos, tan, asin, acos, atan, pi, etc.)
- [ ] Unit conversion functions (unit, get-unit, convert)
- [ ] Remaining utility functions (color, defined, etc.)

### Priority 4: Advanced Features
- [ ] Fix each() loop syntax and evaluation
- [ ] Detached rulesets as values
- [ ] Maps/objects
- [ ] @plugin system

## Current Code Quality Notes

- Lexer structure is good, just needs token recognition fixes
- Parser skeleton is in place but incomplete
- AST types are comprehensive and well-designed
- Test infrastructure (fixtures) is solid and ready
- No external dependencies (as intended)

## Repository State

- All files committed to `/root/github/lessgo/`
- go.mod configured with testify/require
- 3 fixture pairs ready for testing
- Lexer tests are comprehensive and will guide fixes
