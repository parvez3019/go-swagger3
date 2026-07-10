// Package chiswagger adapts swagger.Handler for github.com/go-chi/chi.
//
// Chi uses the standard net/http Handler interface, so this package re-exports
// swagger.Handler for a consistent adapter API.
package chiswagger

import (
	"net/http"

	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns an http.Handler that serves Swagger UI.
//
//	r.Handle("/swagger/*", chiswagger.WrapHandler(spec))
func WrapHandler(spec []byte, opts ...swagger.Option) http.Handler {
	return swagger.Handler(spec, opts...)
}
