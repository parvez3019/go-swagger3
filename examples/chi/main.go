// Example: serve go-swagger3 OpenAPI UI with chi.
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/parvez3019/go-swagger3/swagger"
	chiswagger "github.com/parvez3019/go-swagger3/swagger/chi"
)

var minimalSpec = []byte(`{
  "openapi": "3.0.0",
  "info": {"title": "Chi Example API", "version": "1.0.0"},
  "paths": {}
}`)

func main() {
	r := chi.NewRouter()
	r.Handle("/swagger/*", chiswagger.WrapHandler(minimalSpec, swagger.Title("Chi Example API")))
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8080", r))
}
