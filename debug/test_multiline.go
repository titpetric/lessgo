package main

import (
	"fmt"
	"log"

	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/parser"
)

func main() {
	lessCode := `.box {
  a: isnumber(@num);
  b: isstring(@str);
}`

	// Tokenize
	lexer := parser.NewLexer(lessCode)
	tokens := lexer.Tokenize()

	fmt.Println("=== TOKENS ===")
	for i, tok := range tokens {
		fmt.Printf("%d: %-12s = %q\n", i, tok.Type, tok.Value)
	}

	// Parse with DST
	dstParser := dst.NewParser(tokens, lessCode)
	doc, err := dstParser.Parse()
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// Print DST structure
	fmt.Println("\n=== DST TREE ===")
	printNode(doc.Nodes[0], 0)

	// Render
	fmt.Println("\n=== RENDERED ===")
	r := dst.NewRenderer()
	css := r.Render(doc)
	fmt.Println(css)
}

func printNode(node *dst.Node, indent int) {
	if node == nil {
		return
	}
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	fmt.Printf("%sNode(Type=%s, Name=%q, Value=%q, SelectorsRaw=%q)\n",
		prefix, node.Type, node.Name, node.Value, node.SelectorsRaw)

	for _, child := range node.Children {
		printNode(child, indent+1)
	}
}
