package utils

import (
	"context"
	"retromanager/server/constants"

	"github.com/gin-gonic/gin"
)

func SetTraceID(ctx *gin.Context, traceid string) {
	ctx.Set(constants.KeyTraceID, traceid)
}

func GetTraceId(ctx context.Context) (string, bool) {
	v := ctx.Value(constants.KeyTraceID)
	if v == nil {
		return "", false
	}
	return v.(string), true
}

func MustGetTraceId(ctx context.Context) string {
	if v, ok := GetTraceId(ctx); ok {
		return v
	}
	panic("not found trace id")
}
