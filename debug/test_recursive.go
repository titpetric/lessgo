package main

import (
	"fmt"
	"os"

	"github.com/titpetric/lessgo/parser"
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

	fmt.Println("=== TOKENS ===")
	for i, tok := range tokens {
		fmt.Printf("%3d: %v\n", i, tok)
	}

	fmt.Println("\n=== PARSING ===")
	p := parser.NewParserWithSource(tokens, source)
	stylesheet, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed %d rules\n", len(stylesheet.Rules))
}
