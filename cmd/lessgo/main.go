package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/parser"
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

	// Parse LESS with DST
	lexer := parser.NewLexer(sourceStr)
	tokens := lexer.Tokenize()

	dstParser := dst.NewParser(tokens, sourceStr)
	doc, err := dstParser.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Render to CSS
	r := dst.NewRenderer()
	css := r.Render(doc)

	// Print to stdout
	fmt.Print(css)
	return nil
}

// formatFile reads, parses, formats, and writes back a LESS file
func formatFile(filePath string) error {
	// For now, just return an error - formatting will be implemented later
	return fmt.Errorf("formatting not yet implemented with DST parser")
}
