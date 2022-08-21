package server

import (
	"log"

	"github.com/gin-gonic/gin"
)

func PanicRecoverMiddleware(svr *server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := recover(); err != nil {
			log.Printf("svr panic, path:%s, err:%v", ctx.Request.URL.Path, err)
		}
		ctx.Next()
	}
}

func SupportAttachMiddleware(svr *server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for k, v := range svr.c.attach {
			ctx.Set(k, v)
		}
	}
}
