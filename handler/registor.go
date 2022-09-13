package handler

import (
	"retromanager/handler/file"
	"retromanager/handler/file/bigfile"
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
	//upload
	{
		uploadRouter := router.Group("/upload")
		uploadRouter.POST("/image", naivesvr.WrapHandler(&file.BasicFileUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), file.ImageUpload))
		uploadRouter.POST("/video", naivesvr.WrapHandler(&file.BasicFileUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), file.VideoUpload))
		uploadRouter.POST("/file", naivesvr.WrapHandler(&file.BasicFileUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), file.FileUpload))
		bigFileRouter := uploadRouter.Group("/bigfile")
		bigFileRouter.POST("/begin", naivesvr.WrapHandler(&gameinfo.FileUploadBeginRequest{}, codec.JsonCodec, bigfile.Begin))
		bigFileRouter.POST("/part", naivesvr.WrapHandler(&bigfile.PartUploadRequest{}, codec.CustomCodec(codec.JsonCodec, codec.MultipartCodec), bigfile.Part))
		bigFileRouter.POST("/end", naivesvr.WrapHandler(&gameinfo.FileUploadEndRequest{}, codec.JsonCodec, bigfile.End))

	}
	//download
	{
		router.GET("/file", naivesvr.WrapHandler(&file.BasicFileDownloadRequest{}, codec.CustomCodec(codec.NopCodec, codec.QueryCodec), file.FileDownload)) //input: down_key
	}
	//meta
	{
		router.POST("/file/meta", naivesvr.WrapHandler(&gameinfo.GetFileMetaRequest{}, codec.JsonCodec, file.Meta))
	}
}
