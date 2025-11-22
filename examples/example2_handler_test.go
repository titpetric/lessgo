package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample2_CustomHandlerCompilation(t *testing.T) {
	handler := Example2_CustomHandler()

	// Test: Request to .less file should compile to CSS
	req := httptest.NewRequest("GET", "/app.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))

	css := w.Body.String()
	require.Contains(t, css, "color:")
	require.Contains(t, css, "#2c3e50") // main-color
}

func TestExample2_CustomHandlerNotFound(t *testing.T) {
	handler := Example2_CustomHandler()

	// Test: Non-existent file should return 404
	req := httptest.NewRequest("GET", "/nonexistent.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestExample2_CustomHandlerWrongFileType(t *testing.T) {
	handler := Example2_CustomHandler()

	// Test: Non-.less file should return 404
	req := httptest.NewRequest("GET", "/app.css", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestExample2_CustomHandlerMethodNotAllowed(t *testing.T) {
	handler := Example2_CustomHandler()

	// Test: POST request should fail
	req := httptest.NewRequest("POST", "/app.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestExample2_CustomHandlerHEADRequest(t *testing.T) {
	handler := Example2_CustomHandler()

	// Test: HEAD request should work
	req := httptest.NewRequest("HEAD", "/app.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
	// HEAD request should have no body
	require.Equal(t, "", w.Body.String())
}

func TestExample2_MuxHandler(t *testing.T) {
	// Use the handler directly with the testdata/styles directory
	handler := NewLessCompilerHandler("testdata/styles")

	// Test: CSS from /styles/ should compile LESS
	req := httptest.NewRequest("GET", "/theme.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
	css := w.Body.String()
	require.NotEmpty(t, css, "CSS output should not be empty")
}

func TestExample2_MuxHandlerIndex(t *testing.T) {
	mux := Example2_MuxWithHandler()

	// Test: Index should return HTML
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	require.Contains(t, w.Body.String(), "Custom Handler Example")
}

func TestExample2_HandlerVariableInterpolation(t *testing.T) {
	handler := Example2_CustomHandler()

	req := httptest.NewRequest("GET", "/app.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	css := w.Body.String()

	// Check that variables were interpolated
	require.Contains(t, css, "#2c3e50")   // main-color
	require.Contains(t, css, "1px solid") // border
	require.NotContains(t, css, "@main-color")
	require.NotContains(t, css, "@border")
}

func TestExample2_HandlerCaching(t *testing.T) {
	handler := Example2_CustomHandler()

	req := httptest.NewRequest("GET", "/app.less", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	require.Contains(t, cacheControl, "public")
	require.Contains(t, cacheControl, "max-age=3600")
}

func TestExample2_HandlerMultipleCompilations(t *testing.T) {
	handler := Example2_CustomHandler()

	// Test: Multiple requests should all compile correctly
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/app.less", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "#2c3e50")
	}
}
