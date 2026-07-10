// Package flamingoswagger adapts swagger.Handler for flamingo.me.
//
// Flamingo routing uses named handlers and web.Action. Prefer flamingo's
// built-in web.WrapHTTPHandler with this package's WrapHandler:
//
//	registry.Route("/swagger/*path", "swagger.ui")
//	registry.HandleGet("swagger.ui", web.WrapHTTPHandler(flamingoswagger.WrapHandler(spec)))
//
// This module intentionally avoids depending on flamingo.me (heavy DI stack);
// it only returns a standard http.Handler for use with web.WrapHTTPHandler.
package flamingoswagger

import (
	"net/http"

	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns an http.Handler for flamingo's web.WrapHTTPHandler.
func WrapHandler(spec []byte, opts ...swagger.Option) http.Handler {
	return swagger.Handler(spec, opts...)
}
