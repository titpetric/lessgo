# Lessgo Agent Guide

A comprehensive LESS CSS compiler implementation in Go with no external dependencies for core functionality.

## Project Structure

```
lessgo/
├── AGENTS.md           # This file
├── PROGRESS.md         # Implementation status and open tasks
├── FEATURES.md         # Supported features checklist
├── go.mod
├── parser/             # LESS lexer and parser
│   ├── lexer.go
│   ├── parser.go
│   └── lexer_test.go
├── ast/                # Abstract Syntax Tree definitions
│   ├── types.go
│   └── ast_test.go
├── renderer/           # CSS output generation
│   ├── renderer.go
│   └── renderer_test.go
├── functions/          # Built-in LESS functions
│   ├── colors.go
│   ├── math.go
│   ├── strings.go
│   └── functions_test.go
├── testdata/           # Fixture tests (*.less and *.css pairs)
│   └── fixtures/
├── integration/        # Integration tests with lesscss
│   ├── integration_test.go
│   └── lesscss_compat.go
└── docs/               # Feature documentation
    └── feat-*.md
```

## Common Commands

### Using Task (recommended)
```bash
task fmt          # Format Go files (goimports + go fmt)
task test         # Run all tests
task test:fixture # Run fixture tests only
task test:lexer   # Run lexer tests only
task build        # Build the project
task              # Default: fmt + test
```

### Manual commands
```bash
go test ./... -timeout 5s           # Run all tests (MUST use -timeout, parser can hang)
go test ./testdata -v -timeout 5s   # Run fixture tests
go test ./parser -v -timeout 5s     # Run lexer tests  
goimports -w . && go fmt ./...      # Format code
```

### Best Practices
- **Always run `task fmt` after modifying .go files** - This ensures consistent formatting
- **Run `task test` before committing** - Verify all tests pass
- **Use `task` for rapid development** - Combines formatting and testing in one command
- **Do NOT modify Taskfile.yml unless explicitly requested** - Keep build tasks stable
- **Always use -timeout flag on tests** - Parser can hang on invalid input

## Development Workflow

1. **Check PROGRESS.md** - Review the current status and next tasks
2. **Implement feature** - Add parser, AST, and renderer support
3. **Create fixtures** - Add test .less and expected .css files
4. **Run tests** - `go test ./...`
5. **Update PROGRESS.md** - Mark completion and note any issues
6. **Update FEATURES.md** - Check off implemented features

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

- See PROGRESS.md for current work and blockers
- See FEATURES.md for complete feature list with doc links
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
