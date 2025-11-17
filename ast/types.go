package ast

import "fmt"

// Node is the base interface for all AST nodes
type Node interface {
	node()
}

// Stylesheet is the root node containing all rules and statements
type Stylesheet struct {
	Rules []Statement
}

func (s *Stylesheet) node() {}

// Statement represents a top-level CSS statement
type Statement interface {
	Node
	stmt()
}

// Rule represents a CSS rule with a selector and declarations
type Rule struct {
	Selector     Selector
	Declarations []Declaration
	Rules        []Statement // nested rules
	Position     Position
}

func (r *Rule) node() {}
func (r *Rule) stmt() {}

// Selector represents a CSS selector or multiple selectors
type Selector struct {
	// Can contain multiple selectors separated by commas
	// For now, store as string for simplicity
	Parts []string
}

// Declaration represents a property: value pair
type Declaration struct {
	Property string
	Value    Value
}

// Value represents the value side of a declaration
type Value interface {
	Node
	value()
}

// Literal represents a literal CSS value (color, number, string, keyword)
type Literal struct {
	Type  LiteralType
	Value string
}

func (l *Literal) node()  {}
func (l *Literal) value() {}

type LiteralType string

const (
	ColorLiteral      LiteralType = "color"
	NumberLiteral     LiteralType = "number"
	StringLiteral     LiteralType = "string"
	KeywordLiteral    LiteralType = "keyword"
	URLLiteral        LiteralType = "url"
	UnitLiteral       LiteralType = "unit"
	PercentageLiteral LiteralType = "percentage"
)

// Variable represents a variable reference (@varname)
type Variable struct {
	Name string
}

func (v *Variable) node()  {}
func (v *Variable) value() {}

// FunctionCall represents a function invocation
type FunctionCall struct {
	Name      string
	Arguments []Value
}

func (f *FunctionCall) node()  {}
func (f *FunctionCall) value() {}

// BinaryOp represents a binary operation like +, -, *, /
type BinaryOp struct {
	Left     Value
	Operator string
	Right    Value
}

func (b *BinaryOp) node()  {}
func (b *BinaryOp) value() {}

// MixinCall represents a call to a mixin (.classname() or #namespace.mixin())
type MixinCall struct {
	Path      []string // for namespace support: [namespace, mixin] or just [mixin]
	Arguments []Value
	Important bool
}

func (m *MixinCall) node()  {}
func (m *MixinCall) stmt() {}

// VariableDeclaration represents @variable: value
type VariableDeclaration struct {
	Name  string
	Value Value
}

func (v *VariableDeclaration) node()  {}
func (v *VariableDeclaration) stmt() {}

// AtRule represents @-rules like @media, @import, @keyframes, etc.
type AtRule struct {
	Name       string // "media", "import", "keyframes", etc.
	Parameters string // the part after the @ (e.g., "(min-width: 768px)")
	Block      interface{}
	// Block can be:
	// - []Statement (for @media, @supports)
	// - string (for @import)
	// - Keyframes (for @keyframes)
	Position Position
}

func (a *AtRule) node()  {}
func (a *AtRule) stmt() {}

// List represents comma or space-separated values
type List struct {
	Values    []Value
	Separator string // "," or " "
}

func (l *List) node()  {}
func (l *List) value() {}

// Interpolation represents #{ expression } or @{ variable } syntax
type Interpolation struct {
	Expression Value
}

func (i *Interpolation) node()  {}
func (i *Interpolation) value() {}

// Position tracks location in source for error reporting
type Position struct {
	Line   int
	Column int
	Offset int
}

// String implements fmt.Stringer for Position
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Utility functions

// NewStylesheet creates a new stylesheet
func NewStylesheet() *Stylesheet {
	return &Stylesheet{Rules: []Statement{}}
}

// AddRule adds a rule to the stylesheet
func (s *Stylesheet) AddRule(rule Statement) {
	s.Rules = append(s.Rules, rule)
}

// NewRule creates a new rule
func NewRule(selector Selector) *Rule {
	return &Rule{
		Selector:     selector,
		Declarations: []Declaration{},
		Rules:        []Statement{},
	}
}

// AddDeclaration adds a declaration to a rule
func (r *Rule) AddDeclaration(decl Declaration) {
	r.Declarations = append(r.Declarations, decl)
}

// AddNestedRule adds a nested rule to a rule
func (r *Rule) AddNestedRule(nestedRule Statement) {
	r.Rules = append(r.Rules, nestedRule)
}
