package dst

import (
	"fmt"
	"strings"
)

// Print prints the AST structure in a readable hierarchical format
func Print(file *File) {
	if file == nil {
		return
	}
	for _, node := range file.Nodes {
		printNode(node, 0)
	}
}

// printNode recursively prints a node and its children with indentation
func printNode(node Node, depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := node.(type) {
	case *Block:
		fmt.Printf("%sBlock: selectors=[", indent)
		for i, sel := range n.SelNames {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%q", sel)
		}
		fmt.Printf("]")
		if n.IsMixinFunction {
			fmt.Printf(" (mixin)")
		}
		if n.Params != nil && len(n.Params) > 0 {
			fmt.Printf(" params=%v", n.Params)
		}
		if n.Guard != nil && n.Guard.Valid() {
			fmt.Printf(" guard=%q", n.Guard.Condition)
		}
		fmt.Printf("\n")

		for _, child := range n.Children {
			printNode(child, depth+1)
		}

	case *Decl:
		if n.Key[0:1] == "@" && !strings.Contains(n.Key, "{") {
			// Variable assignment
			fmt.Printf("%sVar: %s = %s\n", indent, n.Key, n.Value)
		} else {
			// CSS property
			fmt.Printf("%sDecl: %s = %s\n", indent, n.Key, n.Value)
		}

	case *Comment:
		if n.Multiline {
			fmt.Printf("%sComment: /* %s */\n", indent, n.Text)
		} else {
			fmt.Printf("%sComment: // %s\n", indent, n.Text)
		}

	case *MixinCall:
		fmt.Printf("%sMixinCall: %s(%v)\n", indent, n.Name, n.Args)

	case *Each:
		fmt.Printf("%sEach: list=%q var=%q\n", indent, n.ListExpr, n.VarName)
		for _, child := range n.Children {
			printNode(child, depth+1)
		}

	default:
		fmt.Printf("%s%T\n", indent, node)
	}
}
