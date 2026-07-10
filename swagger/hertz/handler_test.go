package hertzswagger

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
)

const sampleSpec = `{"openapi":"3.0.0","info":{"title":"Demo","version":"1.0"},"paths":{}}`

func TestWrapHandler(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	h := server.Default(server.WithHostPorts(addr), server.WithListener(ln))
	h.GET("/swagger/*filepath", WrapHandler([]byte(sampleSpec)))
	go h.Spin()
	defer func() { _ = h.Close() }()

	deadline := time.Now().Add(3 * time.Second)
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
