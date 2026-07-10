// Package ginswagger adapts swagger.Handler for github.com/gin-gonic/gin.
package ginswagger

import (
	"github.com/gin-gonic/gin"
	"github.com/parvez3019/go-swagger3/swagger"
)

// WrapHandler returns a gin.HandlerFunc that serves Swagger UI.
//
//	r.Any("/swagger/*any", ginswagger.WrapHandler(spec))
func WrapHandler(spec []byte, opts ...swagger.Option) gin.HandlerFunc {
	h := swagger.Handler(spec, opts...)
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
