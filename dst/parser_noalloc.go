package dst

import (
	"bufio"
	"io"
	"io/fs"
	"os"

	"github.com/titpetric/lessgo/internal/strings"
)

// getTrimmed returns a trimmed view of the string (no allocation via bounds check)
func getTrimmed(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}

	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// splitCommaNoAlloc splits by comma and trims each part, appending to buffer (no Split allocation)
func splitCommaNoAlloc(s string, buf *[]string) {
	*buf = (*buf)[:0]

	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			// Extract and trim substring
			part := getTrimmed(s[start:i])
			if part != "" {
				*buf = append(*buf, part)
			}
			start = i + 1
		}
	}

	// Add last part
	if start < len(s) {
		part := getTrimmed(s[start:])
		if part != "" {
			*buf = append(*buf, part)
		}
	}
}

// ParserNoAlloc is a zero-allocation variant of the parser using pre-allocated buffers
// It trades memory usage for speed by pre-allocating large slices
type ParserNoAlloc struct {
	scanner *bufio.Scanner
	line    string
	lineLen int // Cache line length to avoid len() calls
	eof     bool
	fs      fs.FS

	// Pre-allocated buffers to reduce allocations
	nodeBuffer     []Node          // Reusable slice for file nodes
	childBuffer    []Node          // Reusable slice for block children
	selectorBuffer []string        // Reusable slice for selectors
	paramBuffer    []string        // Reusable slice for parameters
	argBuffer      []string        // Reusable slice for arguments
	lineBuffer     strings.Builder // Reusable string builder
	interpolBuffer strings.Builder // Reusable string builder for interpolation

	// Fast path detection: cache for commonly checked patterns
	firstChar byte
	hasColon  bool
	hasOpen   bool
	hasSemi   bool
}

// NewParserNoAlloc creates a new zero-allocation parser with pre-allocated buffers
func NewParserNoAlloc(r io.Reader) *ParserNoAlloc {
	return &ParserNoAlloc{
		scanner:        bufio.NewScanner(r),
		eof:            false,
		fs:             os.DirFS("."),
		nodeBuffer:     make([]Node, 0, 256),  // Typical file has ~100-300 nodes
		childBuffer:    make([]Node, 0, 64),   // Typical block has ~20-50 children
		selectorBuffer: make([]string, 0, 16), // Selectors per block
		paramBuffer:    make([]string, 0, 8),  // Function parameters
		argBuffer:      make([]string, 0, 8),  // Mixin/function arguments
	}
}

// NewParserNoAllocWithFS creates a zero-allocation parser with custom filesystem
func NewParserNoAllocWithFS(r io.Reader, filesystem fs.FS) *ParserNoAlloc {
	return &ParserNoAlloc{
		scanner:        bufio.NewScanner(r),
		eof:            false,
		fs:             filesystem,
		nodeBuffer:     make([]Node, 0, 256),
		childBuffer:    make([]Node, 0, 64),
		selectorBuffer: make([]string, 0, 16),
		paramBuffer:    make([]string, 0, 8),
		argBuffer:      make([]string, 0, 8),
	}
}

// Parse parses the entire .less file into a File AST with minimal allocations
func (p *ParserNoAlloc) Parse() (*File, error) {
	file := &File{}

	// Clear and reuse buffer
	p.nodeBuffer = p.nodeBuffer[:0]

	for p.scan() {
		// Skip empty lines - use pre-analyzed firstChar
		if p.lineLen == 0 || p.firstChar == 0 {
			continue
		}

		// Check for each() function call first (high priority, specific pattern)
		line := getTrimmed(p.line)
		if strings.HasPrefix(line, "each(") {
			each, err := p.parseEachNoAlloc(line)
			if err != nil {
				return nil, err
			}
			if each != nil {
				p.nodeBuffer = append(p.nodeBuffer, each)
			}
			continue
		}

		// Fast path based on first character analysis
		switch p.firstChar {
		case '/':
			// Could be single or multi-line comment
			if p.lineLen >= 2 && p.line[1] == '/' {
				// Single-line comment - trim and extract text
				text := getTrimmed(p.line)
				text = strings.TrimPrefix(text, "//")
				text = getTrimmed(text)
				p.nodeBuffer = append(p.nodeBuffer, &Comment{
					Text:      text,
					Multiline: false,
				})
				continue
			} else if p.lineLen >= 2 && p.line[1] == '*' {
				// Multi-line comment start
				text := getTrimmed(p.line)
				comment := &Comment{Text: "", Multiline: true}
				p.readMultilineCommentNoAlloc(comment, text)
				p.nodeBuffer = append(p.nodeBuffer, comment)
				continue
			}

		case '@':
			// Variable or directive (@import, @name:)
			line := getTrimmed(p.line)
			if strings.HasPrefix(line, "@import") {
				p.parseImportNoAlloc(file, line)
				continue
			}
			// Variable assignment or block variable
			if p.hasColon {
				if p.hasOpen {
					// Block variable (@name: { ... };) - brace on same line
					blockVar, err := p.parseBlockVariableNoAlloc(line)
					if err != nil {
						return nil, err
					}
					if blockVar != nil {
						p.nodeBuffer = append(p.nodeBuffer, blockVar)
					}
				} else if !p.hasSemi {
					// Possible multi-line block variable (@name: <EOL>, next line should be {)
					// Save the variable name for now
					colonIdx := strings.Index(line, ":")
					varName := strings.TrimPrefix(getTrimmed(line[:colonIdx]), "@")

					// Peek ahead to see if next line is opening brace
					blockFound := false
					for p.scan() {
						nextLine := getTrimmed(p.line)
						if nextLine == "" {
							continue
						}
						if strings.HasPrefix(nextLine, "{") {
							blockFound = true
							// Construct a single-line version for parsing
							line = "@" + varName + ": {" + nextLine[1:]
						}
						break
					}

					if blockFound {
						// Parse as block variable
						blockVar, err := p.parseBlockVariableNoAlloc(line)
						if err != nil {
							return nil, err
						}
						if blockVar != nil {
							p.nodeBuffer = append(p.nodeBuffer, blockVar)
						}
					}
				} else if p.hasSemi {
					// Variable assignment (@name: value;)
					decl, err := p.parseDeclNoAlloc(line)
					if err != nil {
						return nil, err
					}
					if decl != nil {
						p.nodeBuffer = append(p.nodeBuffer, decl)
					}
				}
				continue
			}

		default:
			// Regular selector or declaration
			line := getTrimmed(p.line)

			// Block (selector { ... })
			if p.hasOpen {
				block, err := p.parseBlockNoAlloc(line)
				if err != nil {
					return nil, err
				}
				if block != nil {
					p.nodeBuffer = append(p.nodeBuffer, block)
				}
				continue
			}

			// Declaration without block (property: value;)
			if p.hasColon && p.hasSemi {
				decl, err := p.parseDeclNoAlloc(line)
				if err != nil {
					return nil, err
				}
				if decl != nil {
					p.nodeBuffer = append(p.nodeBuffer, decl)
				}
				continue
			}

			// Mixin call (name();)
			if p.hasSemi && strings.Contains(line, "(") {
				parenIdx := strings.Index(line, "(")
				if parenIdx > 0 {
					firstPart := getTrimmed(line[:parenIdx])
					// Check if mixin (starts with . # &)
					if p.isMixinName(firstPart) {
						args := p.parseArgsNoAlloc(line)
						p.nodeBuffer = append(p.nodeBuffer, &MixinCall{Name: firstPart, Args: args})
					}
				}
				continue
			}
		}
	}

	// Transfer buffer to file
	file.Nodes = p.nodeBuffer
	return file, nil
}

// isMixinName checks if a name is a mixin (starts with . # or &)
func (p *ParserNoAlloc) isMixinName(name string) bool {
	if len(name) == 0 {
		return false
	}
	first := name[0]
	return first == '.' || first == '#' || first == '&'
}

// readMultilineCommentNoAlloc reads a multi-line comment with minimal allocations
func (p *ParserNoAlloc) readMultilineCommentNoAlloc(comment *Comment, line string) {
	p.lineBuffer.Reset()

	// Handle comment start on the same line
	if idx := strings.Index(line, "/*"); idx != -1 {
		start := idx + 2
		if endIdx := strings.Index(line[start:], "*/"); endIdx != -1 {
			comment.Text = line[start : start+endIdx]
			return
		}
		p.lineBuffer.WriteString(line[start:])
	}

	for p.scan() {
		line := getTrimmed(p.line)

		if idx := strings.Index(line, "*/"); idx != -1 {
			if idx > 0 {
				p.lineBuffer.WriteByte('\n')
				p.lineBuffer.WriteString(line[:idx])
			}
			break
		}

		if p.lineBuffer.Len() > 0 {
			p.lineBuffer.WriteByte('\n')
		}
		p.lineBuffer.WriteString(line)
	}

	comment.Text = p.lineBuffer.String()
}

// parseImportNoAlloc parses import statements with minimal allocations
func (p *ParserNoAlloc) parseImportNoAlloc(file *File, line string) {
	// Extract filename from @import "filename.less";
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")

	if start == -1 || end == -1 || start == end {
		return
	}

	filename := line[start+1 : end]

	// Try to read the imported file
	content, err := fs.ReadFile(p.fs, filename)
	if err != nil {
		return
	}

	// Parse imported file
	importedFile, err := NewParserNoAllocWithFS(strings.NewReader(string(content)), p.fs).Parse()
	if err != nil {
		return
	}

	// Prepend imported nodes
	file.Nodes = append(importedFile.Nodes, file.Nodes...)
}

// parseBlockVariableNoAlloc parses block variable definitions with minimal allocations
func (p *ParserNoAlloc) parseBlockVariableNoAlloc(line string) (*BlockVariable, error) {
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return nil, nil
	}

	name := strings.TrimPrefix(getTrimmed(line[:colonIdx]), "@")

	// Check for opening brace
	braceStart := strings.Index(line, "{")
	if braceStart == -1 {
		return nil, nil // No block opening
	}

	blockVar := &BlockVariable{
		Name:     name,
		Children: p.childBuffer[:0], // Clear buffer
	}

	// Read block content
	p.childBuffer = p.childBuffer[:0]

	// Check if block ends on the same line
	braceEnd := strings.LastIndex(line, "}")
	if braceEnd != -1 && braceEnd > braceStart {
		// Single-line block variable
		// Parse any declarations on the first line after {
		contentOnFirstLine := getTrimmed(line[braceStart+1 : braceEnd])
		if contentOnFirstLine != "" && strings.Contains(contentOnFirstLine, ":") {
			if decl, err := p.parseDeclNoAlloc(contentOnFirstLine + ";"); err == nil && decl != nil {
				p.childBuffer = append(p.childBuffer, decl)
			}
		}
	} else {
		// Multi-line block variable - read until closing }
		for p.scan() {
			trimmedLine := getTrimmed(p.line)

			if trimmedLine == "" {
				continue
			}

			if strings.HasPrefix(trimmedLine, "}") || trimmedLine == "};" {
				break
			}

			if strings.Contains(trimmedLine, ":") && strings.HasSuffix(trimmedLine, ";") {
				if decl, err := p.parseDeclNoAlloc(trimmedLine); err == nil && decl != nil {
					p.childBuffer = append(p.childBuffer, decl)
				}
			}
		}
	}

	blockVar.Children = make([]Node, len(p.childBuffer))
	copy(blockVar.Children, p.childBuffer)

	return blockVar, nil
}

// parseDeclNoAlloc parses declarations with minimal allocations
func (p *ParserNoAlloc) parseDeclNoAlloc(line string) (*Decl, error) {
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return nil, nil
	}

	key := getTrimmed(line[:colonIdx])
	value := getTrimmed(line[colonIdx+1:])

	// Remove trailing semicolon
	if strings.HasSuffix(value, ";") {
		value = value[:len(value)-1]
	}

	value = getTrimmed(value)

	return &Decl{Key: key, Value: value}, nil
}

// parseEachNoAlloc parses each(list, { ... }) loops with minimal allocations
func (p *ParserNoAlloc) parseEachNoAlloc(line string) (*Each, error) {
	// Format: each(list_expr, { ... });
	if !strings.HasPrefix(line, "each(") || !strings.Contains(line, "{") {
		return nil, nil
	}

	// Find the opening paren after "each"
	openParen := 5 // Length of "each("

	// Find the comma that separates list and block
	// Need to handle nested parens: each(range(3), { ... })
	commaIdx := -1
	parenDepth := 1
	for i := openParen; i < len(line); i++ {
		c := line[i]
		if c == '(' {
			parenDepth++
		} else if c == ')' {
			parenDepth--
		} else if c == ',' && parenDepth == 1 {
			commaIdx = i
			break
		}
	}

	if commaIdx == -1 {
		return nil, nil // Not a properly formatted each() call
	}

	// Extract the list expression (between each( and ,)
	listExpr := getTrimmed(line[openParen:commaIdx])

	// The block variable name is "value" by default for each()
	each := &Each{
		ListExpr: listExpr,
		VarName:  "value",
		Children: []Node{},
	}

	// Find the { and parse the block content
	blockStart := strings.Index(line[commaIdx:], "{")
	if blockStart == -1 {
		return nil, nil
	}
	blockStart += commaIdx

	// Check if this is a single-line block or multi-line
	if strings.HasSuffix(line, "});") {
		// Single-line block or block starts on this line
		// Parse as a sub-block
		blockLine := "{" + strings.TrimPrefix(line[blockStart:], "{")
		firstContent := blockLine[1 : len(blockLine)-2] // Remove { and });
		firstContent = getTrimmed(firstContent)

		// If there's content on the first line, it's part of the block
		if firstContent != "" {
			// Parse the selector/content
			if strings.Contains(firstContent, "{") {
				// Nested selector on same line
				nestedBlock, _ := p.parseBlockNoAlloc(firstContent)
				if nestedBlock != nil {
					p.childBuffer = p.childBuffer[:0]
					p.childBuffer = append(p.childBuffer, nestedBlock)
					each.Children = make([]Node, len(p.childBuffer))
					copy(each.Children, p.childBuffer)
				}
			}
		}
		return each, nil
	}

	// Multi-line case: read until we find the closing });
	p.childBuffer = p.childBuffer[:0]

	// Parse first line content if any
	firstLineContent := getTrimmed(line[blockStart+1:])
	if firstLineContent != "" && !strings.HasPrefix(firstLineContent, "}") {
		// Parse selector or content from first line
		if strings.Contains(firstLineContent, "{") {
			block, _ := p.parseBlockNoAlloc(firstLineContent)
			if block != nil {
				p.childBuffer = append(p.childBuffer, block)
			}
		}
	}

	// Read remaining lines until we find });
	for p.scan() {
		trimmedLine := getTrimmed(p.line)

		if trimmedLine == "" {
			continue
		}

		// Check for closing });
		if trimmedLine == "});" || strings.HasSuffix(trimmedLine, "});") {
			break
		}

		// Parse content
		if strings.Contains(trimmedLine, "{") {
			// Save current buffer state before calling parseBlockNoAlloc
			savedLen := len(p.childBuffer)
			block, _ := p.parseBlockNoAlloc(trimmedLine)
			if block != nil {
				// Restore buffer and append block
				p.childBuffer = p.childBuffer[:savedLen]
				p.childBuffer = append(p.childBuffer, block)
			}
		} else if strings.Contains(trimmedLine, ":") && strings.HasSuffix(trimmedLine, ";") {
			decl, _ := p.parseDeclNoAlloc(trimmedLine)
			if decl != nil {
				p.childBuffer = append(p.childBuffer, decl)
			}
		}
	}

	each.Children = make([]Node, len(p.childBuffer))
	copy(each.Children, p.childBuffer)

	return each, nil
}

// parseBlockNoAlloc parses block nodes with minimal allocations
func (p *ParserNoAlloc) parseBlockNoAlloc(line string) (*Block, error) {
	// Find the block-opening brace, skipping any braces inside @{...} patterns (interpolation)
	braceIdx := -1
	inInterpolation := false
	for i := 0; i < len(line); i++ {
		if i > 0 && line[i-1] == '@' && line[i] == '{' {
			inInterpolation = true
		} else if inInterpolation && line[i] == '}' {
			inInterpolation = false
		} else if !inInterpolation && line[i] == '{' {
			braceIdx = i
			break
		}
	}

	if braceIdx == -1 {
		return nil, nil
	}

	selectorStr := getTrimmed(line[:braceIdx])
	if selectorStr == "" {
		return nil, nil
	}

	// Parse selectors (comma-separated) - avoid Split allocation when possible
	p.selectorBuffer = p.selectorBuffer[:0]
	if !strings.Contains(selectorStr, ",") {
		// Single selector fast path
		p.selectorBuffer = append(p.selectorBuffer, selectorStr)
	} else {
		// Multiple selectors without Split allocation
		splitCommaNoAlloc(selectorStr, &p.selectorBuffer)
	}

	block := &Block{
		SelNames:        make([]string, len(p.selectorBuffer)),
		Children:        p.childBuffer[:0],
		IsMixinFunction: false,
	}
	copy(block.SelNames, p.selectorBuffer)

	// Clear children buffer
	p.childBuffer = p.childBuffer[:0]

	// Read block content using pre-analyzed pattern
	for p.scan() {
		// Skip empty lines
		if p.lineLen == 0 || p.firstChar == 0 {
			continue
		}

		// Check for closing brace first (most common escape)
		if p.firstChar == '}' {
			break
		}

		trimmedLine := getTrimmed(p.line)

		// Single-line comment
		if p.firstChar == '/' && p.lineLen >= 2 && p.line[1] == '/' {
			text := strings.TrimPrefix(trimmedLine, "//")
			text = getTrimmed(text)
			p.childBuffer = append(p.childBuffer, &Comment{
				Text:      text,
				Multiline: false,
			})
			continue
		}

		// Multi-line comment
		if p.firstChar == '/' && p.lineLen >= 2 && p.line[1] == '*' {
			comment := &Comment{Text: "", Multiline: true}
			p.readMultilineCommentNoAlloc(comment, trimmedLine)
			p.childBuffer = append(p.childBuffer, comment)
			continue
		}

		// Nested block
		if p.hasOpen {
			nestedBlock, err := p.parseBlockNoAlloc(trimmedLine)
			if err == nil && nestedBlock != nil {
				p.childBuffer = append(p.childBuffer, nestedBlock)
			}
			continue
		}

		// Declaration
		if p.hasColon && p.hasSemi {
			if decl, err := p.parseDeclNoAlloc(trimmedLine); err == nil && decl != nil {
				p.childBuffer = append(p.childBuffer, decl)
			}
			continue
		}

		// Mixin call, block variable call, or function call
		if p.hasSemi && strings.Contains(trimmedLine, "(") {
			parenIdx := strings.Index(trimmedLine, "(")
			if parenIdx > 0 {
				firstPart := getTrimmed(trimmedLine[:parenIdx])

				// Check if this is a block variable call (@varname();)
				if strings.HasPrefix(firstPart, "@") && !strings.Contains(firstPart, "{") {
					// Block variable call - parse as declaration
					decl := &Decl{
						Key:   firstPart,
						Value: "()",
					}
					p.childBuffer = append(p.childBuffer, decl)
				} else if p.isMixinName(firstPart) {
					// Regular mixin call
					args := p.parseArgsNoAlloc(trimmedLine)
					p.childBuffer = append(p.childBuffer, &MixinCall{Name: firstPart, Args: args})
				}
			}
		}
	}

	// Make a copy of the buffer contents before returning
	// This is important because the caller might reuse p.childBuffer
	blockChildren := make([]Node, len(p.childBuffer))
	copy(blockChildren, p.childBuffer)
	block.Children = blockChildren

	// DO NOT clear p.childBuffer here - the caller might be using it

	return block, nil
}

// parseArgsNoAlloc extracts function/mixin arguments with minimal allocations
func (p *ParserNoAlloc) parseArgsNoAlloc(line string) []string {
	parenStart := strings.Index(line, "(")
	parenEnd := strings.LastIndex(line, ")")

	if parenStart == -1 || parenEnd == -1 || parenStart >= parenEnd {
		return []string{}
	}

	argStr := line[parenStart+1 : parenEnd]
	if getTrimmed(argStr) == "" {
		return []string{}
	}

	// Split arguments without allocation
	splitCommaNoAlloc(argStr, &p.argBuffer)

	// Copy to new slice
	args := make([]string, len(p.argBuffer))
	copy(args, p.argBuffer)
	return args
}

// scan reads the next line from the scanner and pre-analyzes it
func (p *ParserNoAlloc) scan() bool {
	if p.eof {
		return false
	}

	if p.scanner.Scan() {
		p.line = p.scanner.Text()
		p.lineLen = len(p.line)
		// Pre-analyze the line for common patterns to avoid repeated string ops
		p.analyzeLinePattern()
		return true
	}

	p.eof = true
	return false
}

// analyzeLinePattern pre-scans the line for common structural elements
// This avoids repeated HasPrefix, Contains, and HasSuffix calls
func (p *ParserNoAlloc) analyzeLinePattern() {
	if p.lineLen == 0 {
		p.firstChar = 0
		p.hasColon = false
		p.hasOpen = false
		p.hasSemi = false
		return
	}

	// Find first non-space char and its position
	firstNonSpace := -1
	for i := 0; i < p.lineLen; i++ {
		c := p.line[i]
		if c != ' ' && c != '\t' {
			p.firstChar = c
			firstNonSpace = i
			break
		}
	}

	if firstNonSpace == -1 {
		// All spaces/tabs
		p.firstChar = 0
		p.hasColon = false
		p.hasOpen = false
		p.hasSemi = false
		return
	}

	// Scan for key chars: : { ; )
	// Use indexed loops to avoid allocation
	p.hasColon = false
	p.hasOpen = false
	p.hasSemi = false

	for i := firstNonSpace; i < p.lineLen; i++ {
		c := p.line[i]
		if c == ':' {
			p.hasColon = true
		} else if c == '{' {
			p.hasOpen = true
		} else if c == ';' {
			p.hasSemi = true
		}
	}
}
