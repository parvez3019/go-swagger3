package atreugoswagger

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/savsgio/atreugo/v11"
)

const sampleSpec = `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0"},"paths":{}}`

func TestWrapHandler(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	server := atreugo.New(atreugo.Config{Addr: addr})
	server.GET("/swagger/{filepath:*}", WrapHandler([]byte(sampleSpec)))

	go func() {
		_ = server.ListenAndServe()
	}()
	defer func() { _ = server.Shutdown() }()

	deadline := time.Now().Add(2 * time.Second)
	var resp *http.Response
	for {
		resp, err = http.Get("http://" + addr + "/swagger/index.html")
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not ready: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("index status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "SwaggerUIBundle") {
		t.Fatal("expected Swagger UI HTML")
	}

	resp, err = http.Get("http://" + addr + "/swagger/doc.json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("doc status = %d", resp.StatusCode)
	}
	body, _ = io.ReadAll(resp.Body)
	if string(body) != sampleSpec {
		t.Fatalf("doc body = %q", body)
	}
}
