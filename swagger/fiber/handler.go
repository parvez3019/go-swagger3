// Package fiberswagger adapts swagger.Handler for github.com/gofiber/fiber.
package fiberswagger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns a fiber.Handler that serves Swagger UI.
//
//	app.All("/swagger/*", fiberswagger.WrapHandler(spec))
func WrapHandler(spec []byte, opts ...swagger.Option) fiber.Handler {
	return adaptor.HTTPHandler(swagger.Handler(spec, opts...))
}
