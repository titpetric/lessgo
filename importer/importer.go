// Package importer handles @import resolution and processing
package importer

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/lessgo/ast"
	"github.com/sourcegraph/lessgo/parser"
)

// Importer resolves and processes @import statements
type Importer struct {
	fs fs.FS // filesystem to search for imports
}

// New creates a new Importer with the given filesystem
func New(filesystem fs.FS) *Importer {
	return &Importer{
		fs: filesystem,
	}
}

// ResolveImports processes @import statements in a stylesheet,
// resolving imported files and merging them into the stylesheet.
// It returns an error if any import cannot be found.
func (imp *Importer) ResolveImports(stylesheet *ast.Stylesheet, basePath string) error {
	resolved := []ast.Statement{}

	for _, stmt := range stylesheet.Rules {
		atRule, ok := stmt.(*ast.AtRule)
		if !ok || atRule.Name != "import" {
			resolved = append(resolved, stmt)
			continue
		}

		// Parse @import statement
		importedStmts, err := imp.resolveImport(atRule, basePath)
		if err != nil {
			return err
		}

		resolved = append(resolved, importedStmts...)
	}

	stylesheet.Rules = resolved
	return nil
}

// resolveImport processes a single @import statement
func (imp *Importer) resolveImport(atRule *ast.AtRule, basePath string) ([]ast.Statement, error) {
	// Parse import path and options from parameters
	// Example: @import "file.less";
	// Example: @import url("file.css");
	// Example: @import "file.less" (reference);

	params := strings.TrimSpace(atRule.Parameters)
	if params == "" {
		return nil, fmt.Errorf("empty import statement")
	}

	// Extract the import path
	importPath := extractImportPath(params)
	if importPath == "" {
		return nil, fmt.Errorf("invalid import syntax: %s", params)
	}

	// Check for import options
	options := parseImportOptions(params)

	// Resolve the file path relative to basePath
	// basePath should be a relative path from the fs root
	baseDir := filepath.Dir(basePath)
	if baseDir == "." || baseDir == "" {
		baseDir = ""
	} else {
		baseDir = baseDir + string(filepath.Separator)
	}
	resolvedPath := filepath.Join(baseDir, importPath)
	// Normalize the path for os.DirFS compatibility
	resolvedPath = filepath.ToSlash(resolvedPath)

	// Try to read the imported file
	content, err := fs.ReadFile(imp.fs, resolvedPath)
	if err != nil {
		if options.Optional {
			// Optional imports that don't exist are silently ignored
			return nil, nil
		}
		return nil, fmt.Errorf("import not found: %q (resolved as %q): %w", importPath, resolvedPath, err)
	}

	// Parse the imported content
	lexer := parser.NewLexer(string(content))
	tokens := lexer.Tokenize()

	p := parser.NewParser(tokens)
	importedStylesheet, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse imported file %q: %w", importPath, err)
	}

	// Process nested imports recursively
	if err := imp.ResolveImports(importedStylesheet, resolvedPath); err != nil {
		return nil, fmt.Errorf("error processing imports in %q: %w", importPath, err)
	}

	// Handle import options
	if options.Reference {
		// Reference imports are not output to CSS
		// but their mixins can be used
		// For now, we just skip them
		return nil, nil
	}

	if options.Inline {
		// Inline the content as-is
		return importedStylesheet.Rules, nil
	}

	if options.CSS {
		// Import as CSS (wrap in @import rule for output)
		return importedStylesheet.Rules, nil
	}

	// Default: treat as LESS and merge
	return importedStylesheet.Rules, nil
}

// extractImportPath extracts the file path from an @import statement
func extractImportPath(params string) string {
	params = strings.TrimSpace(params)

	// Handle url("...") syntax
	if strings.HasPrefix(params, "url(") {
		start := strings.Index(params, "(") + 1
		end := strings.LastIndex(params, ")")
		if end > start {
			path := strings.TrimSpace(params[start:end])
			// Remove quotes
			path = strings.Trim(path, `"'`)
			return path
		}
	}

	// Handle quoted path: "file.less" or 'file.less'
	if strings.HasPrefix(params, `"`) || strings.HasPrefix(params, `'`) {
		quote := params[0]
		end := strings.Index(params[1:], string(quote))
		if end >= 0 {
			return params[1 : end+1]
		}
	}

	// Handle unquoted path (less common)
	parts := strings.Fields(params)
	if len(parts) > 0 {
		path := parts[0]
		path = strings.Trim(path, `"'`)
		return path
	}

	return ""
}

// ImportOptions represents @import directive options
type ImportOptions struct {
	Reference bool
	Inline    bool
	Less      bool
	CSS       bool
	Once      bool
	Multiple  bool
	Optional  bool
}

// parseImportOptions extracts import options from parameters
func parseImportOptions(params string) ImportOptions {
	opts := ImportOptions{}

	// Check for (option) syntax at the end
	if strings.Contains(params, "(") && strings.Contains(params, ")") {
		start := strings.LastIndex(params, "(")
		end := strings.LastIndex(params, ")")
		if start < end {
			optStr := params[start+1 : end]
			optStr = strings.TrimSpace(optStr)
			optStr = strings.ToLower(optStr)

			switch optStr {
			case "reference":
				opts.Reference = true
			case "inline":
				opts.Inline = true
			case "less":
				opts.Less = true
			case "css":
				opts.CSS = true
			case "once":
				opts.Once = true
			case "multiple":
				opts.Multiple = true
			case "optional":
				opts.Optional = true
			}
		}
	}

	return opts
}
