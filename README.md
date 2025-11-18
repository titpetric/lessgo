# Lessgo - LESS CSS Compiler in Go

A comprehensive LESS CSS compiler implementation in Go with no CGO dependencies for core functionality.

## Why?

I want some measure of portability of .less assets to a different runtime. This allows me to avoid the nodejs runtime and dependencies in typical MVC development flow. As support for CSS parsing is non existant in Go (or at least not discovered), this is an attempt to provide an importable package and cli tool aiding compilation.

- https://github.com/tystuyfzand/less-go

This lovely person decided that embedding the javascript source code into a go app and evaluate it using a javascript VM package. While I appreciate the ingenuity (it's not stupid if it works), I did want to address formatting of the source code (`lessgo fmt`).

In the spirit of being efficient, I've started a reimplementation not based on sourcecode, but rather based on test fixtures that cover some required functionality like mixins, functions and everything else.

> **Status**: Currently the project has 3 failing test fixtures of 64. The failures are related to mixins in certain styles without parametrization. Likely some swaths of syntax are unsupported.
>
> Image functions are currently not implemented, there's no particular reason why that is, it's just an isolated thing in the backlog.

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
