package examples

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/titpetric/lessgo"
	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/renderer"
)

// Example2_Handler demonstrates using lessgo as a custom http.Handler
//
// This example shows how to create a custom handler that compiles LESS files
// with more control over the request handling and response.
type LessCompilerHandler struct {
	BaseDir    string // Directory containing .less files (for os.Open)
	FileSystem fs.FS  // FileSystem for reading LESS files
}

// ServeHTTP implements http.Handler interface
func (h *LessCompilerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the requested file path
	lessFile := filepath.Join(h.BaseDir, filepath.Clean(r.URL.Path))

	// Security: prevent path traversal
	if !filepath.HasPrefix(lessFile, h.BaseDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if file exists and is a .less file
	if !h.isValidLessFile(lessFile) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Compile the LESS file
	css, err := h.compileLess(lessFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Compilation error: %v", err), http.StatusInternalServerError)
		return
	}

	// Send the compiled CSS
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if r.Method != http.MethodHead {
		w.Write([]byte(css))
	}
}

// isValidLessFile checks if the file exists and is a .less file
func (h *LessCompilerHandler) isValidLessFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}

	// Check file extension
	if filepath.Ext(filePath) != ".less" {
		return false
	}

	return true
}

// compileLess compiles a LESS file to CSS
func (h *LessCompilerHandler) compileLess(lessPath string) (string, error) {
	// Open the LESS file
	file, err := os.Open(lessPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create filesystem for resolving imports
	dir := filepath.Dir(lessPath)
	fileSystem := os.DirFS(dir)

	// Parse the LESS file
	parser := dst.NewParserWithFS(file, fileSystem)
	astFile, err := parser.Parse()
	if err != nil {
		return "", fmt.Errorf("failed to parse: %w", err)
	}

	// Render to CSS
	cssRenderer := renderer.NewRenderer()
	css, err := cssRenderer.Render(astFile)
	if err != nil {
		return "", fmt.Errorf("failed to render: %w", err)
	}

	return css, nil
}

// NewLessCompilerHandler creates a new LESS compiler handler
// If fileSystem is nil, it will be created from baseDir using os.DirFS
func NewLessCompilerHandler(baseDir string, optionalFS ...fs.FS) *LessCompilerHandler {
	handler := &LessCompilerHandler{
		BaseDir: baseDir,
	}

	// Use provided fs.FS or default to os.DirFS(baseDir)
	if len(optionalFS) > 0 && optionalFS[0] != nil {
		handler.FileSystem = optionalFS[0]
	} else {
		handler.FileSystem = os.DirFS(baseDir)
	}

	return handler
}

// Example2_CustomHandler demonstrates using the custom handler
func Example2_CustomHandler() http.Handler {
	// Create a custom handler that compiles LESS files from a specific directory
	return lessgo.NewHandler(os.DirFS("testdata/custom"), "/")
}

// Example2_MuxWithHandler demonstrates using multiple handlers with http.ServeMux
func Example2_MuxWithHandler() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve LESS files from /styles
	mux.Handle("/styles/", NewLessCompilerHandler("testdata/styles"))

	// Serve other assets
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
  <link rel="stylesheet" href="/styles/theme.less">
</head>
<body>
  <h1>Custom Handler Example</h1>
</body>
</html>
`))
	}))

	return mux
}
