// Example: serve go-swagger3 OpenAPI UI with gin.
//
// From this repo (adapter is a separate module):
//
//	cd examples/gin && go mod tidy && go run .
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/parvez3019/go-swagger3/swagger"
	ginswagger "github.com/parvez3019/go-swagger3/swagger/gin"
)

var minimalSpec = []byte(`{
  "openapi": "3.0.0",
  "info": {"title": "Gin Example API", "version": "1.0.0"},
  "paths": {}
}`)

func main() {
	r := gin.Default()
	r.Any("/swagger/*any", ginswagger.WrapHandler(minimalSpec, swagger.Title("Gin Example API")))
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	_ = r.Run(":8080")
}
