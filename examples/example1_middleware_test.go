package examples

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample1_MiddlewareCompilation(t *testing.T) {
	// Create the handler from the example
	handler := Example1_Middleware()

	// Test 1: Request to .less file should compile to CSS
	req := httptest.NewRequest("GET", "/assets/css/style.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))

	css := w.Body.String()
	require.Contains(t, css, "color:")
	require.Contains(t, css, "#3498db") // primary-color value
}

func TestExample1_MiddlewarePassthrough(t *testing.T) {
	handler := Example1_Middleware()

	// Test: Request to HTML should pass through to main handler
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	require.Contains(t, w.Body.String(), "Hello World")
}

func TestExample1_MiddlewareNotFound(t *testing.T) {
	handler := Example1_Middleware()

	// Test: Request to non-existent .less file should fail
	req := httptest.NewRequest("GET", "/assets/css/nonexistent.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should pass through to main handler which returns HTML
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestExample1_MiddlewareHEADRequest(t *testing.T) {
	handler := Example1_Middleware()

	// Test: HEAD request should work
	req := httptest.NewRequest("HEAD", "/assets/css/style.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
	// HEAD request should have no body
	require.Equal(t, "", w.Body.String())
}

func TestExample1_MiddlewareMultipleRequests(t *testing.T) {
	handler := Example1_Middleware()

	// Test: Multiple requests to compile the same file should work
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/assets/css/style.less", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
		require.NotEmpty(t, w.Body.String())
	}
}

func TestExample1_MiddlewareChain(t *testing.T) {
	// Create a base handler
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Base handler"))
	})

	// Wrap with the chain
	handler := Example1_MiddlewareWithChain(baseHandler)

	// Test: Non-.less requests should pass through
	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Base handler")
}

func TestExample1_MiddlewareCaching(t *testing.T) {
	handler := Example1_Middleware()

	// Test: Response should include cache headers
	req := httptest.NewRequest("GET", "/assets/css/style.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	require.Contains(t, cacheControl, "public")
	require.Contains(t, cacheControl, "max-age=3600")
}

func TestExample1_MiddlewareVariableInterpolation(t *testing.T) {
	handler := Example1_Middleware()

	req := httptest.NewRequest("GET", "/assets/css/style.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	css := w.Body.String()

	// Check that variables were interpolated
	require.Contains(t, css, "#3498db") // primary-color
	require.Contains(t, css, "14px")    // font-size
	require.NotContains(t, css, "@primary-color")
}

func TestExample1_MiddlewareColorFunctions(t *testing.T) {
	handler := Example1_Middleware()

	req := httptest.NewRequest("GET", "/assets/css/style.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	css := w.Body.String()

	// Check that color functions were applied
	// darken(@primary-color, 20%) should produce a darker color
	require.NotContains(t, css, "darken(")
	require.NotContains(t, css, "@primary-color")

	// Should have hex color values (not the function call)
	hexCount := strings.Count(css, "#")
	require.Greater(t, hexCount, 1, "Should have multiple hex colors after darken function")
}
