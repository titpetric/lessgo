package lessgo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/renderer"
)

// TestFixtures tests all fixture files against lessc output
func TestFixtures(t *testing.T) {
	fixturesFS := os.DirFS("testdata/fixtures")
	entries, err := fs.ReadDir(fixturesFS, ".")
	if err != nil {
		t.Fatalf("failed to read fixtures directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".less") {
			continue
		}

		fixtureName := entry.Name()
		t.Run(fixtureName, func(t *testing.T) {
			fixturePath := filepath.Join("testdata/fixtures", fixtureName)
			testFixture(t, fixturePath)
		})
	}
}

// testFixture tests a single fixture file
func testFixture(t *testing.T, fixturePath string) {
	// Parse the .less file with lessgo
	file, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer file.Close()

	// Get the directory of the file for resolving imports
	dir := filepath.Dir(fixturePath)
	fileSystem := os.DirFS(dir)

	parser := dst.NewParserWithFS(file, fileSystem)
	astFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}

	// Render to CSS with lessgo
	lessgoRenderer := renderer.NewRenderer()
	lessgoCSS, err := lessgoRenderer.RenderWithBaseDir(astFile, dir)
	require.NoError(t, err)

	// Read expected CSS output from the .css file
	expectedCSS, err := readExpectedCSS(fixturePath)
	if err != nil {
		t.Fatalf("failed to read expected CSS: %v", err)
	}

	diff := cmp.Diff(expectedCSS, lessgoCSS)

	if diff != "" {
		t.Error(diff)
	}
}

// readExpectedCSS reads the expected CSS from the .css file adjacent to the .less file
func readExpectedCSS(lessPath string) (string, error) {
	cssPath := strings.TrimSuffix(lessPath, ".less") + ".css"
	data, err := os.ReadFile(cssPath)
	if err != nil {
		return "", fmt.Errorf("failed to read CSS file %s: %w", cssPath, err)
	}
	return string(data), nil
}
