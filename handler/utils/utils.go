package utils

import (
	"retromanager/constants"
	"retromanager/handler/config"
	"retromanager/server"

	"github.com/gin-gonic/gin"
)

func MustGetConfig(ctx *gin.Context) *config.Config {
	attach := server.GetAttach(ctx)
	c, exist := attach[constants.KeyConfigAttach]
	if !exist {
		panic("config not exist")
	}
	return c.(*config.Config)
}
