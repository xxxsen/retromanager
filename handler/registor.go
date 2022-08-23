package handler

import (
	"retromanager/codec"
	"retromanager/handler/file"
	"retromanager/handler/file/bigfile"
	"retromanager/handler/game"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/gin-gonic/gin"
)

func OnRegist(router *gin.Engine) {
	//game
	{
		gameRouter := router.Group("/game")
		gameRouter.POST("/list", WrapHandler(&gameinfo.ListGameRequest{}, codec.JsonCodec, game.ListGame))
		gameRouter.POST("/search", WrapHandler(&gameinfo.SearchGameRequest{}, codec.JsonCodec, game.SearchGame))
		gameRouter.POST("/create", WrapHandler(&gameinfo.CreateGameRequest{}, codec.JsonCodec, game.CreateGame))
		gameRouter.POST("/modify", WrapHandler(&gameinfo.ModifyGameRequest{}, codec.JsonCodec, game.ModifyGame))
		gameRouter.POST("/delete", WrapHandler(&gameinfo.DeleteGameRequest{}, codec.JsonCodec, game.DeleteGame))
	}
	//upload
	{
		uploadRouter := router.Group("/upload")
		uploadRouter.POST("/image", WrapHandler(nil, codec.CustomCodec(codec.JsonCodec, codec.NopCodec), file.ImageUpload))
		uploadRouter.POST("/video", WrapHandler(nil, codec.CustomCodec(codec.JsonCodec, codec.NopCodec), file.VideoUpload))
		uploadRouter.POST("/file", WrapHandler(nil, codec.CustomCodec(codec.JsonCodec, codec.NopCodec), file.FileUpload))
		bigFileRouter := uploadRouter.Group("/bigfile")
		bigFileRouter.POST("/begin", WrapHandler(&gameinfo.FileUploadBeginRequest{}, codec.JsonCodec, bigfile.Begin))
		bigFileRouter.POST("/part", WrapHandler(nil, codec.CustomCodec(codec.JsonCodec, codec.NopCodec), bigfile.Part))
		bigFileRouter.POST("/end", WrapHandler(&gameinfo.FileUploadEndRequest{}, codec.JsonCodec, bigfile.End))

	}
	//download
	{
		router.GET("/file", WrapHandler(nil, codec.NopCodec, file.FileDownload)) //input: down_key
	}
}
