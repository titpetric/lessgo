package dst

import "strings"

// NodeType identifies the node category
type NodeType string

const (
	NodeComment     NodeType = "comment"
	NodeDeclaration NodeType = "declaration" // Rule with selector + properties
	NodeProperty    NodeType = "property"    // property: value pair
	NodeVariable    NodeType = "variable"    // @var: value
	NodeMixin       NodeType = "mixin"       // .mixin() definition
	NodeAtRule      NodeType = "atrule"      // @media, @import, etc
)

// Position tracks source location for errors/source maps
type Position struct {
	Line   int
	Column int
	Offset int
}

// Node is the generic unit in the DST
type Node struct {
	Type NodeType
	Pos  Position

	// Tree structure - all nodes can have children (preserves ordering)
	Parent   *Node
	Children []*Node

	// Primary identifier (selector, @keyword, variable name, mixin name)
	Name string

	// Raw unparsed value (for complex expressions, properties, etc)
	Value string

	// For selectors - supports CSV ".slide h1, .slide h2"
	SelectorsRaw string

	// For conditional mixins - raw "when" expression
	When string

	// Mixin parameters: ["@size", "@color"]
	Params []string

	// Original source text
	Raw string
}

// Document is the root of the tree
type Document struct {
	Nodes []*Node
}

// NewNode creates a new DST node
func NewNode(nodeType NodeType) *Node {
	return &Node{
		Type:     nodeType,
		Children: make([]*Node, 0),
	}
}

// NewNodeWithPosition creates a new DST node with position info
func NewNodeWithPosition(nodeType NodeType, pos Position) *Node {
	return &Node{
		Type:     nodeType,
		Pos:      pos,
		Children: make([]*Node, 0),
	}
}

// AddChild adds a child node and sets parent reference
func (n *Node) AddChild(child *Node) {
	if child == nil {
		return
	}
	child.Parent = n
	n.Children = append(n.Children, child)
}

// Names returns parsed selector list from SelectorsRaw (CSV split)
// e.g., ".slide h1, .slide h2" returns [".slide h1", ".slide h2"]
func (n *Node) Names() []string {
	if n.SelectorsRaw == "" && n.Name != "" {
		return []string{n.Name}
	}

	if n.SelectorsRaw == "" {
		return nil
	}

	selectors := strings.Split(n.SelectorsRaw, ",")
	result := make([]string, 0, len(selectors))
	for _, sel := range selectors {
		trimmed := strings.TrimSpace(sel)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// IsContainer returns true if node can have children
func (n *Node) IsContainer() bool {
	switch n.Type {
	case NodeDeclaration, NodeAtRule, NodeMixin:
		return true
	default:
		return false
	}
}

// NewDocument creates a new document
func NewDocument() *Document {
	return &Document{
		Nodes: make([]*Node, 0),
	}
}

// AddNode adds a top-level node to the document
func (d *Document) AddNode(node *Node) {
	if node != nil {
		d.Nodes = append(d.Nodes, node)
	}
}
