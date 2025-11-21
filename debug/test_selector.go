package main

import (
	"fmt"
	"github.com/titpetric/lessgo/parser"
	"github.com/titpetric/lessgo/ast"
)

func main() {
	// Test case: multi-part selector in nested context
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

	fmt.Println("=== AST Structure ===")
	for i, rule := range stylesheet.Rules {
		if r, ok := rule.(*ast.Rule); ok {
			fmt.Printf("Rule %d: selector parts=%v (len=%d)\n", i, r.Selector.Parts, len(r.Selector.Parts))
			for j, nested := range r.Rules {
				if nr, ok := nested.(*ast.Rule); ok {
					fmt.Printf("  Nested %d: selector parts=%v (len=%d)\n", j, nr.Selector.Parts, len(nr.Selector.Parts))
					for k, nested2 := range nr.Rules {
						if nr2, ok := nested2.(*ast.Rule); ok {
							fmt.Printf("    Nested2 %d: selector parts=%v (len=%d)\n", k, nr2.Selector.Parts, len(nr2.Selector.Parts))
						}
					}
				}
			}
		}
	}

	fmt.Println("\n=== Token Stream ===")
	for i, tok := range tokens {
		if tok.Type == "EOF" {
			break
		}
		fmt.Printf("%d: %s = %q\n", i, tok.Type, tok.Value)
	}
}
