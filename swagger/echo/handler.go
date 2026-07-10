// Package echoswagger adapts swagger.Handler for github.com/labstack/echo.
package echoswagger

import (
	"github.com/labstack/echo/v4"
	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns an echo.HandlerFunc that serves Swagger UI.
//
//	e.Any("/swagger/*", echoswagger.WrapHandler(spec))
func WrapHandler(spec []byte, opts ...swagger.Option) echo.HandlerFunc {
	return echo.WrapHandler(swagger.Handler(spec, opts...))
}
