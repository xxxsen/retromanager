package handler

import (
	"retromanager/handler/game"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/xxxsen/common/naivesvr"
	"github.com/xxxsen/common/naivesvr/codec"

	"github.com/gin-gonic/gin"
)

func OnRegist(router *gin.Engine) {
	//game
	{
		gameRouter := router.Group("/game")
		gameRouter.POST("/list", naivesvr.WrapHandler(&gameinfo.ListGameRequest{}, codec.JsonCodec, game.ListGame))
		gameRouter.POST("/search", naivesvr.WrapHandler(&gameinfo.SearchGameRequest{}, codec.JsonCodec, game.SearchGame))
		gameRouter.POST("/create", naivesvr.WrapHandler(&gameinfo.CreateGameRequest{}, codec.JsonCodec, game.CreateGame))
		gameRouter.POST("/modify", naivesvr.WrapHandler(&gameinfo.ModifyGameRequest{}, codec.JsonCodec, game.ModifyGame))
		gameRouter.POST("/delete", naivesvr.WrapHandler(&gameinfo.DeleteGameRequest{}, codec.JsonCodec, game.DeleteGame))
	}
}
