package dst

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/lessgo/internal/strings"
)

func TestParserBlockDetection(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNodes int
		checkNode func(*testing.T, Node)
	}{
		{
			name: "simple selector block",
			input: `.button {
  color: red;
}`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				block, ok := node.(*Block)
				require.True(t, ok, "expected Block, got %T", node)
				require.Equal(t, []string{".button"}, block.SelNames)
			},
		},
		{
			name: "block variable declaration",
			input: `@styles: {
  color: blue;
  font-size: 14px;
};`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				blockVar, ok := node.(*BlockVariable)
				require.True(t, ok, "expected BlockVariable, got %T", node)
				require.Equal(t, "styles", blockVar.Name)
				require.Len(t, blockVar.Children, 2)
			},
		},
		{
			name: "each function call",
			input: `each(range(3), {
  .col-@{value} {
    height: (@value * 50px);
  }
});`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				each, ok := node.(*Each)
				require.True(t, ok, "expected Each, got %T", node)
				require.Equal(t, "range(3)", each.ListExpr)
				require.Equal(t, "value", each.VarName)
				require.Len(t, each.Children, 1)
				block, ok := each.Children[0].(*Block)
				require.True(t, ok, "expected Block child, got %T", each.Children[0])
				require.Equal(t, []string{".col-@{value}"}, block.SelNames)
			},
		},
		{
			name: "function call in declarations",
			input: `.button {
  background: lighten(#333, 10%);
}`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				block, ok := node.(*Block)
				require.True(t, ok, "expected Block, got %T", node)
				require.Equal(t, []string{".button"}, block.SelNames)
				require.Len(t, block.Children, 1)
			},
		},
		{
			name: "mixin call in block",
			input: `.button {
  .mixin();
  color: blue;
}`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				block, ok := node.(*Block)
				require.True(t, ok, "expected Block, got %T", node)
				require.Len(t, block.Children, 2)
				_, isMixin := block.Children[0].(*MixinCall)
				require.True(t, isMixin, "expected MixinCall, got %T", block.Children[0])
			},
		},
		{
			name: "block variable call in block",
			input: `.container {
  @styles();
  padding: 10px;
}`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				block, ok := node.(*Block)
				require.True(t, ok, "expected Block, got %T", node)
				require.Len(t, block.Children, 2)
				decl, isDecl := block.Children[0].(*Decl)
				require.True(t, isDecl, "expected Decl, got %T", block.Children[0])
				require.Equal(t, "@styles", decl.Key)
				require.Equal(t, "()", strings.TrimSpace(decl.Value))
			},
		},
		{
			name: "multiple selectors with nested blocks",
			input: `h1, h2, h3 {
  color: blue;
  &:hover {
    color: red;
  }
}`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				block, ok := node.(*Block)
				require.True(t, ok, "expected Block, got %T", node)
				require.Equal(t, []string{"h1", "h2", "h3"}, block.SelNames)
				require.Len(t, block.Children, 2) // decl + nested block
			},
		},
		{
			name: "variable declaration",
			input: `@primary-color: #0066cc;
@border-radius: 4px;`,
			wantNodes: 2,
			checkNode: func(t *testing.T, node Node) {
				decl, ok := node.(*Decl)
				require.True(t, ok, "expected Decl, got %T", node)
				require.True(t, strings.HasPrefix(decl.Key, "@"))
			},
		},
		{
			name: "nested blocks with parent selector",
			input: `.button {
  background: blue;
  &.active {
    background: red;
  }
}`,
			wantNodes: 1,
			checkNode: func(t *testing.T, node Node) {
				block, ok := node.(*Block)
				require.True(t, ok, "expected Block, got %T", node)
				require.Equal(t, []string{".button"}, block.SelNames)
				require.Len(t, block.Children, 2)
				nestedBlock, ok := block.Children[1].(*Block)
				require.True(t, ok, "expected Block, got %T", block.Children[1])
				require.Equal(t, []string{"&.active"}, nestedBlock.SelNames)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(strings.NewReader(tt.input))
			file, err := parser.Parse()
			require.NoError(t, err, "parse error")
			require.Len(t, file.Nodes, tt.wantNodes, "unexpected number of nodes")

			if tt.checkNode != nil && len(file.Nodes) > 0 {
				tt.checkNode(t, file.Nodes[0])
			}
		})
	}
}
