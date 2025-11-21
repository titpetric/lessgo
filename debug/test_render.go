package main

import (
	"fmt"
	"github.com/titpetric/lessgo/parser"
	"github.com/titpetric/lessgo/renderer"
	"github.com/titpetric/lessgo/ast"
)

func main() {
	src := `.slide {
	h1, h2, h3 {
		a {
			background: none;
		}
	}
}`

	lexer := parser.NewLexer(src)
	tokens := lexer.Tokenize()
	p := parser.NewParserWithSource(tokens, src)
	stylesheet, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse Error: %v\n", err)
		return
	}

	fmt.Println("=== Selector Parts ===")
	for i, rule := range stylesheet.Rules {
		if r, ok := rule.(*ast.Rule); ok {
			fmt.Printf("Rule %d: parts=%v (len=%d)\n", i, r.Selector.Parts, len(r.Selector.Parts))
			for j, nested := range r.Rules {
				if nr, ok := nested.(*ast.Rule); ok {
					fmt.Printf("  Nested %d: parts=%v (len=%d)\n", j, nr.Selector.Parts, len(nr.Selector.Parts))
				}
			}
		}
	}

	r := renderer.NewRenderer()
	output := r.Render(stylesheet)
	fmt.Println("\n=== Output ===")
	fmt.Println(output)
}
