package swagger

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleSpec = `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0"},"paths":{}}`

func TestHandlerIndexAndDoc(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/swagger/", Handler([]byte(sampleSpec), Title("Demo API")))

	t.Run("redirect root to index", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		assert.Equal(t, "/swagger/index.html", rec.Header().Get("Location"))
	})

	t.Run("serve index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
		body, _ := io.ReadAll(rec.Body)
		assert.Contains(t, string(body), "SwaggerUIBundle")
		assert.Contains(t, string(body), `url: "doc.json"`)
		assert.Contains(t, string(body), "<title>Demo API</title>")
	})

	t.Run("serve doc.json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
		body, _ := io.ReadAll(rec.Body)
		assert.JSONEq(t, sampleSpec, string(body))
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger/missing.js", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/swagger/index.html", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}

func TestHandlerFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oas.json")
	require.NoError(t, os.WriteFile(path, []byte(sampleSpec), 0o644))

	h, err := HandlerFromFile(path)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/swagger/", h)

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.JSONEq(t, sampleSpec, string(body))
}

func TestHandlerCustomURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/swagger/", Handler([]byte(sampleSpec), URL("oas.json")))

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), `url: "oas.json"`)

	req = httptest.NewRequest(http.MethodGet, "/swagger/oas.json", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandlerTrailingSlashOnFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/swagger/", Handler([]byte(sampleSpec)))

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "SwaggerUIBundle")

	req = httptest.NewRequest(http.MethodGet, "/swagger/doc.json/", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body, _ = io.ReadAll(rec.Body)
	assert.JSONEq(t, sampleSpec, string(body))
}

func TestHandlerWithoutSpec(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/swagger/", WrapHandler)

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.True(t, strings.Contains(string(body), "swagger-ui"))
}
