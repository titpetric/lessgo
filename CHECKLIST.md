# Start of Day Checklist

## Before Starting
```bash
cd /root/github/lessgo
```

## Quick Status Check
```bash
# See what tests are failing
go test ./parser -v

# See line counts
wc -l **/*.go
```

## Today's Tasks (in order)

### ☐ Lexer Fixes (~45 min)
- [ ] Open `parser/lexer.go`
- [ ] **Bug 1:** Fix color detection (lines ~210-220 area)
  - Look at how `-` is handled (line 210)
  - Should check if `-` + digit calls `readNumber()` without the intermediate minus token
- [ ] **Bug 2:** Fix color token (lines ~249, switch statement case '#')
  - Check if digit follows `#` to determine COLOR vs HASH token
  - May need to check `peekAhead(1)` for hex digit
- [ ] **Bug 3:** Fix escape sequences in `readString()` (lines ~260-280)
  - Add map: `\n` → newline, `\t` → tab, `\\` → backslash, etc.
  - Or just handle common cases
- [ ] **Bug 4:** Verify variable parsing with hyphens
  - Run test to confirm it works: `TestLexerBasics/variable`
  - If fails, check loop in `readVariable()`

**Command to verify fixes:**
```bash
go test ./parser -run TestLexer -v
```

### ☐ Parser Implementation (~90 min)
- [ ] Open `parser/parser.go`
- [ ] Review skeleton (it's mostly there)
- [ ] Fix selector parsing - trace through `parseSelector()` 
- [ ] Test selector building in nesting
- [ ] Verify all `parse*()` functions return correct types
- [ ] Run fixture tests:
```bash
go test ./testdata -v
```

### ☐ Renderer Enhancement (~45 min)
- [ ] Open `renderer/renderer.go`
- [ ] Implement variable scope tracking (currently just has flat map)
- [ ] Fix parent selector handling (the `&` symbol)
- [ ] Improve CSS formatting
- [ ] Test with fixtures again

### ☐ Validation (~15 min)
- [ ] All unit tests pass: `go test ./...`
- [ ] At least 2 fixtures work: `go test ./testdata -v`
- [ ] Update PROGRESS.md with what you've done
- [ ] Update FEATURES.md checkboxes for any completed features

## If You Get Stuck

1. Check PROGRESS.md "Next Session Action Plan" - very detailed
2. The bug locations are documented with line numbers
3. All test cases are in `parser/lexer_test.go` - they show expected behavior
4. Fixture files show expected output - use them as reference

## Files to Know

| File | Purpose | Status |
|------|---------|--------|
| `ast/types.go` | AST definitions | ✅ Complete |
| `parser/lexer.go` | Tokenizer | ⚠️ Has 4 bugs |
| `parser/parser.go` | AST builder | 🔨 Partial |
| `renderer/renderer.go` | CSS output | 🔨 Partial |
| `testdata/testdata_test.go` | Fixture tests | ✅ Ready |
| `PROGRESS.md` | Detailed next steps | 📖 Read this! |

## Test Commands

```bash
# Lexer tests only
go test ./parser -run TestLexer -v

# Parser tests (when ready)
go test ./parser -run TestParser -v

# Fixture tests (when parser works)
go test ./testdata -v

# All tests
go test ./...

# With coverage
go test -cover ./...
```

## Documentation References

- **AGENTS.md** - Project guide and common commands
- **PROGRESS.md** - Detailed blockers and action plan ← **READ THIS FIRST**
- **FEATURES.md** - Feature checklist with doc links
- **SUMMARY.md** - Session 1 recap and current state

## Success Criteria for Today

✅ **Minimum:** Lexer tests all pass
✅ **Good:** 2 fixture tests pass (basic-css, variables)
✅ **Excellent:** All 3 fixture tests pass (basic-css, variables, nesting)
✅ **Outstanding:** Start implementing a feature (operations or simple mixins)

## Estimated Time Breakdown

- Lexer fixes: 30-45 min
- Parser impl: 60-90 min  
- Renderer: 30-45 min
- Testing/fixing: 15-30 min
- **Total: 2-3 hours**

## Notes

- The lexer is 90% done - just needs bug fixes
- The parser structure exists but needs completion
- The AST is already comprehensive
- Tests are well-written and will guide your implementation
- No dependencies to fight with - just Go stdlib + testify
