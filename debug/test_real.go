package main

import (
	"fmt"
	"github.com/titpetric/lessgo/parser"
	"github.com/titpetric/lessgo/renderer"
)

func main() {
	src := `.slide h1 {
	color: red;
}

.slide {
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

	r := renderer.NewRenderer()
	output := r.Render(stylesheet)
	fmt.Println(output)
}
