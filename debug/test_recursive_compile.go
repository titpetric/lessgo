package main

import (
	"fmt"
	"os"

	"github.com/titpetric/lessgo/ast"
	"github.com/titpetric/lessgo/parser"
	"github.com/titpetric/lessgo/renderer"
)

func main() {
	source := `// Recursive mixins - mixin calling itself

.generate-classes(@n) when (@n > 0) {
  .class-@{n} {
    width: (10px * @n);
  }
  .generate-classes((@n - 1));
}

.generate-classes(0) {
  // Base case - stop recursion
}

.generate-classes(3);`

	lexer := parser.NewLexer(source)
	tokens := lexer.Tokenize()

	fmt.Println("=== PARSING ===")
	p := parser.NewParserWithSource(tokens, source)
	stylesheet, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed %d rules\n", len(stylesheet.Rules))
	for i, rule := range stylesheet.Rules {
		fmt.Printf("Rule %d: %T", i, rule)
		if r, ok := rule.(*ast.Rule); ok {
			fmt.Printf(" selector=%v params=%v guard=%v nested=%d", r.Selector.Parts, r.Parameters, r.Guard != nil, len(r.Rules))
		}
		if mc, ok := rule.(*ast.MixinCall); ok {
			fmt.Printf(" path=%v args=%d", mc.Path, len(mc.Arguments))
		}
		fmt.Printf("\n")
	}

	fmt.Println("\n=== RENDERING ===")
	r := renderer.NewRenderer()
	css := r.Render(stylesheet)

	fmt.Println(css)
}
