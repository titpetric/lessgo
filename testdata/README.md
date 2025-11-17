# Lessgo Code Examples

This document provides working code examples for using Lessgo in your applications.

## Compatibility Statement

Lessgo aims for compatibility with [LESS.js](https://lesscss.org/). The implementation focuses on core features and common use cases. Advanced features and edge cases may differ from the official LESS compiler. See [FEATURES.md](../docs/FEATURES.md) for complete feature status.

## Basic Compilation

Compile LESS source to CSS:

```go
package main

import (
	"fmt"
	"os"

	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
@primary: #333;
body {
  color: @primary;
  font-size: 14px;
}
`

	// Tokenize
	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()

	// Parse
	p := parser.NewParser(tokens)
	stylesheet, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Render
	r := renderer.NewRenderer()
	css := r.Render(stylesheet)
	fmt.Println(css)
}
```

**Output:**
```css
body {
  color: #333;
  font-size: 14px;
}
```

## Variables and Interpolation

Use variables and interpolate them in selectors and values:

```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
@primary: #3498db;
@primary-name: primary;

.color-@{primary-name} {
  background: @primary;
  color: lighten(@primary, 30%);
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

**Output:**
```css
.color-primary {
  background: #3498db;
  color: #81c1e8;
}
```

## Nesting and Parent Selector

Organize styles hierarchically:

```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
.button {
  padding: 10px 20px;
  background: #3498db;

  &:hover {
    background: #2980b9;
  }

  &.active {
    background: #1a5276;
  }
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

**Output:**
```css
.button {
  padding: 10px 20px;
  background: #3498db;
}

.button:hover {
  background: #2980b9;
}

.button.active {
  background: #1a5276;
}
```

## Mixins with Parameters

Create reusable style blocks:

```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
.rounded(@radius: 5px) {
  border-radius: @radius;
  -webkit-border-radius: @radius;
}

.box {
  .rounded(8px);
  background: #ecf0f1;
  padding: 15px;
}

.button {
  .rounded();
  padding: 8px 16px;
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

**Output:**
```css
.box {
  border-radius: 8px;
  -webkit-border-radius: 8px;
  background: #ecf0f1;
  padding: 15px;
}

.button {
  border-radius: 5px;
  -webkit-border-radius: 5px;
  padding: 8px 16px;
}
```

## Arithmetic Operations

Perform calculations with values:

```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
@base-size: 16px;
@grid-width: 1200px;
@columns: 12;

body {
  font-size: @base-size;
}

.column {
  width: (@grid-width / @columns);
}

h1 {
  font-size: @base-size * 2;
  line-height: @base-size * 1.5;
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

**Output:**
```css
body {
  font-size: 16px;
}

.column {
  width: 100px;
}

h1 {
  font-size: 32px;
  line-height: 24px;
}
```

## Color Functions

Manipulate colors dynamically:

```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
@primary: #3498db;

.theme {
  color-base: @primary;
  color-light: lighten(@primary, 25%);
  color-dark: darken(@primary, 25%);
  color-saturated: saturate(@primary, 20%);
  color-gray: greyscale(@primary);
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

**Output:**
```css
.theme {
  color-base: #3498db;
  color-light: #85c8e3;
  color-dark: #1a7fa0;
  color-saturated: #1794d9;
  color-gray: #6b7a81;
}
```

## Using os.DirFS for Imports

Handle LESS imports with filesystem context:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/lessgo/importer"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func compileWithImports(filePath string) error {
	// Read the LESS file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse LESS
	lexer := parser.NewLexer(string(source))
	tokens := lexer.Tokenize()
	p := parser.NewParserWithSource(tokens, string(source))
	stylesheet, err := p.Parse()
	if err != nil {
		return err
	}

	// Resolve imports relative to file's directory
	dir := filepath.Dir(filePath)
	basename := filepath.Base(filePath)
	imp := importer.New(os.DirFS(dir))
	if err := imp.ResolveImports(stylesheet, basename); err != nil {
		return err
	}

	// Render to CSS
	r := renderer.NewRenderer()
	css := r.Render(stylesheet)
	fmt.Println(css)
	return nil
}

func main() {
	if err := compileWithImports("styles/main.less"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

## Math Functions

Use built-in math functions:

```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
.math {
  value1: ceil(3.2);
  value2: floor(3.8);
  value3: round(3.5);
  value4: sqrt(16);
  value5: abs(-10);
  value6: min(5px, 3px);
  value7: max(5px, 3px);
  value8: percentage(0.5);
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

**Output:**
```css
.math {
  value1: 4;
  value2: 3;
  value3: 4;
  value4: 4;
  value5: 10;
  value6: 3px;
  value7: 5px;
  value8: 50%;
}
```

## Type Checking Functions

Check and validate types:

```go
package main

import (
	"fmt"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
.types {
  is-number: isnumber(42);
  is-string: isstring("hello");
  is-color: iscolor(#3498db);
  is-pixel: ispixel(10px);
  is-percentage: ispercentage(50%);
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, _ := p.Parse()
	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```

## Testing Guide

Test fixtures are located in `fixtures/`. Each test case is a pair of files:

- `NNN-name.less` - LESS source
- `NNN-name.css` - Expected CSS output

Run all fixture tests:

```bash
go test ./testdata -v
```

Add a new test case:

```bash
# Create your LESS file
cat > fixtures/101-my-feature.less << 'EOF'
@color: #333;
body { color: @color; }
EOF

# Create expected CSS output
cat > fixtures/101-my-feature.css << 'EOF'
body {
  color: #333;
}
EOF

# Run tests
go test ./testdata -v
```

## Error Handling

Always check for parse errors:

```go
package main

import (
	"fmt"
	"os"

	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	lessSource := `
.invalid {
  property: ;
}
`

	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	r := renderer.NewRenderer()
	fmt.Println(r.Render(stylesheet))
}
```
