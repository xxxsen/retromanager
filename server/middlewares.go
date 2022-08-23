package server

import (
	"log"
	"retromanager/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func EnableServerTrace(svr *server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestid := ctx.GetHeader("x-request-id")
		if len(requestid) == 0 {
			requestid = uuid.NewString()
		}
		ctx.Writer.Header().Set("x-request-id", requestid)
		utils.SetTraceID(ctx, requestid)
	}
}
