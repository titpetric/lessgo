package lessgo

import (
	"io/fs"
	"net/http"
	"strings"
)

// NewMiddleware creates an HTTP middleware that compiles .less files to CSS on-the-fly.
// It intercepts requests to files with .less extension, compiles them using lessgo,
// and returns the resulting CSS with the appropriate Content-Type header.
//
// Parameters:
//   - basePath: The URL path prefix to match (e.g., "/assets/css")
//   - fileSystem: The filesystem to read .less files from (e.g., os.DirFS("./assets/css"))
//
// Example usage with chi:
//
//	chi.Use(lessgo.NewMiddleware("/assets/css", os.DirFS("./assets/css")))
//
// When a request to /assets/css/style.less is made, it will:
// 1. Check if the request path matches basePath and ends with .less
// 2. Read the file from the provided filesystem
// 3. Parse and compile it from LESS to CSS
// 4. Return the compiled CSS with Content-Type: text/css
// 5. If the file is not .less or doesn't exist, pass to next handler
func NewMiddleware(basePath string, fileSystem fs.FS) func(http.Handler) http.Handler {
	// Handle the .less request
	handler := NewHandler(basePath, fileSystem)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only handle GET and HEAD requests
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// Check if request path starts with basePath and ends with .less
			if !strings.HasPrefix(r.URL.Path, basePath) {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasSuffix(r.URL.Path, ".less") {
				next.ServeHTTP(w, r)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}
