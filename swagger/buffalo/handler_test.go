package buffaloswagger

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
)

const sampleSpec = `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0"},"paths":{}}`

func TestWrapHandler(t *testing.T) {
	app := buffalo.New(buffalo.Options{Env: "test"})
	app.ANY("/swagger/{doc:.*}", WrapHandler([]byte(sampleSpec)))

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("index status = %d body=%s loc=%s", rec.Code, rec.Body.String(), rec.Header().Get("Location"))
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "SwaggerUIBundle") {
		t.Fatal("expected Swagger UI HTML")
	}

	req = httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("doc status = %d body=%s", rec.Code, rec.Body.String())
	}
	body, _ = io.ReadAll(rec.Body)
	if string(body) != sampleSpec {
		t.Fatalf("doc body = %q", body)
	}
}
