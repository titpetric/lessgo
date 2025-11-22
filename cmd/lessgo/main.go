package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/renderer"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "ast":
		astCmd(os.Args[2:])
	case "fmt":
		fmtCmd(os.Args[2:])
	case "generate":
		generateCmd(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func astCmd(args []string) {
	fs := flag.NewFlagSet("ast", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: lessgo ast <file.less>\n")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	filePath := fs.Arg(0)

	// Parse the .less file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Get the directory of the file for resolving imports
	dir := filepath.Dir(filePath)
	if dir == "" {
		dir = "."
	}
	fileSystem := os.DirFS(dir)

	parser := dst.NewParserWithFS(file, fileSystem)
	astFile, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing file: %v\n", err)
		os.Exit(1)
	}

	dst.Print(astFile)
}

func fmtCmd(args []string) {
	fs := flag.NewFlagSet("fmt", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: lessgo fmt [options] <file.less>\n")
		fs.PrintDefaults()
	}

	write := fs.Bool("w", false, "write formatted output back to file")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	filePath := fs.Arg(0)

	// Parse the .less file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Get the directory of the file for resolving imports
	dir := filepath.Dir(filePath)
	if dir == "" {
		dir = "."
	}
	fileSystem := os.DirFS(dir)

	parser := dst.NewParserWithFS(file, fileSystem)
	astFile, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing file: %v\n", err)
		os.Exit(1)
	}

	// Format the AST
	formatter := dst.NewFormatter()
	formatted := formatter.Format(astFile)

	if *write {
		// Write back to file
		err := os.WriteFile(filePath, []byte(formatted), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("formatted: %s\n", filePath)
	} else {
		// Print to stdout
		fmt.Print(formatted)
	}
}

func generateCmd(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: lessgo generate [options] <glob-pattern>\n")
		fs.PrintDefaults()
	}

	output := fs.String("o", "", "output file (default: stdout)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	pattern := fs.Arg(0)

	// Find .less files matching pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error matching pattern: %v\n", err)
		os.Exit(1)
	}

	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "no files matching: %s\n", pattern)
		os.Exit(1)
	}

	// Generate CSS output for all matched files
	var allCSS string

	for _, filePath := range matches {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening %s: %v\n", filePath, err)
			continue
		}

		// Get the directory of the file for resolving imports
		dir := filepath.Dir(filePath)
		if dir == "" {
			dir = "."
		}
		fileSystem := os.DirFS(dir)

		parser := dst.NewParserWithFS(file, fileSystem)
		astFile, err := parser.Parse()
		file.Close()

		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing %s: %v\n", filePath, err)
			continue
		}

		// Render to CSS
		cssRenderer := renderer.NewRenderer()
		css, err := cssRenderer.Render(astFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error rendering %s: %v\n", filePath, err)
			continue
		}

		allCSS += fmt.Sprintf("%s\n", css)
	}

	if *output != "" {
		// Write to output file
		err := os.WriteFile(*output, []byte(allCSS), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated: %s\n", *output)
	} else {
		// Print to stdout
		fmt.Print(allCSS)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `usage: lessgo <command> [options]

commands:
  fmt       - Format .less files for consistent indentation
  generate  - Generate CSS files from glob pattern of .less files
  help      - Show this help message

examples:
  lessgo fmt style.less
  lessgo fmt -w style.less
  lessgo generate "**/*.less" -o all.css
`)
}
