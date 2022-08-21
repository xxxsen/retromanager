package file

import (
	"net/http"
	"retromanager/codec"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/handler/utils"
	"retromanager/model"
	"retromanager/s3"

	"github.com/gin-gonic/gin"
)

type GameDownloadRequest struct {
	GameId uint64 `schema:"game_id"`
}

func RomDownload(ctx *gin.Context, request interface{}) (statuscode int, retErr errs.IError, response interface{}) {
	defer func() {
		if errs.IsErrOK(retErr) {
			return
		}
		codec.JsonCodec.Encode(ctx, statuscode, retErr, response)
	}()

	req := request.(*GameDownloadRequest)
	daoRsp, exist, err := dao.GameInfoDao.GetGame(ctx, &model.GetGameRequest{
		GameId: req.GameId,
	})
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "read gameinfo fail", err), nil
	}
	if !exist {
		return http.StatusOK, errs.New(constants.ErrNotFound, "not found").WithDebugMsg("gameid:%s", req.GameId), nil
	}
	c := utils.MustGetConfig(ctx)
	reader, err := s3.Client.Download(ctx, c.BucketInfo.RomBucket, daoRsp.Item.DownKey)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrS3, "s3 download fail", err), nil
	}
	defer reader.Close()
	fileToDownload(ctx, reader, daoRsp.Item.DownKey, daoRsp.Item.FileName, daoRsp.Item.FileSize)
	return http.StatusOK, nil, nil
}
