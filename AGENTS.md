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
│   └── lexer_test.go
│
├── ast/                        # Abstract Syntax Tree
│   ├── types.go
│   └── ast_test.go
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

### Fixture Tests
- Located in `testdata/fixtures/`
- File pairs: `name.less` and `name.css`
- Automatically discovered and tested
- Each fixture name is a test case

### Integration Tests
- Compare output with actual LESS compiler
- Document feature compatibility/incompatibility
- Located in `integration/` package

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
