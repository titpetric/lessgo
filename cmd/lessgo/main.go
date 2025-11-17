package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/lessgo/formatter"
	"github.com/sourcegraph/lessgo/importer"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: lessgo <command> [args]\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  fmt <files>  Format LESS files\n")
		fmt.Fprintf(os.Stderr, "  compile <file>  Compile LESS to CSS\n")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "fmt":
		fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
		fmtCmd.Parse(os.Args[2:])

		files := fmtCmd.Args()
		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "Usage: lessgo fmt <files...>\n")
			os.Exit(1)
		}

		if err := formatFiles(files); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "compile":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: lessgo compile <file>\n")
			os.Exit(1)
		}

		if err := compileFile(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

// formatFiles formats one or more LESS files
func formatFiles(patterns []string) error {
	for _, pattern := range patterns {
		// Expand glob patterns
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("no files matching %q", pattern)
		}

		for _, filepath := range matches {
			if !strings.HasSuffix(filepath, ".less") {
				fmt.Printf("Skipping non-.less file: %s\n", filepath)
				continue
			}

			if err := formatFile(filepath); err != nil {
				return fmt.Errorf("failed to format %s: %w", filepath, err)
			}

			fmt.Printf("Formatted: %s\n", filepath)
		}
	}

	return nil
}

// compileFile reads, parses, compiles LESS to CSS and prints to stdout
func compileFile(filePath string) error {
	// Read file
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	sourceStr := string(source)

	// Parse LESS
	lexer := parser.NewLexer(sourceStr)
	tokens := lexer.Tokenize()

	p := parser.NewParserWithSource(tokens, sourceStr)
	stylesheet, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Resolve imports
	dir := filepath.Dir(filePath)
	basename := filepath.Base(filePath)
	imp := importer.New(os.DirFS(dir))
	if err := imp.ResolveImports(stylesheet, basename); err != nil {
		return fmt.Errorf("import error: %w", err)
	}

	// Render to CSS
	r := renderer.NewRenderer()
	css := r.Render(stylesheet)

	// Print to stdout
	fmt.Print(css)
	return nil
}

// formatFile reads, parses, formats, and writes back a LESS file
func formatFile(filePath string) error {
	// Read file
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	sourceStr := string(source)

	// Parse LESS
	lexer := parser.NewLexer(sourceStr)
	tokens := lexer.Tokenize()

	p := parser.NewParserWithSource(tokens, sourceStr)
	stylesheet, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Resolve imports - this will error if any import is not found
	dir := filepath.Dir(filePath)
	basename := filepath.Base(filePath)
	imp := importer.New(os.DirFS(dir))
	if err := imp.ResolveImports(stylesheet, basename); err != nil {
		return fmt.Errorf("import error: %w", err)
	}

	// Format with double-space indentation
	fmt := formatter.New(2) // 2 spaces indentation
	formatted := fmt.Format(stylesheet)

	// Write back
	return ioutil.WriteFile(filePath, []byte(formatted), 0644)
}
