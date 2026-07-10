// Package muxswagger adapts swagger.Handler for github.com/gorilla/mux.
package muxswagger

import (
	"net/http"

	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns an http.Handler that serves Swagger UI.
//
//	r.PathPrefix("/swagger/").Handler(muxswagger.WrapHandler(spec))
func WrapHandler(spec []byte, opts ...swagger.Option) http.Handler {
	return swagger.Handler(spec, opts...)
}
