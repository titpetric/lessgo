package testdata_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sourcegraph/lessgo/parser"
	"github.com/sourcegraph/lessgo/renderer"
)

func TestFixtures(t *testing.T) {
	// Find all .less files in fixtures directory
	fixturesDir := "fixtures"
	entries, err := ioutil.ReadDir(fixturesDir)
	require.NoError(t, err, "failed to read fixtures directory")

	// Group by fixture name
	fixtures := make(map[string]map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		baseName := strings.TrimSuffix(name, ext)

		if fixtures[baseName] == nil {
			fixtures[baseName] = make(map[string]string)
		}

		path := filepath.Join(fixturesDir, name)
		content, err := ioutil.ReadFile(path)
		require.NoError(t, err, "failed to read %s", name)

		// Store by extension (without the dot)
		fixtures[baseName][strings.TrimPrefix(ext, ".")] = string(content)
	}

	// Test each fixture
	for fixtureName, files := range fixtures {
		t.Run(fixtureName, func(t *testing.T) {
			less, ok := files["less"]
			require.True(t, ok, "missing .less file for fixture %s", fixtureName)

			expected, ok := files["css"]
			require.True(t, ok, "missing .css file for fixture %s", fixtureName)

			// Parse and compile
			compiled, err := compileLESS(less)
			require.NoError(t, err, "failed to compile LESS")

			// Normalize whitespace for comparison
			compiledNorm := normalizeCSS(compiled)
			expectedNorm := normalizeCSS(expected)

			require.Equal(t, expectedNorm, compiledNorm,
				"compiled CSS does not match expected output for fixture %s", fixtureName)
		})
	}
}

// compileLESS takes LESS source and returns compiled CSS
func compileLESS(lessSource string) (string, error) {
	// Tokenize
	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()

	// Parse
	p := parser.NewParser(tokens)
	stylesheet, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// Render
	r := renderer.NewRenderer()
	css := r.Render(stylesheet)

	return css, nil
}

// normalizeCSS normalizes CSS for comparison by removing extra whitespace
func normalizeCSS(css string) string {
	// Remove leading/trailing whitespace
	css = strings.TrimSpace(css)

	// Replace multiple newlines with single newline
	for strings.Contains(css, "\n\n") {
		css = strings.ReplaceAll(css, "\n\n", "\n")
	}

	// Trim each line
	lines := strings.Split(css, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	return strings.Join(lines, "\n")
}
