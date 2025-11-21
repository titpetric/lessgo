package main

import (
	"fmt"
	"github.com/titpetric/lessgo/parser"
)

func main() {
	source := `// Recursive mixins

.generate-classes(@n) when (@n > 0) {
  color: red;
}

.generate-classes(0) {
  color: blue;
}`

	lexer := parser.NewLexer(source)
	tokens := lexer.Tokenize()

	for i, tok := range tokens {
		fmt.Printf("%3d: %v\n", i, tok)
		if i > 20 { break }
	}
}
