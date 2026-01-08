package examples

import (
	"net/http"
	"os"

	"github.com/titpetric/lessgo"
)

// Example1_Middleware demonstrates using lessgo as a middleware with standard http.Handler
//
// This example shows how to compile LESS to CSS on-the-fly using the middleware pattern.
// Requests to .less files matching the path prefix are automatically compiled to CSS.
func Example1_Middleware() http.Handler {
	// Create a middleware that intercepts requests to /assets/css/*.less
	// and compiles them to CSS on-the-fly
	lessMiddleware := lessgo.NewMiddleware(os.DirFS("testdata/assets/css"), "/assets/css")

	// Wrap your main handler with the LESS middleware
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
  <link rel="stylesheet" href="/assets/css/style.less">
</head>
<body>
  <h1>Hello World</h1>
  <p>This page uses compiled LESS stylesheets.</p>
</body>
</html>
`))
	})

	// Apply the LESS middleware to your handler
	return lessMiddleware(mainHandler)
}

// Example1_MiddlewareWithChain demonstrates chaining multiple middleware
func Example1_MiddlewareWithChain(h http.Handler) http.Handler {
	// Create the LESS middleware
	lessMiddleware := lessgo.NewMiddleware(os.DirFS("testdata/styles"), "/styles")

	// Apply it to your handler
	h = lessMiddleware(h)

	return h
}
