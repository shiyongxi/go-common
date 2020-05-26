package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type (
	Cors struct {
		Headers       []string
		Methods       []string
		Credentials   string
		ExposeHeaders string
	}
)

var (
	corsHeader = []string{
		"appid",
		"appversion",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"X-Auth-Token",
		"Authorization",
		"Accept",
		"Origin",
		"Cache-Control",
		"X-Requested-With"}

	corsMethods = []string{
		"POST",
		"OPTIONS",
		"GET",
		"PUT",
		"DELETE",
	}
)

func NewCors() *Cors {
	return &Cors{
		Headers:       corsHeader,
		Methods:       corsMethods,
		Credentials:   "true",
		ExposeHeaders: "*",
	}
}

func (c *Cors) Defualt(ctx *gin.Context) {
	ctx.Writer.Header().Set("Access-Control-Expose-Headers", c.ExposeHeaders)
	ctx.Writer.Header().Set("Access-Control-Allow-Credentials", c.Credentials)
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", ctx.GetHeader("Origin"))
	ctx.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(c.Headers, ","))
	ctx.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(c.Methods, ","))
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(200)
		ctx.Next()
		return
	}

	ctx.Next()
}
