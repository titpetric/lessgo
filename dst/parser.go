package dst

import (
	"bufio"
	"io"
	"io/fs"
	"os"

	"github.com/titpetric/lessgo/expression"
	"github.com/titpetric/lessgo/internal/strings"
)

// Parser reads and parses .less files into a DST
type Parser struct {
	scanner *bufio.Scanner
	line    string
	eof     bool
	fs      fs.FS // filesystem for resolving imports

	// Pre-allocated buffers for zero-alloc splitting
	selectorBuf []string // For selector splitting (comma-separated)
	declBuf     []string // For declaration splitting (semicolon-separated)
	argBuf      []string // For argument splitting (comma-separated)
}

// NewParser creates a new parser from a reader with OS filesystem
func NewParser(r io.Reader) *Parser {
	return &Parser{
		scanner:     bufio.NewScanner(r),
		eof:         false,
		fs:          os.DirFS("."),
		selectorBuf: make([]string, 0, 16),
		declBuf:     make([]string, 0, 32),
		argBuf:      make([]string, 0, 16),
	}
}

// NewParserWithFS creates a new parser with a custom filesystem
func NewParserWithFS(r io.Reader, filesystem fs.FS) *Parser {
	return &Parser{
		scanner:     bufio.NewScanner(r),
		eof:         false,
		fs:          filesystem,
		selectorBuf: make([]string, 0, 16),
		declBuf:     make([]string, 0, 32),
		argBuf:      make([]string, 0, 16),
	}
}

// Parse parses the entire .less file into a File AST
func (p *Parser) Parse() (*File, error) {
	file := &File{}

	for p.scan() {

		line := strings.TrimSpace(p.line)

		// Skip empty lines

		if line == "" {
			continue
		}

		// Single-line comment

		if strings.HasPrefix(line, "//") {

			file.Nodes = append(file.Nodes, &Comment{
				Text: strings.TrimPrefix(line, "//"),

				Multiline: false,
			})

			continue

		}

		// Multi-line comment start

		if strings.HasPrefix(line, "/*") {

			comment := &Comment{Text: "", Multiline: true}

			p.readMultilineComment(comment, line)

			file.Nodes = append(file.Nodes, comment)

			continue

		}

		// Import statement (@import "file.less";)

		if strings.HasPrefix(line, "@import") {

			p.parseImport(file, line)

			continue

		}

		// Block variable definition (@name: { ... };)
		if strings.HasPrefix(line, "@") && strings.Contains(line, ":") && strings.Contains(line, "{") {

			blockVar, err := p.parseBlockVariable(line)
			if err != nil {
				return nil, err
			}

			if blockVar != nil {
				file.Nodes = append(file.Nodes, blockVar)
				continue
			}

		}

		// Variable definition (@name: value;)

		if strings.HasPrefix(line, "@") && strings.Contains(line, ":") && strings.HasSuffix(line, ";") {

			decl := p.parseDecl(line)

			if decl != nil {
				file.Nodes = append(file.Nodes, decl)
			}

			continue

		}

		// Each function call (each(list, { ... });)
		if strings.HasPrefix(line, "each(") {
			each, err := p.parseEach(line)
			if err != nil {
				return nil, err
			}

			if each != nil {
				file.Nodes = append(file.Nodes, each)
			}

			continue
		}

		// Block or declaration or top-level mixin call

		if strings.Contains(line, "{") {

			block, err := p.parseBlock(line)
			if err != nil {
				return nil, err
			}

			file.Nodes = append(file.Nodes, block)

		} else if strings.Contains(line, ":") && strings.HasSuffix(line, ";") {

			decl := p.parseDecl(line)

			if decl != nil {
				file.Nodes = append(file.Nodes, decl)
			}

		} else if strings.Contains(line, "(") && strings.HasSuffix(line, ");") {

			// Top-level mixin call (e.g., .mixin();)
			parenIdx := strings.Index(line, "(")
			firstPart := strings.TrimSpace(line[:parenIdx])

			// Check if this is a mixin call (selector-like: starts with . # &)
			isMixin := strings.HasPrefix(firstPart, ".") || strings.HasPrefix(firstPart, "#") || strings.HasPrefix(firstPart, "&")

			if isMixin {
				argsStr := strings.TrimSpace(line[parenIdx+1 : len(line)-2])
				var args []string
				if argsStr != "" {
					for _, arg := range splitParameterList(argsStr) {
						args = append(args, strings.TrimSpace(arg))
					}
				}
				file.Nodes = append(file.Nodes, &MixinCall{Name: firstPart, Args: args})
			}

		}

	}

	return file, nil
}

// parseImport parses an import statement like @import "file.less";

func (p *Parser) parseImport(file *File, line string) {
	// Extract the file path from @import "path";

	// Match patterns like @import "path/file.less"; or @import 'path/file.less';

	line = strings.TrimSpace(line)

	// Remove @import and semicolon

	line = strings.TrimPrefix(line, "@import")

	line = strings.TrimSuffix(line, ";")

	line = strings.TrimSpace(line)

	// Remove quotes

	var filePath string

	if strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
		filePath = line[1 : len(line)-1]
	} else if strings.HasPrefix(line, "'") && strings.HasSuffix(line, "'") {
		filePath = line[1 : len(line)-1]
	} else {
		return
	}

	// Check if this is a URL import (http://, https://, or protocol-relative //)
	// These should pass through to CSS output, not be processed
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") || strings.HasPrefix(filePath, "//") {
		file.Nodes = append(file.Nodes, &Import{Path: filePath})
		return
	}

	// Open the file from the filesystem

	f, err := p.fs.Open(filePath)
	if err != nil {
		return
	}

	defer f.Close()

	// Parse the imported file with the same filesystem

	importedParser := NewParserWithFS(f, p.fs)

	importedFile, err := importedParser.Parse()
	if err != nil {
		return
	}

	// Then, prepend imported nodes to file nodes

	file.Nodes = append(importedFile.Nodes, file.Nodes...)
}

// parseBlock parses a selector block with nested nodes

func (p *Parser) parseBlock(line string) (*Block, error) {
	// Extract selector (everything before {)
	// Find the block-opening brace, skipping any braces inside @{...} patterns

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
		braceIdx = len(line)
	}

	selectorStr := strings.TrimSpace(line[:braceIdx])

	var guard *Guard

	find := " when "

	if strings.Contains(selectorStr, find) {

		parts := strings.SplitN(selectorStr, find, 2)

		selectorStr = parts[0]

		guard = &Guard{
			Condition: parts[1],
		}

	}

	// Only treat as mixin function if contains parentheses AND starts with . or # (valid mixin prefixes)
	// Exclude at-rules (@media) and pseudo-classes (:nth-child, :hover, etc.)
	trimmedSel := strings.TrimSpace(selectorStr)
	isMixinFunction := strings.Contains(selectorStr, "(") && strings.Contains(selectorStr, ")") &&
		(strings.HasPrefix(trimmedSel, ".") || strings.HasPrefix(trimmedSel, "#")) &&
		!strings.Contains(selectorStr, ":")

	// Parse selectors and parameters
	// Split selectors by comma, but respect @{...} interpolation blocks

	var selectors []string

	var params []string

	selectorList := splitSelectorList(selectorStr)

	for _, sel := range selectorList {

		sel = strings.TrimSpace(sel)

		// Check if this is a parametric mixin (e.g., .mixin(@v))
		// Only treat as mixin if it starts with . or # (valid mixin prefixes)
		// Do not treat CSS pseudo-classes like :nth-child() or at-rules like @media as mixins

		isMixinSelector := (strings.HasPrefix(sel, ".") || strings.HasPrefix(sel, "#")) &&
			strings.Contains(sel, "(") && strings.HasSuffix(sel, ")") &&
			!strings.Contains(sel, ":")

		if isMixinSelector {

			parenIdx := strings.Index(sel, "(")

			selectorName := strings.TrimSpace(sel[:parenIdx])

			paramsStr := strings.TrimSpace(sel[parenIdx+1 : len(sel)-1])

			selectors = append(selectors, selectorName)

			// Parse parameters (comma-separated, but respect @{...})

			if paramsStr != "" {
				for _, param := range splitParameterList(paramsStr) {
					params = append(params, strings.TrimSpace(param))
				}
			}

		} else {
			selectors = append(selectors, sel)
		}

	}

	block := &Block{
		SelNames: selectors,

		Parent: nil, // Will be set by caller if nested

		Children: []Node{},

		Params: params,

		IsMixinFunction: isMixinFunction,

		Guard: guard,
	}

	// Read nested content until closing }

	for p.scan() {

		line := strings.TrimSpace(p.line)

		if line == "" {
			continue
		}

		if line == "}" {
			break
		}

		// Single-line comment

		if strings.HasPrefix(line, "//") {

			block.Children = append(block.Children, &Comment{
				Text: strings.TrimPrefix(line, "//"),

				Multiline: false,
			})

			continue

		}

		// Multi-line comment

		if strings.HasPrefix(line, "/*") {

			comment := &Comment{Text: "", Multiline: true}

			p.readMultilineComment(comment, line)

			block.Children = append(block.Children, comment)

			continue

		}

		// Single-line block (e.g., "p { margin: 0; padding: 0; }")
		// But skip if braces are part of @{...} interpolation
		braceOpen, braceClose := findBlockBraces(line)
		if braceOpen != -1 && braceClose != -1 {

			// Parse inline declarations

			selectorStr := strings.TrimSpace(line[:braceOpen])

			contentStr := strings.TrimSpace(line[braceOpen+1 : braceClose])

			// Split comma-separated selectors (zero-alloc)
			strings.SplitCommaNoAlloc(selectorStr, &p.selectorBuf)
			// Make a copy since the buffer will be reused
			selectors := make([]string, len(p.selectorBuf))
			copy(selectors, p.selectorBuf)

			inlineBlock := &Block{
				SelNames: selectors,

				Parent: block,

				Children: []Node{},
			}

			// Parse declarations in the inline block (zero-alloc)
			strings.SplitByteNoAlloc(contentStr, ';', &p.declBuf)
			for _, declStr := range p.declBuf {
				if declStr == "" {
					continue
				}

				decl := p.parseDecl(declStr + ";")

				if decl != nil {
					inlineBlock.Children = append(inlineBlock.Children, decl)
				}

			}

			block.Children = append(block.Children, inlineBlock)

		} else if containsRealBrace(line) {

			// Multi-line nested block

			nestedBlock, err := p.parseBlock(line)
			if err != nil {
				return nil, err
			}

			nestedBlock.Parent = block // Set parent reference for & resolution

			block.Children = append(block.Children, nestedBlock)

		} else if strings.Contains(line, "(") && strings.HasSuffix(line, ");") {

			// Mixin call (with or without arguments) or block variable call or function call in declaration

			parenIdx := strings.Index(line, "(")

			firstPart := strings.TrimSpace(line[:parenIdx])

			// Check if this is a block variable call (@varname();)
			if strings.HasPrefix(firstPart, "@") && !strings.Contains(firstPart, "{") {
				// Block variable call - parse as declaration
				decl := &Decl{
					SelNames: []string{},
					Key:      firstPart,
					Value:    "()",
				}
				block.Children = append(block.Children, decl)
			} else {
				// Check if this is a mixin call (selector-like: starts with . # &) not a function call

				isMixin := strings.HasPrefix(firstPart, ".") || strings.HasPrefix(firstPart, "#") || strings.HasPrefix(firstPart, "&")

				// Also check if it's a known LESS function - if so, it's not a mixin

				isKnownFunc := expression.IsFunctionCall(line) && expression.IsRegisteredFunction(firstPart)

				if isMixin && !isKnownFunc {

					argsStr := strings.TrimSpace(line[parenIdx+1 : len(line)-2])

					var args []string
					if argsStr != "" {
						strings.SplitCommaNoAlloc(argsStr, &p.argBuf)
						// Make a copy since the buffer will be reused
						args = make([]string, len(p.argBuf))
						copy(args, p.argBuf)
					}

					block.Children = append(block.Children, &MixinCall{Name: firstPart, Args: args})

				} else {

					// It's a function call in a declaration value, treat as declaration

					decl := p.parseDecl(line)

					if decl != nil {
						block.Children = append(block.Children, decl)
					}

				}
			}

		} else if strings.HasSuffix(line, ";") && !strings.Contains(line, ":") && !strings.Contains(line, "{") && (strings.HasPrefix(strings.TrimSpace(line), ".") || strings.HasPrefix(strings.TrimSpace(line), "#") || strings.HasPrefix(strings.TrimSpace(line), "&")) {

			// Mixin call without parentheses (e.g., ".mixin;")
			mixinName := strings.TrimSuffix(strings.TrimSpace(line), ";")
			block.Children = append(block.Children, &MixinCall{Name: mixinName, Args: []string{}})

		} else if strings.Contains(line, ":") && strings.HasSuffix(line, ";") {

			// Declaration

			decl := p.parseDecl(line)

			if decl != nil {
				block.Children = append(block.Children, decl)
			}

		}

	}

	return block, nil
}

// parseBlockVariable parses a block variable (@name: { ... };)
func (p *Parser) parseBlockVariable(line string) (*BlockVariable, error) {
	// Check if this looks like a block variable: @name: { ... };
	if !strings.HasPrefix(line, "@") || !strings.Contains(line, ":") || !strings.Contains(line, "{") {
		return nil, nil
	}

	// Extract variable name (between @ and :)
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return nil, nil
	}

	varName := strings.TrimSpace(line[1:colonIdx])
	if varName == "" || !isValidVarName(varName) {
		return nil, nil
	}

	// The rest should be { ... };
	rest := strings.TrimSpace(line[colonIdx+1:])
	if !strings.HasPrefix(rest, "{") {
		return nil, nil
	}

	// Check if this is a single-line block variable
	if strings.HasSuffix(rest, "};") {
		// Single-line case
		blockContent := rest[1 : len(rest)-2] // Remove { and };
		blockVar := &BlockVariable{
			Name:     varName,
			Children: []Node{},
		}

		// Parse simple single-line declarations (zero-alloc)
		strings.SplitByteNoAlloc(blockContent, ';', &p.declBuf)
		for _, declStr := range p.declBuf {
			if declStr == "" {
				continue
			}
			if !strings.Contains(declStr, ":") {
				continue
			}
			decl := p.parseDecl(declStr + ";")
			if decl != nil {
				blockVar.Children = append(blockVar.Children, decl)
			}
		}

		return blockVar, nil
	}

	// Multi-line case: we need to read until we find the closing };
	blockVar := &BlockVariable{
		Name:     varName,
		Children: []Node{},
	}

	// Read content after opening {
	// If the first line contains {, start after it
	firstLineContent := rest[1:] // Skip opening {
	if firstLineContent != "" && firstLineContent != "}" {
		declStr := strings.TrimSpace(firstLineContent)
		if strings.Contains(declStr, ":") && strings.HasSuffix(declStr, ";") {
			decl := p.parseDecl(declStr)
			if decl != nil {
				blockVar.Children = append(blockVar.Children, decl)
			}
		}
	}

	// Read remaining lines
	for p.scan() {
		line := strings.TrimSpace(p.line)

		if line == "" {
			continue
		}

		if line == "}" || line == "};" {
			break
		}

		// Single-line comment
		if strings.HasPrefix(line, "//") {
			blockVar.Children = append(blockVar.Children, &Comment{
				Text:      strings.TrimPrefix(line, "//"),
				Multiline: false,
			})
			continue
		}

		// Declaration
		if strings.Contains(line, ":") && strings.HasSuffix(line, ";") {
			decl := p.parseDecl(line)
			if decl != nil {
				blockVar.Children = append(blockVar.Children, decl)
			}
		}
	}

	return blockVar, nil
}

// parseEach parses an each() loop (each(list, { ... });)
func (p *Parser) parseEach(line string) (*Each, error) {
	// Check if this looks like each(list, { ... });
	if !strings.HasPrefix(line, "each(") || !strings.Contains(line, "{") {
		return nil, nil
	}

	// Find the opening paren after "each"
	openParen := 5 // Length of "each("

	// Find the matching closing paren for the list argument
	// The format is: each(list_expr, { ... });
	// We need to find the comma that separates list and block
	commaIdx := -1
	parenDepth := 1
	for i := openParen; i < len(line); i++ {
		if line[i] == '(' {
			parenDepth++
		} else if line[i] == ')' {
			parenDepth--
		} else if line[i] == ',' && parenDepth == 1 {
			commaIdx = i
			break
		}
	}

	if commaIdx == -1 {
		return nil, nil // Not an each() call
	}

	// Extract the list expression (between each( and ,)
	listExpr := strings.TrimSpace(line[openParen:commaIdx])

	// The block variable name is typically "value" (default for each)
	// The block is parsed as a regular block starting from the {
	each := &Each{
		ListExpr: listExpr,
		VarName:  "value", // Default variable name
		Children: []Node{},
	}

	// Find the { and parse the block content
	blockStart := strings.Index(line[commaIdx:], "{")
	if blockStart == -1 {
		return nil, nil
	}
	blockStart += commaIdx

	// Check if this is a single-line block
	if strings.HasSuffix(line, "});") {
		// Single-line case
		blockContent := line[blockStart+1 : len(line)-2] // Remove { and });
		blockContent = strings.TrimSpace(blockContent)

		// Parse block selectors and content
		// For now, we just store the raw block content for multi-line processing
		// but we need to read it as nested blocks/declarations
		// This is simplified - in reality we'd need to recursively parse
		return each, nil
	}

	// Multi-line case: read until we find the closing });
	// Read content after opening {
	firstLineContent := strings.TrimSpace(line[blockStart+1:])
	if firstLineContent != "" && firstLineContent != "}" && !strings.HasPrefix(firstLineContent, "}") {
		// This could be a selector or declaration on the first line
		// For now, store it for parsing
	}

	// Read remaining lines until we find });
	for p.scan() {
		line := strings.TrimSpace(p.line)

		if line == "" {
			continue
		}

		// Check for closing });
		if line == "});" || strings.HasSuffix(line, "});") {
			break
		}

		// Skip comments for now
		if strings.HasPrefix(line, "//") {
			continue
		}

		// Try to parse as block selector
		if strings.Contains(line, "{") {
			block, err := p.parseBlock(line)
			if err != nil {
				return nil, err
			}
			if block != nil {
				each.Children = append(each.Children, block)
			}
		} else if strings.Contains(line, ":") && strings.HasSuffix(line, ";") {
			// Declaration
			decl := p.parseDecl(line)
			if decl != nil {
				each.Children = append(each.Children, decl)
			}
		}
	}

	return each, nil
}

// isValidVarName checks if a string is a valid LESS variable name
func isValidVarName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if !isLetterFunc(rune(name[0])) && name[0] != '_' && name[0] != '-' {
		return false
	}
	for _, ch := range name[1:] {
		if !isLetterFunc(ch) && !isDigitFunc(ch) && ch != '_' && ch != '-' {
			return false
		}
	}
	return true
}

func isLetterFunc(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigitFunc(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// parseDecl parses a CSS declaration (key: value;)

func (p *Parser) parseDecl(line string) *Decl {
	// Remove trailing semicolon

	line = strings.TrimSuffix(strings.TrimSpace(line), ";")

	// Split on first colon only (to handle colons in values like URLs or format strings)
	parts := strings.SplitN(line, ":", 2)

	if len(parts) != 2 {
		return nil
	}

	key := strings.TrimSpace(parts[0])

	value := strings.TrimSpace(parts[1])

	// Normalize commas in the value (ensure each comma has a space after it)
	value = normalizeCommas(value)

	return &Decl{
		SelNames: []string{},

		Key: key,

		Value: value,
	}
}

// normalizeCommas ensures each comma in a value is followed by a space.
// Handles nested functions (parentheses) and respects quoted strings.
func normalizeCommas(value string) string {
	var result strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	parenDepth := 0

	for i, ch := range value {
		// Track quoted strings
		if (ch == '"' || ch == '\'') && (i == 0 || value[i-1] != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
		}

		// Track parenthesis depth (for nested functions)
		if !inQuotes {
			if ch == '(' {
				parenDepth++
			} else if ch == ')' {
				parenDepth--
			}
		}

		result.WriteRune(ch)

		// After a comma, ensure there's a space (if not in quotes and inside/outside parens)
		if ch == ',' && !inQuotes {
			// Look ahead to see if next char is not already a space
			if i+1 < len(value) && value[i+1] != ' ' {
				result.WriteRune(' ')
			}
		}
	}

	return result.String()
}

// readMultilineComment reads a multi-line comment block

func (p *Parser) readMultilineComment(comment *Comment, startLine string) {
	text := startLine

	// Check if comment ends on same line

	if strings.Contains(startLine, "*/") {

		text = strings.TrimPrefix(startLine, "/*")

		text = strings.TrimSuffix(text, "*/")

		comment.Text = strings.TrimSpace(text)

		return

	}

	// Read until closing */

	text = strings.TrimPrefix(startLine, "/*")

	for p.scan() {

		line := p.line

		text += "\n" + line

		if strings.Contains(line, "*/") {

			text = strings.TrimSuffix(text, "*/")

			break

		}

	}

	comment.Text = strings.TrimSpace(text)
}

// scan reads the next line

func (p *Parser) scan() bool {
	if p.eof {
		return false
	}

	if !p.scanner.Scan() {

		p.eof = true

		return false

	}

	p.line = p.scanner.Text()

	return true
}

// containsRealBrace checks if a line contains an actual opening brace (not from interpolation)
// This is used to detect multi-line nested blocks
func containsRealBrace(line string) bool {
	inInterpolation := false
	for i := 0; i < len(line); i++ {
		if line[i] == '@' && i+1 < len(line) && line[i+1] == '{' {
			inInterpolation = true
			i++ // Skip the '{'
		} else if inInterpolation && line[i] == '}' {
			inInterpolation = false
		} else if !inInterpolation && line[i] == '{' {
			return true
		}
	}
	return false
}

// findBlockBraces finds the opening and closing braces for a block in a line,
// skipping any braces that are part of @{...} interpolation patterns.
// Returns (-1, -1) if no actual block braces are found.
func findBlockBraces(line string) (int, int) {
	openIdx := -1
	inInterpolation := false

	// Find the opening brace (skipping @{...} patterns)
	for i := 0; i < len(line); i++ {
		if line[i] == '@' && i+1 < len(line) && line[i+1] == '{' {
			inInterpolation = true
			i++ // Skip the '{'
		} else if inInterpolation && line[i] == '}' {
			inInterpolation = false
		} else if !inInterpolation && line[i] == '{' {
			openIdx = i
			break
		}
	}

	if openIdx == -1 {
		return -1, -1
	}

	// Find the closing brace from the end (skipping @{...} patterns)
	inInterpolation = false
	closeIdx := -1
	for i := len(line) - 1; i > openIdx; i-- {
		if line[i] == '}' && !inInterpolation {
			closeIdx = i
			break
		} else if line[i] == '}' && inInterpolation {
			inInterpolation = false
		} else if i > 0 && line[i-1] == '@' && line[i] == '{' {
			inInterpolation = true
			i-- // Move back to skip the @
		}
	}

	if closeIdx == -1 {
		return -1, -1
	}

	return openIdx, closeIdx
}

// splitSelectorList splits a selector string by commas, respecting @{...} interpolation blocks and parentheses
func splitSelectorList(selectorStr string) []string {
	var result []string
	var current strings.Builder
	inInterpolation := false
	parenDepth := 0

	for i := 0; i < len(selectorStr); i++ {
		if selectorStr[i] == '@' && i+1 < len(selectorStr) && selectorStr[i+1] == '{' {
			inInterpolation = true
			current.WriteByte('@')
			current.WriteByte('{')
			i++ // Skip the '{'
		} else if inInterpolation && selectorStr[i] == '}' {
			inInterpolation = false
			current.WriteByte('}')
		} else if !inInterpolation && selectorStr[i] == '(' {
			parenDepth++
			current.WriteByte('(')
		} else if !inInterpolation && selectorStr[i] == ')' {
			parenDepth--
			current.WriteByte(')')
		} else if !inInterpolation && parenDepth == 0 && selectorStr[i] == ',' {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(selectorStr[i])
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// splitParameterList splits a parameter string by commas, respecting @{...} interpolation blocks and (...)
func splitParameterList(paramStr string) []string {
	var result []string
	var current strings.Builder
	inInterpolation := false
	parenDepth := 0

	for i := 0; i < len(paramStr); i++ {
		if paramStr[i] == '@' && i+1 < len(paramStr) && paramStr[i+1] == '{' {
			inInterpolation = true
			current.WriteByte('@')
			current.WriteByte('{')
			i++ // Skip the '{'
		} else if inInterpolation && paramStr[i] == '}' {
			inInterpolation = false
			current.WriteByte('}')
		} else if paramStr[i] == '(' {
			parenDepth++
			current.WriteByte('(')
		} else if paramStr[i] == ')' {
			parenDepth--
			current.WriteByte(')')
		} else if !inInterpolation && parenDepth == 0 && paramStr[i] == ',' {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(paramStr[i])
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
