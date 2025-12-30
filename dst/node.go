package dst

// Node represents a single element in a .less file
type Node interface {
	// Names returns the selectors/names for this node (e.g., ".class", "h1", "&.modifier", "p b")
	Names() []string
	// Type returns the node type
	Type() NodeType
}

// NodeType indicates the kind of node
type NodeType string

const (
	TypeDecl          NodeType = "decl"           // Value declaration (key: value;)
	TypeComment       NodeType = "comment"        // Single or multi-line comment
	TypeBlock         NodeType = "block"          // Nested selector block with children
	TypeMixinCall     NodeType = "mixin_call"     // Mixin invocation (.mixin();)
	TypeBlockVariable NodeType = "block_variable" // Block variable (@var: { ... };)
	TypeEach          NodeType = "each"           // Each loop (each(list, { ... });)
	TypeImport        NodeType = "import"         // CSS @import passthrough
)

// Decl represents a CSS declaration (property: value;)
type Decl struct {
	SelNames []string // selectors (e.g., ".class", "h1 span")
	Key      string   // property name (e.g., "color")
	Value    string   // property value (e.g., "#000")
}

func (d *Decl) Names() []string { return d.SelNames }
func (d *Decl) Type() NodeType  { return TypeDecl }

// Comment represents a single or multi-line comment
type Comment struct {
	Text      string // comment text without // or /* */
	Multiline bool   // true if /* */ style, false if // style
}

func (c *Comment) Names() []string { return nil }
func (c *Comment) Type() NodeType  { return TypeComment }

type Guard struct {
	Condition string // Raw condition string, e.g., "@theme = dark"
}

func (g *Guard) Valid() bool {
	return g != nil && g.Condition != ""
}

// Block represents a selector with nested nodes
type Block struct {
	SelNames        []string // selectors (e.g., ".class", "&.modifier", "h1 span")
	IsMixinFunction bool
	Children        []Node   // nested declarations and blocks
	Parent          *Block   // parent block for & resolution
	Params          []string // mixin parameters (e.g., "@v", "@color")
	Guard           *Guard   // Guard conditions for mixin
}

func (b *Block) Names() []string {
	return b.SelNames
}

func (b *Block) Type() NodeType { return TypeBlock }

// MixinCall represents a mixin invocation (.mixin(); or .mixin(args);)
type MixinCall struct {
	Name string   // mixin name (e.g., ".mixin")
	Args []string // arguments (e.g., ["10px"], ["@color", "blue"])
}

func (m *MixinCall) Names() []string { return nil }
func (m *MixinCall) Type() NodeType  { return TypeMixinCall }

// BlockVariable represents a block variable assignment (@var: { ... };)
type BlockVariable struct {
	Name     string // variable name (e.g., "styles")
	Children []Node // declarations and blocks in the variable
}

func (bv *BlockVariable) Names() []string { return nil }
func (bv *BlockVariable) Type() NodeType  { return TypeBlockVariable }

// Each represents an each() loop (each(range(3), { ... });)
type Each struct {
	ListExpr string // the list expression (e.g., "range(3)")
	VarName  string // loop variable name (e.g., "value" for @value)
	Children []Node // declarations and blocks to repeat
}

func (e *Each) Names() []string { return nil }
func (e *Each) Type() NodeType  { return TypeEach }

// Import represents a CSS @import statement that should pass through to output
type Import struct {
	Path string // the import path (URL or file reference)
}

func (i *Import) Names() []string { return nil }
func (i *Import) Type() NodeType  { return TypeImport }

// File represents the entire parsed .less file
type File struct {
	Nodes []Node
}
