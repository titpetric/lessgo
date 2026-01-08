# Package lessgo

```go
import (
	"github.com/titpetric/lessgo"
}
```

Package lessgo is a comprehensive Less CSS compiler implementation in pure Go.

## Types

```go
// Handler handles LESS file compilation and serving
type Handler struct {
	pathPrefix string
	fileSystem fs.FS
}
```

## Vars

```go
// Error types for LESS compilation and serving
var (
	ErrNotFound          = errors.New("not found")
	ErrCompilationFailed = errors.New("compilation failed")
)
```

## Function symbols

- `func NewHandler (fileSystem fs.FS, pathPrefix string) http.Handler`
- `func NewMiddleware (fileSystem fs.FS, basePath string) func(http.Handler) http.Handler`
- `func (*Handler) ServeHTTP (w http.ResponseWriter, r *http.Request)`

### NewHandler

NewHandler creates a new LESS compilation handler. fileSystem is where to read .less files from pathPrefix is the URL path prefix to match and strip (e.g., "/assets/css")

```go
func NewHandler(fileSystem fs.FS, pathPrefix string) http.Handler
```

### NewMiddleware

NewMiddleware creates an HTTP middleware that compiles .less files to CSS on-the-fly. It intercepts requests to files with .less extension, compiles them using lessgo, and returns the resulting CSS with the appropriate Content-Type header.

Parameters:
- fileSystem: The filesystem to read .less files from (e.g., os.DirFS("./assets/css"))
- basePath: The URL path prefix to match (e.g., "/assets/css")

Example usage with chi:

```
chi.Use(lessgo.NewMiddleware("/assets/css", os.DirFS("./assets/css")))
```

When a request to /assets/css/style.less is made, it will:
1. Check if the request path matches basePath and ends with .less
2. Read the file from the provided filesystem
3. Parse and compile it from LESS to CSS
4. Return the compiled CSS with Content-Type: text/css
5. If the file is not .less or doesn't exist, pass to next handler

```go
func NewMiddleware(fileSystem fs.FS, basePath string) func(http.Handler) http.Handler
```

### ServeHTTP

ServeHTTP implements http.Handler

```go
func (*Handler) ServeHTTP(w http.ResponseWriter, r *http.Request)
```
