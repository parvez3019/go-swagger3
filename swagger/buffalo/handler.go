// Package buffaloswagger adapts swagger.Handler for github.com/gobuffalo/buffalo.
package buffaloswagger

import (
	"github.com/gobuffalo/buffalo"
	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns a buffalo.Handler that serves Swagger UI.
//
//	app.ANY("/swagger/{doc:.*}", buffaloswagger.WrapHandler(spec))
//
// Buffalo normalizes paths with a trailing slash; the core swagger handler
// tolerates that for file-like paths such as index.html and doc.json.
func WrapHandler(spec []byte, opts ...swagger.Option) buffalo.Handler {
	return buffalo.WrapHandler(swagger.Handler(spec, opts...))
}
