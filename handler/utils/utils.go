package utils

import (
	"retromanager/constants"
	"retromanager/handler/config"

	"github.com/gin-gonic/gin"
	"github.com/xxxsen/naivesvr"
)

func MustGetConfig(ctx *gin.Context) *config.Config {
	attach := naivesvr.GetAttach(ctx)
	c, exist := attach[constants.KeyConfigAttach]
	if !exist {
		panic("config not exist")
	}
	return c.(*config.Config)
}
