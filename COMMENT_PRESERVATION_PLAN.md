# Comment Preservation and Spacing Improvements

## Problem Statement
Currently, lessgo drops all comments during parsing. This affects both:
1. **CSS Output** - Comments are lost in compiled CSS, reducing code documentation
2. **Formatter** - The formatter cannot preserve comments, making it unsuitable for development

## Requirements
1. **Preserve comments** in source during parsing
2. **Render comments to CSS output** (convert `//` to `/* */`)
3. **Formatter preserves comments** in formatted output
4. **Spacing consistency** - blank lines before comment+statement blocks

## Challenge: Design Complexity
Adding comment support requires significant changes to the lexer/parser:
- Current lexer skips comments in `skipWhitespaceAndComments()`
- Parser has no AST node type for comments
- Adding TokenComment tokens causes parser to hang (not expecting comment tokens)
- Integrating comments into existing AST requires careful design

## Two Possible Approaches

### Approach A: Reparse for Comments (Quick, Limited)
1. Keep lexer/parser as-is (skips comments)
2. Add separate comment extraction pass:
   - Scan source for `//` and `/* */` comments
   - Extract comment text and position
   - Store comments in parallel structure
3. During rendering:
   - Merge comments back with statements using position info
   - Convert `//` to `/* */` for CSS output
4. **Pros**: Minimal lexer/parser changes
5. **Cons**: Comments at EOF/edge positions may be lost, more complex logic

### Approach B: Comment Tokens (Clean, Complex)
1. Refactor lexer to return comment tokens alongside other tokens
2. Modify parser to handle comment tokens specially:
   - Skip comment tokens but track them
   - Associate comments with following statements
3. Add comments to AST node types where needed
4. Renderer outputs comments
5. **Pros**: Clean, follows normal lexer/parser design
6. **Cons**: Requires significant parser refactoring, risk of hangs

## Recommendation: Defer to Future
Given the 5-second timeout budget and complexity, comment preservation should be deferred:
1. Current system works well for production CSS output
2. Users can use original LESS files for comments
3. Focus on CSS output quality (spacing, compatibility) instead
4. Document as known limitation

## Current Spacing Issues (High Priority)
These can be fixed without comment support:
1. **Attribute selector spacing** - `[data-test = value]` should be `[data-test="value"]`
2. **Selector combinator spacing** - Should preserve spaces around `>`, `+`, `~`
3. **Blank lines** - Add consistent spacing between rules

## Phase 1: Spacing Fixes (This Session)
- [ ] Fix attribute selector rendering (no spaces)
- [ ] Fix selector combinator spacing
- [ ] Add blank line before rules
- [ ] Test with edge case fixtures

## Phase 2: Comment Support (Future, Major Effort)
- [ ] Evaluate both approaches
- [ ] Prototype chosen approach
- [ ] Full test coverage
- [ ] Performance testing
