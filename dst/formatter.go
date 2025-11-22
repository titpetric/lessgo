package dst

import (
	"bytes"
	"strings"
)

// Formatter formats a DST back into .less source with consistent indentation
type Formatter struct {
	indent int
	buf    *bytes.Buffer
}

// NewFormatter creates a new formatter
func NewFormatter() *Formatter {
	return &Formatter{
		indent: 0,
		buf:    &bytes.Buffer{},
	}
}

// Format formats a File into .less source code
func (f *Formatter) Format(file *File) string {
	f.buf.Reset()
	f.indent = 0

	for _, node := range file.Nodes {
		f.formatNode(node)
	}

	return f.buf.String()
}

// formatNode formats a single node
func (f *Formatter) formatNode(node Node) {
	switch n := node.(type) {
	case *Comment:
		f.formatComment(n)
	case *Decl:
		f.formatDecl(n)
	case *Block:
		f.formatBlock(n)
	case *MixinCall:
		f.formatMixinCall(n)
	case *Each:
		f.formatEach(n)
	}
}

// formatComment formats a comment node
func (f *Formatter) formatComment(c *Comment) {
	f.writeIndent()

	if c.Multiline {
		f.buf.WriteString("/* ")
		f.buf.WriteString(c.Text)
		f.buf.WriteString(" */")
	} else {
		f.buf.WriteString("// ")
		f.buf.WriteString(c.Text)
	}

	f.buf.WriteString("\n")
}

// formatDecl formats a declaration node
func (f *Formatter) formatDecl(d *Decl) {
	// Skip variable assignments (they're handled separately)
	if d.Key[0:1] == "@" && !strings.Contains(d.Key, "{") {
		// Variable assignment - output but don't store
		f.writeIndent()
		f.buf.WriteString(d.Key)
		f.buf.WriteString(": ")
		f.buf.WriteString(d.Value)
		// Add semicolon if not present
		if !strings.HasSuffix(d.Value, ";") {
			f.buf.WriteString(";")
		}
		f.buf.WriteString("\n")
		return
	}

	f.writeIndent()
	f.buf.WriteString(d.Key)
	f.buf.WriteString(": ")
	f.buf.WriteString(d.Value)

	// Add semicolon if not present
	if !strings.HasSuffix(d.Value, ";") {
		f.buf.WriteString(";")
	}

	f.buf.WriteString("\n")
}

// formatBlock formats a block node with nested children
func (f *Formatter) formatBlock(b *Block) {
	// Skip parametric mixin definitions (they're only invoked, not output)
	if len(b.Params) > 0 {
		return
	}

	f.writeIndent()

	// Write selectors (comma-separated if multiple)
	for i, name := range b.SelNames {
		if i > 0 {
			f.buf.WriteString(",\n")
			f.writeIndent()
		}
		f.buf.WriteString(name)
	}

	f.buf.WriteString(" {\n")
	f.indent++

	// Format children
	for _, child := range b.Children {
		f.formatNode(child)
	}

	f.indent--
	f.writeIndent()
	f.buf.WriteString("}")

	// Add blank line after top-level blocks
	if f.indent == 0 {
		f.buf.WriteString("\n\n")
	} else {
		f.buf.WriteString("\n")
	}
}

// formatMixinCall formats a mixin call
func (f *Formatter) formatMixinCall(m *MixinCall) {
	f.writeIndent()
	f.buf.WriteString(m.Name)
	f.buf.WriteString("(")
	for i, arg := range m.Args {
		if i > 0 {
			f.buf.WriteString("; ")
		}
		f.buf.WriteString(arg)
	}
	f.buf.WriteString(");\n")
}

// formatEach formats an each loop
func (f *Formatter) formatEach(e *Each) {
	f.writeIndent()
	f.buf.WriteString("each(")
	f.buf.WriteString(e.ListExpr)
	f.buf.WriteString(", ")
	f.buf.WriteString(e.VarName)
	f.buf.WriteString(") {\n")
	f.indent++

	for _, child := range e.Children {
		f.formatNode(child)
	}

	f.indent--
	f.writeIndent()
	f.buf.WriteString("}\n")
}

// writeIndent writes the current indentation
func (f *Formatter) writeIndent() {
	for i := 0; i < f.indent; i++ {
		f.buf.WriteString("  ")
	}
}
