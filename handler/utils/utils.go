package utils

import (
	"retromanager/constants"
	"retromanager/handler/config"

	"github.com/gin-gonic/gin"
)

func MustGetConfig(ctx *gin.Context) *config.Config {
	c, exist := ctx.Get(constants.KeyConfigAttach)
	if !exist {
		panic("config not exist")
	}
	return c.(*config.Config)
}
