# Agent Workflow for lessgo (LESS to CSS compiler)

Module: `github.com/titpetric/lessgo`

## Architectural Constraints

### ⚠️ DO NOT MODIFY `dst/` Package Without Explicit Confirmation

The `dst/` package contains the parser and formatter. Before making ANY changes to files in `dst/`:
1. Request explicit confirmation from the user
2. Understand that changes may impact parsing behavior globally
3. Document the change in this AGENTS.md file
4. Test against all fixtures

### Expression Evaluation Strategy
- Use `expression/` package for token parsing and function evaluation
- `evaluator/` package tokenizer is for guard conditions (simple expressions)
- `expression/` package should handle quoted strings and nested function calls properly
- Prefer to solve problems in `renderer/`, `expression/functions/`, and evaluator chains

### Package Consolidation Notes

**Potential Future Cleanup: Merge evaluator/ → expression/**
- The `evaluator/` package only contains a tokenizer for guard conditions
- The `expression/` package is the full expression evaluator with variables & functions
- They solve related problems but at different abstraction levels
- When consolidating, move evaluator tokenizer into expression/ (or deprecate in favor of expression)
- This is NOT urgent - the current separation works fine and maintains clarity

## Development Commands

### Import Management

**Fix missing/unused imports automatically**:

```bash
goimports -w .
```

This should be used instead of manually adjusting imports to ensure consistency.

### Testing & Validation

**Run all tests with coverage** (via task runner):

```bash
task test
```

**Test specific fixture**:

```bash
go test -v -run 'TestFixtures/NNN-description' ./...
```

**Clear test cache before running tests with -count=1**:

```bash
go clean -testcache && go test ./... -count=1
```

Use `-count=1` to disable caching and force fresh test runs. Without clearing the cache first, you'll see "(cached)" in output even with `-count=1`.

**Run benchmarks with multi-CPU profiling**:

```bash
task bench
```

**Generate coverage reports**:

```bash
task cover
```

**Note**: Fixture tests now read pre-computed `.css` files (via lessc) instead of calling lessc at test time. This speeds up tests 280x (15s → 53ms). The .css files are the ground truth and are maintained in version control.

**Render via lessgo (direct CSS output)**:

```bash
./bin/lessgo generate 'testdata/fixtures/FILE.less' > /tmp/lessgo_out.css
```

**Verify identical CSS output**:

```bash
diff /tmp/lessc_out.css /tmp/lessgo_out.css
```

### Formatting (LESS to LESS)

**Format single file** (stdout):

```bash
./bin/lessgo fmt testdata/fixtures/FILE.less
```

**Format in-place**:

```bash
./bin/lessgo fmt -w testdata/fixtures/FILE.less
```

**Format multiple files**:

```bash
./bin/lessgo fmt -w testdata/fixtures/*.less
```

### Compilation (LESS to CSS)

**Generate CSS from single file**:

```bash
./bin/lessgo generate 'testdata/fixtures/FILE.less'
```

**Generate CSS from multiple files**:

```bash
./bin/lessgo generate 'testdata/fixtures/*.less' -o output.css
```

### Inspecting AST

To inspect the AST for a `.less` file:

```bash
go run ./cmd/lessgo ast testdata/fixtures/FILE.less
```

### Building

```bash
go build -o bin/lessgo ./cmd/lessgo
```

### Testing Packages

```bash
go test ./dst ./expr ./fn ./renderer -v
```

## Testing Fixtures

Test fixtures are organized by feature number (e.g., `004-operations.less`, `050-math-functions-basic.less`).

Run against all fixtures:

```bash
for f in testdata/fixtures/*.less; do
  echo "Testing $f..."
  ./bin/lessgo fmt "$f" | lessc - > /tmp/out.css 2>/dev/null
  lessc "$f" > /tmp/expected.css 2>/dev/null
  diff -q /tmp/out.css /tmp/expected.css || echo "FAIL: $f"
done
```

## Development Pattern

1. Add feature to `FEATURES.md` in "In Progress" section
2. Create test fixture in `testdata/fixtures/NNN-description.less`
3. Verify fixture compiles with lessc
4. Implement feature in appropriate package (dst, expr, fn, etc)
5. Validate: `./bin/lessgo fmt fixture.less | lessc - | diff - <(lessc fixture.less)`
6. Move feature to "Done" in FEATURES.md

## Integration Test Runner

Run all fixtures against lessc to identify failures:

**Failfast mode (stop on first failure with diff)**:

```bash
./bin/lessgo fmt fixture.less | lessc - > /tmp/out.css 2>/dev/null
lessc fixture.less > /tmp/expected.css 2>/dev/null
diff /tmp/expected.css /tmp/out.css
```

**Count passing/failing**:

```bash
for f in testdata/fixtures/*.less; do
  ./bin/lessgo fmt "$f" | lessc - > /tmp/out.css 2>/dev/null
  lessc "$f" > /tmp/expected.css 2>/dev/null
  if ! diff -q /tmp/out.css /tmp/expected.css > /dev/null 2>&1; then
    echo "FAIL: $f"
  fi
done
```

## Package Organization

- **dst/**: Data structure tree (parser, formatter, node types)
  - `parser.go`: Parses .less files into DST
  - `formatter.go`: Formats DST back to .less (less->less)
  - `node.go`: Node types (Block, Decl, Comment, MixinCall)
  - `resolver.go`: Variable and expression resolution (used by formatter)
  - `stack.go`: Variable scope stack management

- **renderer/**: CSS code generation (less->css)
  - `renderer.go`: Renders DST to CSS with expression evaluation
  - Handles nested selectors, parent references (&), mixin expansion
  - Evaluates embedded LESS functions in property values

- **expr/**: Expression evaluation (Value, Parse, Evaluator)
  - `parse.go`: Parses numeric values with units
  - `evaluator.go`: Evaluates expressions with variables and functions
  - `color.go`: Color representation and parsing
  - `value.go`: Value type with arithmetic operations

- **expression/functions/**: LessCSS functions (math, color, type, list, image functions)
  - `func_math.go`: ceil, floor, round, abs, sqrt, pow, min, max
  - `func_colors.go`: lighten, darken, saturate, desaturate, rgb, rgba
  - `func_types.go`: isnumber, isstring, iscolor, etc.
  - `func_strings.go`: escape, e, replace, format
  - `func_images.go`: image-width, image-height, image-size (for local files)

- **cmd/lessgo/**: CLI tool
  - `fmt` command: Format .less files for consistent indentation
  - `generate` command: Compile .less files to CSS

## Expression Evaluation

The expr package handles arithmetic with units:

```go
v, _ := expr.Parse("10px")
v2, _ := expr.Parse("5")
result, _ := v.Multiply(v2) // 50px

// Percentages convert to decimals
pct, _ := expr.Parse("50%") // 0.5 (no unit)
```

## Function Organization

Math functions are separated into the `fn` package for independent testing:

```bash
go test ./fn -v          # Test only functions
go test ./expr -v        # Test only expression evaluation
go test ./dst -v         # Test only DST parsing
go test ./... -v         # Test everything
```

## Workflow with lessc

When implementing a feature:

1. Create fixture: `testdata/fixtures/NNN-description.less`
2. Verify with lessc: `lessc testdata/fixtures/NNN-description.less > /tmp/expected.css`
3. Implement feature in appropriate package (dst, expr, fn, renderer)
4. Test with lessgo: `./bin/lessgo generate testdata/fixtures/NNN-description.less > /tmp/actual.css`
5. Verify match: `diff /tmp/expected.css /tmp/actual.css`

## Rendering Context

The `NodeContext` struct holds rendering state including:
- `Buf`: Output buffer
- `Stack`: Variable scope stack
- `BaseDir`: Base directory for resolving relative file paths (e.g., for image functions)

When implementing file-aware features (like image functions):
1. Store context in `NodeContext.BaseDir`
2. Call `renderer.RenderWithBaseDir(astFile, baseDir)` to set the directory
3. Functions can access `functions.BaseDir` global variable
4. The global is set during rendering and automatically propagated through NodeContext

## Image Functions

The `image-width()`, `image-height()`, and `image-size()` functions read local image files to extract dimensions:

- Registered as `image-width`, `image-height`, `image-size` in expression/html_template.go
- Implemented in expression/functions/func_images.go
- Support PNG, JPEG, GIF formats via Go's standard image packages
- Resolve relative paths from the base directory passed to Renderer
- Return unimplemented error for external URLs (http://, https://)
- Cache dimensions to avoid re-reading files

Example usage:

```less
.hero {
  width: image-width('hero.png');
  height: image-height('hero.png');
  background-size: image-size('hero.png');
}
```

## Key Principles

- **DST is data model**: Parse .less as declarative structure, not imperative
- **Lessc is source of truth**: All CSS output must match lessc compilation
- **Separate concerns**:
  - `dst` handles LESS->LESS formatting
  - `renderer` handles LESS->CSS code generation
  - `expr` handles expression evaluation
  - `fn` handles LESS/CSS functions
- **Unit test packages separately**: expr and fn packages should have isolated tests
- **Minimal parser**: Keep parser simple, complexity goes in expr/fn/renderer packages
