# Implementation Progress

## Phase 1: Core Infrastructure (Current)

### Lexer & Parser Foundation
- [x] Project structure and go.mod
- [x] AST type definitions
- [x] Basic lexer with token recognition
  - [x] Whitespace/comment handling (// and /* */)
  - [ ] String and number literals (needs fixes)
  - [ ] Variable references (@var) (needs fixes)
  - [ ] Operators (+, -, *, /, =, >, <, etc.)
  - [x] Selectors and CSS keywords
  - [ ] Color literals (#fff, #ffffff) - needs debugging
  
- [ ] Core parser structure
  - [ ] Rule parsing (selector + declarations)
  - [ ] Variable declarations
  - [ ] Nesting support
  - [ ] Mixin calls

### AST Definitions
- [ ] Node types for all LESS constructs
- [ ] Basic utility functions

### Renderer Foundation
- [ ] CSS output from basic AST
- [ ] Proper formatting and indentation

### Test Infrastructure
- [ ] Fixture test harness
- [ ] Integration test framework
- [ ] Basic test fixtures

## Phase 2: Core Features

### Variables
- [ ] Variable declaration and resolution
- [ ] Scoped variables (lazy evaluation)
- [ ] Variable interpolation in selectors

### Nesting
- [ ] Basic nesting (child selectors)
- [ ] Parent selector (&)
- [ ] Nested at-rules (@media, @supports)

### Mixins
- [ ] Simple mixins (classname mixin calls)
- [ ] Parametric mixins
- [ ] Mixin guards
- [ ] Namespace mixins (#namespace > .mixin)

### Operations
- [ ] Arithmetic operations (+, -, *, /)
- [ ] Color operations
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
- [ ] @import (basic, not full import system)
- [ ] @media and other at-rules
- [ ] Extend/::extend
- [ ] Maps
- [ ] Detached rulesets
- [ ] Plugins (if time permits)

## Known Issues & Blockers

### Critical Lexer Bugs (MUST FIX FIRST)
1. **Color Detection** - `#fff` and `#ffffff` being tokenized as HASH instead of COLOR
   - Location: `parser/lexer.go` line ~249 in `nextToken()` 
   - Issue: Color check at line 201 happens before we handle `#` in switch, needs to check digit after `#`
   - Fix: Move color detection logic or improve the hash handling

2. **Negative Numbers** - `-10` being tokenized as MINUS + NUMBER instead of single NUMBER token
   - Location: `parser/lexer.go` lines ~210-216 (minus handling in switch)
   - Issue: When we see `-`, we check if next char is digit, but `isDigit` might not be working or logic is inverted
   - Fix: Ensure `-` followed by digit immediately calls `readNumber()` without intermediate steps

3. **Variable Token** - Variables with hyphens like `@primary-color` may not be fully captured
   - Location: `parser/lexer.go` lines ~276-286 in `readVariable()`
   - Issue: Already handles hyphens in loop condition, but verify loop is working correctly
   - Status: May be working, verify with tests

4. **String Escape Sequences** - `"\n"` becoming literal `n` instead of newline
   - Location: `parser/lexer.go` lines ~262-271 in `readString()`
   - Issue: Escape handling reads next char but doesn't interpret it (e.g., `\n` should be newline)
   - Fix: Create escape sequence map or handle common escapes: `\n`, `\t`, `\\`, `\"`, `\'`

### Test Status
- Tests created: `parser/lexer_test.go` with 6 test groups
- Failing tests:
  - TestLexerBasics/variable - expects 5 tokens, gets 7 (variable parsing issue)
  - TestLexerStrings/string_with_escapes - escape sequences not interpreted
  - TestLexerNumbers/negative_number - minus sign separated from number
  - TestLexerColors/* - all 3 color tests fail (hash vs color token)

### Fixture Tests Not Yet Run
- 3 fixtures created (basic-css, variables, nesting)
- Can't run testdata_test.go yet because parser/renderer not complete
- Once lexer fixed, will expose parser issues

## Next Session Action Plan

### Step 1: Fix Lexer (30-45 min)
1. [ ] Fix color detection - move before hash switch case or improve logic
2. [ ] Fix negative numbers - ensure `-` + digit path works
3. [ ] Fix escape sequences - add escape map in readString
4. [ ] Run `go test ./parser -v` - all lexer tests should pass
5. [ ] Update PROGRESS.md with results

### Step 2: Implement Parser (1-1.5 hours)
1. [ ] Fix selector parsing - currently has issues with selector building
2. [ ] Implement nesting properly - track parent context
3. [ ] Test with fixture tests: `go test ./testdata -v`
4. [ ] Fix any parser panics/errors
5. [ ] Check PROGRESS.md for parser-specific blockers

### Step 3: Enhance Renderer (45-60 min)
1. [ ] Implement variable scope tracking
2. [ ] Fix parent selector (&) handling in buildSelector
3. [ ] Test with fixture tests
4. [ ] Add better CSS formatting

### Step 4: Feature Implementation (if time permits)
- Start with basic features from FEATURES.md
- Variables and nesting should work after steps 1-3
- Then add: operations, mixins, functions

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
