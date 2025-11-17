# TODO: Next Session Tasks

## High Priority

### 1. ✅ COMPLETED: Stack-Based Variable Scoping
- [x] Implemented parser/stack.go with Push/Pop scope management
- [x] Updated Renderer to use Stack instead of flat map
- [x] Proper scope management for mixin parameters
- [x] All existing tests passing

### 2. Implement Extend/Inheritance
- [ ] Support `&:extend(.class)` syntax
- [ ] Parse extend in rule declarations
- [ ] Merge selectors from extended classes
- [ ] Example: `.success { &:extend(.message); }` should include .message selector
- [ ] Add test fixture for extend functionality

### 3. Implement Nested At-Rules
- [ ] Support `@media` and other at-rules nested inside rules
- [ ] Bubble at-rules back to stylesheet level
- [ ] Prepend parent selector to rules inside at-rule
- [ ] Example: `.btn { @media (...) { width: 100%; } }` should become `@media (...) { .btn { width: 100%; } }`
- [ ] Add test fixtures for nested @media, @supports, etc.

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
