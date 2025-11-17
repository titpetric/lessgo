# TODO: Next Session Tasks

## High Priority

### 1. Improve Formatter - Handle Nested Rules Better
- [ ] Formatter currently doesn't handle nested rules properly
- [ ] Need to maintain proper nesting indentation for nested selectors
- [ ] Example: `.parent { .child { color: red; } }` should format with proper indentation
- [ ] Currently renders all nested rules at the same level

### 2. Improve Value Rendering in Formatter  
- [ ] The formatter uses `renderer.RenderValuePublic()` which evaluates variables
- [ ] We want the formatter to preserve the original source form (e.g., keep `@primary` as `@primary`)
- [ ] Solution: Create a separate `formatValue()` method that renders values without evaluation
- [ ] This will preserve variable references in formatted output

### 3. Better Handling of Missing Semicolons
- [ ] Current solution uses `peekAhead(1)` to detect new properties
- [ ] This is fragile and only works because properties use `: ` pattern
- [ ] Better solution: Add NEWLINE tokens to lexer (or at least detect declaration boundaries)
- [ ] Consider how this affects CSS constructs like space-separated values (e.g., `box-shadow: 1px 2px 3px red;`)

### 4. Implement Parametric Mixins
- [ ] Parser needs to support mixin parameters: `.mixin(@param1; @param2) { ... }`
- [ ] Renderer needs to bind arguments to parameters when applying mixins
- [ ] Add test fixture for parametric mixins
- [ ] Handle parameter defaults and guards

## Medium Priority

### 5. Implement Mixin Guards
- [ ] Support `@when` and `@unless` conditions for mixins
- [ ] Example: `.mixin() when (@theme = dark) { ... }`
- [ ] Requires expression evaluation in parser/renderer

### 6. Add More Built-in Functions
- [ ] Type checking: `isnumber()`, `isstring()`, `iscolor()`, `islist()`, etc.
- [ ] String functions: `escape()`, `e()`, `@{interpolation}`
- [ ] Unit functions: `unit()`, `percentage()`
- [ ] Add test fixtures for each function

### 7. Improve Error Messages
- [ ] Add line/column information to parse errors
- [ ] Use Position info from lexer more effectively
- [ ] Create helpful error suggestions

## Lower Priority

### 8. Support for @import
- [ ] Parse `@import` statements
- [ ] Implement file loading (be careful about security)
- [ ] Handle circular imports

### 9. Support for @media and other At-Rules
- [ ] Properly parse `@media` conditions
- [ ] Support nesting inside at-rules
- [ ] Handle other at-rules like `@supports`, `@keyframes`

### 10. Extend / Inheritance
- [ ] Support `&:extend(.class)` syntax
- [ ] Implement selector inheritance

### 11. Detached Rulesets  
- [ ] Support storing rules in variables
- [ ] Example: `@rules: { color: red; }`
- [ ] Calling stored rulesets as mixins

### 12. Maps
- [ ] Support LESS maps/objects
- [ ] Example: `@colors: { primary: #3498db; secondary: #2ecc71; }`
- [ ] Map access and iteration

---

## Notes for Implementation

### Formatter Architecture
The current formatter works but has limitations:
- Uses `renderValue()` from renderer which evaluates variables
- Need to decouple formatting from evaluation
- Consider: `formatValue(value)` that doesn't evaluate, vs `renderValue(value)` that does

### Parser Improvements Needed
- The peekAhead(1) check for COLON is a band-aid solution
- Consider lexer tokenization of newlines for better boundary detection
- Need to be careful not to break CSS like: `box-shadow: 1px 2px 3px red;` (multiple values)

### Testing
- Add unit tests for formatter
- Add integration tests comparing with official lessc
- Test edge cases for missing semicolons
