package lessgo

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

// TestMiddlewarePassthrough tests that non-.less requests pass through
func TestMiddlewarePassthrough(t *testing.T) {
	mockFS := fstest.MapFS{
		"style.less": &fstest.MapFile{Data: []byte("body { color: red; }")},
	}

	middleware := NewMiddleware("/assets/css", mockFS)

	// Create a simple next handler that we can verify is called
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("next handler"))
	})

	handler := middleware(next)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "CSS file (non-.less) should pass through",
			path:       "/assets/css/style.css",
			wantStatus: http.StatusTeapot,
			wantBody:   "next handler",
		},
		{
			name:       "Request without basePath should pass through",
			path:       "/other/style.less",
			wantStatus: http.StatusTeapot,
			wantBody:   "next handler",
		},
		{
			name:       "POST request should pass through",
			path:       "/assets/css/style.less",
			wantStatus: http.StatusTeapot,
			wantBody:   "next handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.name == "POST request should pass through" {
				req = httptest.NewRequest(http.MethodPost, tt.path, nil)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			require.True(t, nextCalled, "next handler should be called")
			require.Equal(t, tt.wantStatus, w.Code)
			require.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}

// TestMiddlewareLessCompilation tests successful .less file compilation
func TestMiddlewareLessCompilation(t *testing.T) {
	mockFS := fstest.MapFS{
		"style.less": &fstest.MapFile{Data: []byte(`
@primary: #0066cc;
body {
  color: @primary;
}
`)},
	}

	middleware := NewMiddleware("/assets/css", mockFS)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/assets/css/style.less", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
	require.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
	require.Contains(t, w.Body.String(), "color: #0066cc")
	require.Contains(t, w.Body.String(), "body")
}

// TestMiddlewareNotFound tests 404 handling
func TestMiddlewareNotFound(t *testing.T) {
	mockFS := fstest.MapFS{}

	middleware := NewMiddleware("/assets/css", mockFS)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/assets/css/nonexistent.less", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.False(t, nextCalled)
	require.Equal(t, http.StatusNotFound, w.Code)
}

// TestMiddlewareHEADRequest tests HEAD request handling
func TestMiddlewareHEADRequest(t *testing.T) {
	mockFS := fstest.MapFS{
		"style.less": &fstest.MapFile{Data: []byte("body { color: red; }")},
	}

	middleware := NewMiddleware("/assets/css", mockFS)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	req := httptest.NewRequest(http.MethodHead, "/assets/css/style.less", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
	// HEAD request should not have body
	require.Equal(t, "", w.Body.String())
}

// TestMiddlewareVariableInterpolation tests LESS features work through middleware
func TestMiddlewareVariableInterpolation(t *testing.T) {
	mockFS := fstest.MapFS{
		"variables.less": &fstest.MapFile{Data: []byte(`
@color: #ff0000;
@size: 16px;

body {
  color: @color;
  font-size: @size;
}

.button {
  color: @color;
  padding: 10px;
}
`)},
	}

	middleware := NewMiddleware("/assets/css", mockFS)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/assets/css/variables.less", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	css := w.Body.String()
	require.Contains(t, css, "#ff0000")
	require.Contains(t, css, "16px")
	require.Contains(t, css, ".button")
	require.Contains(t, css, "body")
}

// TestMiddlewareNesting tests nested selectors work through middleware
func TestMiddlewareNesting(t *testing.T) {
	mockFS := fstest.MapFS{
		"nested.less": &fstest.MapFile{Data: []byte(`
.container {
  background: white;
  
  .header {
    color: blue;
    
    h1 {
      font-size: 24px;
    }
  }
}
`)},
	}

	middleware := NewMiddleware("/assets/css", mockFS)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/assets/css/nested.less", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	css := w.Body.String()
	require.Contains(t, css, ".container")
	require.Contains(t, css, ".container .header")
	require.Contains(t, css, ".container .header h1")
}

// TestMiddlewareBasePathWithoutSlash tests base path handling
func TestMiddlewareBasePathWithoutSlash(t *testing.T) {
	lessContent := []byte(`
.button {
  color: red;
}
`)
	mockFS := fstest.MapFS{
		"style.less": &fstest.MapFile{Data: lessContent},
	}

	middleware := NewMiddleware("/assets/css", mockFS)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/assets/css/style.less", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	css := w.Body.String()
	require.NotEmpty(t, css, "CSS output should not be empty")
	require.Contains(t, css, ".button")
}

// TestMiddlewareNestedDirectory tests files in subdirectories
func TestMiddlewareNestedDirectory(t *testing.T) {
	lessContent := []byte(`
.btn {
  color: blue;
  padding: 10px;
}
`)
	mockFS := fstest.MapFS{
		"components/button.less": &fstest.MapFile{Data: lessContent},
	}

	middleware := NewMiddleware("/css", mockFS)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/css/components/button.less", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	css := w.Body.String()
	require.NotEmpty(t, css, "CSS output should not be empty")
	require.Contains(t, css, ".btn")
}

// BenchmarkMiddleware benchmarks the middleware compilation and serving of LESS files
func BenchmarkMiddleware(b *testing.B) {
	lessContent, err := readFixture("testdata/fixtures/999-docker-ljubljana-index.less")
	if err != nil {
		b.Fatalf("failed to read fixture: %v", err)
	}

	mockFS := fstest.MapFS{
		"style.less": &fstest.MapFile{Data: lessContent},
	}

	middleware := NewMiddleware("/assets/css", mockFS)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := middleware(next)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/assets/css/style.less", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// readFixture reads a fixture file from disk
func readFixture(path string) ([]byte, error) {
	return os.ReadFile(path)
}
