# Lessgo Agent Guide

A comprehensive LESS CSS compiler implementation in Go with no external dependencies for core functionality.

## Project Structure

```
lessgo/
├── AGENTS.md                   # This file (developer guide)
├── README.md                   # Quick start and overview
├── LICENSE                     # MIT License
├── go.mod
├── go.sum
│
├── cmd/lessgo/                 # CLI tool
│   ├── main.go
│   └── README.md
│
├── parser/                     # Lexer and parser
│   ├── lexer.go
│   ├── parser.go
│   ├── dst.go                  # Document Structure Tree (simpler alternative to AST)
│   └── lexer_test.go
│
├── renderer/                   # CSS output generation
│   └── renderer.go
│
├── evaluator/                  # Expression evaluation
│   └── evaluator.go
│
├── functions/                  # Built-in LESS functions
│   ├── colors.go
│   ├── math.go
│   ├── strings.go
│   └── functions_test.go
│
├── importer/                   # Import resolution
│   └── importer.go
│
├── formatter/                  # LESS code formatting
│   └── formatter.go
│
├── testdata/                   # Fixture tests
│   ├── README.md              # Code examples
│   ├── testdata_test.go       # Fixture test runner
│   └── fixtures/              # Test cases (*.less and *.css pairs)
│
└── docs/                       # Documentation
    ├── features.md            # Feature status
    ├── progress.md            # Implementation status
    └── *.md                   # Other reference docs
```

## Architecture Refactoring: From AST to DST

**Current Status**: Migrating from complex AST to simpler Document Structure Tree (DST)

**DST (Document Structure Tree)** is a simpler, entity-focused alternative to a full Abstract Syntax Tree:
- Generic `Node` structure: `Type`, `Name`, `Value`, `Children`, `Parent`
- Simpler node types: "stylesheet", "rule", "declaration", "variable", "atrule", etc
- Direct parent-child relationships for DOM-like traversal
- Easier to manipulate and debug than complex typed AST nodes

**Benefits of DST**:
- Reduced complexity in tree manipulation
- Better for generic traversal (all nodes are same type)
- Simpler code generation and transformation
- More flexible for LESS-specific features (nested rules, variables)

## Debug Scripts

Create test programs in the `debug/` folder for reusable testing:
```bash
go run ./debug/test_selector.go
go run ./debug/test_parse.go
```

This keeps the codebase clean and allows iterating on test cases.

## Common Commands

### Using Task (recommended)

```bash
task fmt          # Format Go files (goimports + go fmt)
task test         # Run all tests (5s budget for test, 5s for compile)
task test:fixture # Run fixture tests only (5s timeout)
task test:lexer   # Run lexer tests only (5s timeout)
task build        # Build the project
task              # Default: fmt + test
```

### Manual commands

```bash
# Always use -timeout 5s - this is our strict budget
go test ./... -timeout 5s           # Run all tests
go test ./testdata -v -timeout 5s   # Run fixture tests
go test ./parser -v -timeout 5s     # Run lexer tests  
timeout 5s go run ./cmd/lessgo compile <file>.less  # Compile with 5s timeout
timeout 5s go run ./cmd/lessgo fmt <file>.less      # Format with 5s timeout
goimports -w . && go fmt ./...      # Format code
```

### Best Practices
- **Always run `task fmt` after modifying .go files** - This ensures consistent formatting
- **Run `task test` before committing** - Verify all tests pass
- **Use `task` for rapid development** - Combines formatting and testing in one command
- **Do NOT modify Taskfile.yml unless explicitly requested** - Keep build tasks stable
- **Always use -timeout flag on tests** - Parser can hang on invalid input

## Development Workflow

1. **Check docs/progress.md** - Review the current status and next tasks
2. **Implement feature** - Add parser, AST, and renderer support
3. **Create fixtures** - Add test .less and expected .css files
4. **Run tests** - `go test ./...`
5. **Update docs/progress.md** - Mark completion and note any issues
6. **Update docs/features.md** - Check off implemented features

## Testing Strategy

### Fixture Tests (Preferred for Agent Work)
- Located in `testdata/fixtures/`
- File pairs: `name.less` and `name.css`
- Use `./test_fixtures_vs_lessc.sh` to validate against official lessc
- Each fixture name is a test case

**Recommended workflow for fixing failing tests:**
```bash
# 1. Run tests with prefix to focus on specific failures
./test_fixtures_vs_lessc.sh 999          # Test only 999-* fixtures
./test_fixtures_vs_lessc.sh 200-         # Test only 200-* fixtures

# 2. Compare lessc vs lessgo output for a specific test
lessc testdata/fixtures/999-sinog-index.less                    # Official output
./bin/lessgo compile testdata/fixtures/999-sinog-index.less     # Our output

# 3. Use diff to see exact differences
diff -u <(lessc testdata/fixtures/999-sinog-index.less) <(./bin/lessgo compile testdata/fixtures/999-sinog-index.less)

# 4. After fixes, re-run with same prefix to verify
./test_fixtures_vs_lessc.sh 999
```

### Integration Tests
- Compare output with actual LESS compiler (lessc)
- Document feature compatibility/incompatibility
- Use the test script above for validation

### Unit Tests
- Direct testing of parser, AST, and renderer
- Use `testify/require` for assertions
- Test edge cases and error conditions

## Key Implementation Notes

- **No external dependencies** for core functionality
- **DOM/AST-based** - Parse to tree, manipulate, render to CSS
- **Separate concerns** - Lexer → Parser → AST → Renderer
- **Variable resolution** - Scoped variable lookups following LESS semantics
- **Lazy evaluation** - Variables evaluated when needed, not at parse time

## References

- See docs/progress.md for current work and blockers
- See docs/features.md for complete feature list
- LESS official docs: https://lesscss.org/features/

## Integration Points

When working on integration with the actual LESS compiler, use:
- **lessc command-line tool** - Available as `lessc` in the environment
- **less command** - LESS CLI available for testing
- Node.js less package (if needed)
- Docker container with less compiler
- Online playground: https://lesscss.org/less-preview

### Using lessc for validation

```bash
lessc testdata/fixtures/001-basic-css.less # Compile with official LESS
lessc testdata/fixtures/001-basic-css.less > /tmp/official.css
# Compare against our output
```
