# Lessgo - LESS CSS Compiler in Go

A comprehensive LESS CSS compiler implementation in Go with no external dependencies for core functionality.

## Current Status

**Phase 1: Core Infrastructure** - 60% Complete

- ✅ Project structure and documentation
- ✅ AST type definitions (comprehensive)
- ⚠️ Lexer (functional but 4 bugs need fixing)
- 🔨 Parser (skeleton complete, needs finishing)
- 🔨 Renderer (basic, needs enhancement)
- ✅ Test infrastructure (fixtures ready)

## Quick Start

### View Project Organization
```bash
cat AGENTS.md      # Project guide
cat CHECKLIST.md   # Start of day guide
cat PROGRESS.md    # Detailed next steps
cat FEATURES.md    # Feature checklist
```

### Run Tests
```bash
# See current test status
go test ./parser -v

# Run all tests
go test ./...
```

### What Needs Work
1. Fix 4 lexer bugs (documented in PROGRESS.md)
2. Complete parser implementation
3. Enhance renderer for variables and nesting
4. Test with fixtures

## Project Structure

```
lessgo/
├── AGENTS.md           # Development guide
├── CHECKLIST.md        # Daily checklist
├── PROGRESS.md         # Detailed blockers & next steps
├── FEATURES.md         # Feature checklist with links
├── SUMMARY.md          # Session 1 recap
├── README.md           # This file
│
├── ast/                # Abstract Syntax Tree
│   └── types.go        # Node definitions for all LESS constructs
│
├── parser/             # Lexer and Parser
│   ├── lexer.go        # Tokenizer (has 4 bugs to fix)
│   ├── parser.go       # AST builder (needs completion)
│   └── lexer_test.go   # Comprehensive lexer tests
│
├── renderer/           # CSS output generation
│   └── renderer.go     # AST → CSS conversion (needs enhancement)
│
└── testdata/           # Tests and fixtures
    ├── testdata_test.go     # Fixture test harness
    └── fixtures/            # Test cases (name.less → name.css)
        ├── 001-basic-css.*
        ├── 002-variables.*
        └── 003-nesting.*
```

## Development Workflow

1. **Start Session:** Read CHECKLIST.md
2. **Check Status:** `go test ./...`
3. **Review Next Steps:** Read PROGRESS.md "Next Session Action Plan"
4. **Implement:** Follow the detailed steps
5. **Test:** Run tests after each major change
6. **Document:** Update PROGRESS.md and FEATURES.md

## Features Planned

### Phase 1 (Current)
- [x] CSS Passthrough
- [ ] Variables
- [ ] Nesting
- [ ] Basic Mixins

### Phase 2
- [ ] Operations (+, -, *, /)
- [ ] Mixin Guards
- [ ] Color Functions

### Phase 3
- [ ] String Functions
- [ ] Math Functions
- [ ] Type Functions
- [ ] Import System
- [ ] Extend
- [ ] Maps

See FEATURES.md for complete checklist.

## Code Statistics

- **Total Lines:** ~2200 (code + tests + docs)
- **Core Code:** ~1200 lines (Go)
- **Tests:** ~300 lines
- **Documentation:** ~700 lines
- **External Dependencies:** 0 (for core), testify/require for tests

## Key Design Decisions

1. **No External Dependencies** - Pure Go implementation
2. **Separate Concerns** - Lexer → Parser → AST → Renderer
3. **Fixture-Based Testing** - Easy to add test cases
4. **Comprehensive AST** - Every LESS feature has a node type
5. **Scoped Variables** - Lazy evaluation following LESS semantics

## Quick Reference

### Build & Test
```bash
go build ./...
go test ./...
go test -cover ./...
```

### Common Commands
```bash
# Test lexer
go test ./parser -run TestLexer -v

# Test fixtures
go test ./testdata -v

# See what's not implemented
grep "TODO\|FIXME\|BUG" **/*.go
```

## Next Steps

**Priority 1 (blocks everything):**
1. Fix lexer bugs → `go test ./parser -v` should pass
2. Complete parser → `go test ./testdata -v` should pass
3. Enhance renderer → fixtures should compile to correct CSS

**Priority 2 (implement features):**
- Variables with proper scoping
- Nesting with parent selector (&)
- Mixins (simple first, then parametric)
- Operations and color functions

See PROGRESS.md for detailed action plan with line numbers and specific fixes.

## Documentation

- **AGENTS.md** - Guide for developers, common commands
- **CHECKLIST.md** - Daily start checklist
- **PROGRESS.md** - Issues, blockers, and next action plan ← **Read this first!**
- **FEATURES.md** - Feature completeness checklist
- **SUMMARY.md** - Session 1 recap
- **README.md** - This file

## Testing Strategy

### Unit Tests
- Lexer tests in `parser/lexer_test.go`
- All token types covered
- Tests guide implementation fixes

### Fixture Tests
- Located in `testdata/fixtures/`
- Pair format: `name.less` + `name.css`
- Automatically discovered and tested
- Easy to add: just create a .less and .css file pair

### Integration Tests
- To be added in `integration/` package
- Compare output with actual less.js compiler
- Document compatibility/incompatibility

## Getting Help

1. Check PROGRESS.md "Known Issues & Blockers" section
2. Read detailed "Next Session Action Plan" with line numbers
3. Look at test cases to understand expected behavior
4. Examine fixture .css files for expected output

## License

MIT (placeholder - update as needed)
