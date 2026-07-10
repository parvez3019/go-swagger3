package flamingoswagger

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleSpec = `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0"},"paths":{}}`

func TestWrapHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/swagger/", WrapHandler([]byte(sampleSpec)))

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("index status = %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "SwaggerUIBundle") {
		t.Fatal("expected Swagger UI HTML")
	}

	req = httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("doc status = %d", rec.Code)
	}
	body, _ = io.ReadAll(rec.Body)
	if string(body) != sampleSpec {
		t.Fatalf("doc body = %q", body)
	}
}
