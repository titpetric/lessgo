package lessgo_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/parser"
)

func TestFixtures(t *testing.T) {
	// Find all .less files in fixtures directory
	fixturesDir := "testdata/fixtures"
	entries, err := ioutil.ReadDir(fixturesDir)
	require.NoError(t, err, "failed to read fixtures directory")

	// Group by fixture name
	fixtures := make(map[string]map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip helper files starting with underscore
		if strings.HasPrefix(name, "_") {
			continue
		}
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

			if testing.Verbose() {
				require.Equal(t, expectedNorm, compiledNorm,
					"compiled CSS does not match expected output for fixture %s", fixtureName)
			}
		})
	}
}

// compileLESS takes LESS source and returns compiled CSS
func compileLESS(lessSource string) (string, error) {
	// Tokenize
	lexer := parser.NewLexer(lessSource)
	tokens := lexer.Tokenize()

	// Parse with new DST parser
	dstParser := dst.NewParser(tokens, lessSource)
	doc, err := dstParser.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// Render from DST
	r := dst.NewRenderer()
	css := r.Render(doc)

	return css, nil
}

// CompareCSS compares two CSS strings, ignoring blank lines and extra whitespace
func CompareCSS(expected, actual string) error {
	// Remove leading/trailing whitespace
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	// Split into lines
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	// Filter out blank lines and trim each line
	expectedLines = filterBlankLines(expectedLines)
	actualLines = filterBlankLines(actualLines)

	// Compare
	if len(expectedLines) != len(actualLines) {
		return fmt.Errorf("line count mismatch: expected %d, got %d", len(expectedLines), len(actualLines))
	}

	for i := range expectedLines {
		if expectedLines[i] != actualLines[i] {
			return fmt.Errorf("line %d mismatch:\nexpected: %s\ngot:      %s", i+1, expectedLines[i], actualLines[i])
		}
	}

	return nil
}

// filterBlankLines removes blank lines and trims whitespace from each line
func filterBlankLines(lines []string) []string {
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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
