package bigfile

import (
	"net/http"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/utils"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

func End(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.FileUploadEndRequest)
	if len(req.GetHash()) == 0 || len(req.GetFileName()) == 0 ||
		len(req.GetUploadCtx()) == 0 {
		return http.StatusOK, errs.New(constants.ErrParam, "invalid hash/filename/partctx/uploadctx"), nil
	}
	uploadctx, err := utils.DecodeUploadID(req.GetUploadCtx())
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "decode upload ctx fail", err), nil
	}
	maxpartsz := utils.CalcFileBlockCount(uploadctx.GetFileSize(), constants.BlockSize)
	err = s3.Client.EndUpload(ctx, *uploadctx.DownKey, uploadctx.GetUploadId(), maxpartsz)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrS3, "complete upload fail", err), nil
	}
	_, err = dao.FileInfoDao.CreateFile(ctx, &model.CreateFileRequest{
		Item: &model.FileItem{
			FileName:   req.GetFileName(),
			Hash:       req.GetHash(),
			FileSize:   uploadctx.GetFileSize(),
			CreateTime: uint64(time.Now().UnixMilli()),
			DownKey:    uploadctx.GetDownKey(),
		},
	})
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "write file record fail", err), nil
	}
	return http.StatusOK, nil, &gameinfo.FileUploadEndResponse{
		DownKey: proto.String(uploadctx.GetDownKey()),
	}
}