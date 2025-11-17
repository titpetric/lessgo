# Lessgo - LESS CSS Compiler in Go

A comprehensive LESS CSS compiler implementation in Go with no external dependencies for core functionality.

## Quick Start

**Install:**
```bash
go install github.com/sourcegraph/lessgo/cmd/lessgo@latest
```

**CLI Usage:**
```bash
lessgo compile styles.less > styles.css
lessgo fmt styles.less
```

**Library Usage:**
```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lexer := parser.NewLexer("@color: #333; body { color: @color; }")
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

See [testdata/README.md](testdata/README.md) for more examples.

## Features

- Variables (`@name: value`)
- Nesting with parent selector (`&`)
- Mixins (simple, parametric, with guards)
- Operations and color functions
- Imports with `@import`
- Extend with `&:extend()`
- Type and math functions

See [docs/features.md](docs/features.md) for complete status.

## Development

```bash
task          # Format and test
task test     # Run tests
task fmt      # Format code
```

## Documentation

- [testdata/README.md](testdata/README.md) - Code examples and testing
- [docs/features.md](docs/features.md) - Feature completeness
- [docs/progress.md](docs/progress.md) - Implementation status
- [AGENTS.md](AGENTS.md) - Developer guide

## License

MIT - See [LICENSE](LICENSE)
