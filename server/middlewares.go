package server

import (
	"context"
	"log"
	"retromanager/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	keyServerAttach = "key_server_attach"
)

func PanicRecoverMiddleware(svr *server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := recover(); err != nil {
			log.Printf("svr panic, path:%s, err:%v", ctx.Request.URL.Path, err)
		}
		ctx.Next()
	}
}

var emptyAttach = map[string]interface{}{}

func GetAttach(ctx context.Context) map[string]interface{} {
	iVal := ctx.Value(keyServerAttach)
	if iVal == nil {
		return emptyAttach
	}
	return iVal.(map[string]interface{})
}

func SupportAttachMiddleware(svr *server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(keyServerAttach, svr.c.attach)
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
