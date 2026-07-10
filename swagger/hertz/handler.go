// Package hertzswagger adapts swagger.Handler for github.com/cloudwego/hertz.
package hertzswagger

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns a Hertz app.HandlerFunc that serves Swagger UI.
//
//	h.GET("/swagger/*filepath", hertzswagger.WrapHandler(spec))
func WrapHandler(spec []byte, opts ...swagger.Option) app.HandlerFunc {
	return adaptor.HertzHandler(swagger.Handler(spec, opts...))
}
