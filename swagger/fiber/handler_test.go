package fiberswagger

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

const sampleSpec = `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0"},"paths":{}}`

func TestWrapHandler(t *testing.T) {
	app := fiber.New()
	app.All("/swagger/*", WrapHandler([]byte(sampleSpec)))

	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("index status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "SwaggerUIBundle") {
		t.Fatal("expected Swagger UI HTML")
	}

	req = httptest.NewRequest("GET", "/swagger/doc.json", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("doc status = %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)
	if string(body) != sampleSpec {
		t.Fatalf("doc body = %q", body)
	}
}
