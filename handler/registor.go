package handler

import (
	"retromanager/handler/file"
	"retromanager/handler/file/bigfile"
	"retromanager/handler/game"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/xxxsen/naivesvr/codec"

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
		uploadRouter.POST("/image", WrapHandler(&file.BasicFileUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), file.ImageUpload))
		uploadRouter.POST("/video", WrapHandler(&file.BasicFileUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), file.VideoUpload))
		uploadRouter.POST("/file", WrapHandler(&file.BasicFileUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), file.FileUpload))
		bigFileRouter := uploadRouter.Group("/bigfile")
		bigFileRouter.POST("/begin", WrapHandler(&gameinfo.FileUploadBeginRequest{}, codec.JsonCodec, bigfile.Begin))
		bigFileRouter.POST("/part", WrapHandler(&bigfile.PartUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), bigfile.Part))
		bigFileRouter.POST("/end", WrapHandler(&gameinfo.FileUploadEndRequest{}, codec.JsonCodec, bigfile.End))

	}
	//download
	{
		router.GET("/file", WrapHandler(&file.BasicFileDownloadRequest{}, codec.CustomCodec(codec.NopCodec, codec.QueryCodec), file.FileDownload)) //input: down_key
	}
	//meta
	{
		router.POST("/file/meta", WrapHandler(&gameinfo.GetFileMetaRequest{}, codec.JsonCodec, file.Meta))
	}
}
