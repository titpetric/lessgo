package importer

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/sourcegraph/lessgo/ast"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/stretchr/testify/require"
)

func TestExtractImportPath(t *testing.T) {
	tests := []struct {
		name   string
		params string
		want   string
	}{
		{
			name:   "quoted path",
			params: `"file.less"`,
			want:   "file.less",
		},
		{
			name:   "single quoted",
			params: `'file.less'`,
			want:   "file.less",
		},
		{
			name:   "url syntax",
			params: `url("file.less")`,
			want:   "file.less",
		},
		{
			name:   "url with spaces",
			params: `url( "file.less" )`,
			want:   "file.less",
		},
		{
			name:   "path with options",
			params: `"file.less" (reference)`,
			want:   "file.less",
		},
		{
			name:   "invalid - no quotes",
			params: ``,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractImportPath(tt.params)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseImportOptions(t *testing.T) {
	tests := []struct {
		name   string
		params string
		want   ImportOptions
	}{
		{
			name:   "reference option",
			params: `"file.less" (reference)`,
			want:   ImportOptions{Reference: true},
		},
		{
			name:   "inline option",
			params: `"file.less" (inline)`,
			want:   ImportOptions{Inline: true},
		},
		{
			name:   "optional option",
			params: `"file.less" (optional)`,
			want:   ImportOptions{Optional: true},
		},
		{
			name:   "no options",
			params: `"file.less"`,
			want:   ImportOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseImportOptions(tt.params)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestResolveImports(t *testing.T) {
	// Create a test filesystem
	fsys := fstest.MapFS{
		"imported.less": &fstest.MapFile{
			Data: []byte(`@imported-var: 10px;\n.imported { width: @imported-var; }`),
		},
		"missing.less": nil,
	}

	importer := New(fsys)

	t.Run("successful import", func(t *testing.T) {
		stylesheet := &ast.Stylesheet{
			Rules: []ast.Statement{
				&ast.AtRule{
					Name:       "import",
					Parameters: `"imported.less"`,
				},
			},
		}

		err := importer.ResolveImports(stylesheet, "main.less")
		require.NoError(t, err)
		// The imported rule should replace the @import statement
		require.Len(t, stylesheet.Rules, 2) // @imported-var and .imported rule
	})

	t.Run("missing import error", func(t *testing.T) {
		stylesheet := &ast.Stylesheet{
			Rules: []ast.Statement{
				&ast.AtRule{
					Name:       "import",
					Parameters: `"missing.less"`,
				},
			},
		}

		err := importer.ResolveImports(stylesheet, "main.less")
		require.Error(t, err)
		require.Contains(t, err.Error(), "import not found")
	})

	t.Run("optional missing import", func(t *testing.T) {
		stylesheet := &ast.Stylesheet{
			Rules: []ast.Statement{
				&ast.AtRule{
					Name:       "import",
					Parameters: `"missing.less" (optional)`,
				},
			},
		}

		err := importer.ResolveImports(stylesheet, "main.less")
		require.NoError(t, err)
		// Optional missing imports are silently ignored
		require.Len(t, stylesheet.Rules, 0)
	})
}

func TestResolveImportsWithRealFiles(t *testing.T) {
	// Use OS filesystem for more realistic testing
	tmpDir := t.TempDir()

	// Create test files
	mainFile := tmpDir + "/main.less"
	importedFile := tmpDir + "/imported.less"

	importedContent := `@color: red;
.card { background: @color; }`

	mainContent := `@import "imported.less";
.container { padding: 10px; }`

	err := os.WriteFile(importedFile, []byte(importedContent), 0644)
	require.NoError(t, err)

	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	require.NoError(t, err)

	// Parse the main file
	source, err := os.ReadFile(mainFile)
	require.NoError(t, err)

	lexer := parser.NewLexer(string(source))
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, err := p.Parse()
	require.NoError(t, err)

	// Resolve imports using OS filesystem
	importer := New(os.DirFS(tmpDir))
	err = importer.ResolveImports(stylesheet, "main.less")
	require.NoError(t, err)

	// Should have variable declaration and 2 rules
	require.Greater(t, len(stylesheet.Rules), 1)
}

func TestNestedImports(t *testing.T) {
	// Create a filesystem with nested imports
	fsys := fstest.MapFS{
		"main.less": &fstest.MapFile{
			Data: []byte(`@import "level1.less";`),
		},
		"level1.less": &fstest.MapFile{
			Data: []byte(`@import "level2.less";\n.level1 { color: blue; }`),
		},
		"level2.less": &fstest.MapFile{
			Data: []byte(`.level2 { color: green; }`),
		},
	}

	importer := New(fsys)

	// Parse main.less
	lexer := parser.NewLexer(`@import "level1.less";`)
	tokens := lexer.Tokenize()
	p := parser.NewParser(tokens)
	stylesheet, err := p.Parse()
	require.NoError(t, err)

	err = importer.ResolveImports(stylesheet, "main.less")
	require.NoError(t, err)

	// Should have resolved nested imports
	require.Len(t, stylesheet.Rules, 2) // level1 and level2 rules
}
