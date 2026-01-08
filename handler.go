package lessgo

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/internal/strings"
	"github.com/titpetric/lessgo/renderer"
)

// Error types for LESS compilation and serving
var (
	ErrNotFound          = errors.New("not found")
	ErrCompilationFailed = errors.New("compilation failed")
)

// Handler handles LESS file compilation and serving
type Handler struct {
	pathPrefix string
	fileSystem fs.FS
}

// NewHandler creates a new LESS compilation handler.
// fileSystem is where to read .less files from
// pathPrefix is the URL path prefix to match and strip (e.g., "/assets/css")
func NewHandler(fileSystem fs.FS, pathPrefix string) http.Handler {
	return &Handler{
		pathPrefix: pathPrefix,
		fileSystem: fileSystem,
	}
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle GET and HEAD requests
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if request path starts with pathPrefix and ends with .less
	if h.pathPrefix != "" && !strings.HasPrefix(r.URL.Path, h.pathPrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if !strings.HasSuffix(r.URL.Path, ".less") {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Extract the relative path within the prefix
	lessPath := strings.TrimPrefix(r.URL.Path, h.pathPrefix)
	// If pathPrefix is "/", don't remove leading slash again
	if h.pathPrefix != "/" {
		lessPath = strings.TrimPrefix(lessPath, "/")
	}

	// Check if file exists
	info, err := fs.Stat(h.fileSystem, lessPath)
	if err != nil || info.IsDir() {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Open and read the LESS file
	file, err := h.fileSystem.Open(lessPath)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	var astFile *dst.File

	// Parse the LESS file using configured parser
	if dst.DefaultParserConfig.UseNoAlloc {
		parser := dst.NewParserNoAllocWithFS(file, h.fileSystem)
		astFile, err = parser.Parse()
	} else {
		parser := dst.NewParserWithFS(file, h.fileSystem)
		astFile, err = parser.Parse()
	}
	if err != nil {
		http.Error(w, "Compilation Error", http.StatusInternalServerError)
		return
	}

	// Render to CSS
	cssRenderer := renderer.NewRenderer()
	css, err := cssRenderer.Render(astFile)
	if err != nil {
		http.Error(w, "Compilation Error", http.StatusInternalServerError)
		return
	}

	// Send the compiled CSS
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if r.Method != http.MethodHead {
		w.Write([]byte(css))
	}
}
