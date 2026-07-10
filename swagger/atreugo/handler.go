// Package atreugoswagger adapts swagger.Handler for github.com/savsgio/atreugo.
package atreugoswagger

import (
	"github.com/parvez3019/go-swagger3/swagger"
	"github.com/savsgio/atreugo/v11"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// WrapHandler returns an atreugo.View that serves Swagger UI.
//
//	server.GET("/swagger/{filepath:*}", atreugoswagger.WrapHandler(spec))
//
// Alternatively register via NetHTTPPath with swagger.Handler directly.
func WrapHandler(spec []byte, opts ...swagger.Option) atreugo.View {
	h := fasthttpadaptor.NewFastHTTPHandler(swagger.Handler(spec, opts...))
	return func(ctx *atreugo.RequestCtx) error {
		h(ctx.RequestCtx)
		return nil
	}
}
