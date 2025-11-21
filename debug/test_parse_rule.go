package main

import (
	"fmt"
	"github.com/titpetric/lessgo/parser"
)

func main() {
	source := `.generate-classes(@n) when (@n > 0) {
  color: red;
}`

	lexer := parser.NewLexer(source)
	tokens := lexer.Tokenize()

	fmt.Println("=== TOKENS ===")
	for i, tok := range tokens {
		fmt.Printf("%3d: %v\n", i, tok)
	}

	fmt.Println("\n=== PARSING ===")
	p := parser.NewParserWithSource(tokens, source)
	stylesheet, err := p.Parse()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %d rules\n", len(stylesheet.Rules))
}
