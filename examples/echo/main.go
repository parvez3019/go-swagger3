// Example: serve go-swagger3 OpenAPI UI with echo.
package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/parvez3019/go-swagger3/swagger"
	echoswagger "github.com/parvez3019/go-swagger3/swagger/echo"
)

var minimalSpec = []byte(`{
  "openapi": "3.0.0",
  "info": {"title": "Echo Example API", "version": "1.0.0"},
  "paths": {}
}`)

func main() {
	e := echo.New()
	e.Any("/swagger/*", echoswagger.WrapHandler(minimalSpec, swagger.Title("Echo Example API")))
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	e.Logger.Fatal(e.Start(":8080"))
}
