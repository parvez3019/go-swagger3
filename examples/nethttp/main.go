// Example: serve go-swagger3 OpenAPI UI with net/http.
package main

import (
	"log"
	"net/http"

	"github.com/parvez3019/go-swagger3/swagger"
)

var minimalSpec = []byte(`{
  "openapi": "3.0.0",
  "info": {"title": "Example API", "version": "1.0.0"},
  "paths": {}
}`)

func main() {
	http.Handle("/swagger/", swagger.Handler(minimalSpec, swagger.Title("Example API")))
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
