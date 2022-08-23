package bigfile

import (
	"net/http"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/utils"

	"github.com/gin-gonic/gin"
)

type PartUploadRequest struct {
	PartId    uint64 `form:"part_id" binding:"required"`
	MD5       string `form:"md5" binding:"required"`
	UploadCtx string `form:"upload_ctx" binding:"required"`
}

func Part(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*PartUploadRequest)
	var (
		partid     = req.PartId
		md5        = req.MD5
		suploadctx = req.UploadCtx
	)
	uploadctx, err := utils.DecodeUploadID(suploadctx)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "parse uploadid fail", err), nil
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "get file fail", err), nil
	}
	defer file.Close()
	maxpartid := utils.CalcFileBlockCount(uploadctx.GetFileSize(), constants.BlockSize)
	if partid > uint64(maxpartid) || partid == 0 {
		return http.StatusOK, errs.New(constants.ErrParam, "partid invalid").
			WithDebugMsg("partid:%d", partid).WithDebugMsg("maxid:%d", maxpartid), nil
	}
	if header.Size > constants.BlockSize {
		return http.StatusOK, errs.New(constants.ErrParam, "block size out of limit"), nil
	}
	if header.Size < constants.BlockSize && partid != uint64(maxpartid) {
		return http.StatusOK, errs.New(constants.ErrParam, "part size invalid, should eq to block size"), nil
	}
	if partid == uint64(maxpartid) && (maxpartid-1)*constants.BlockSize+int(header.Size) != int(uploadctx.GetFileSize()) {
		return http.StatusOK, errs.New(constants.ErrParam, "full block size != file size").WithDebugMsg("last block size:%d", header.Size), nil
	}
	err = s3.Client.UploadPart(ctx, uploadctx.GetDownKey(), uploadctx.GetUploadId(), int(partid), file, md5)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrS3, "upload part fail", err), nil
	}
	return http.StatusOK, nil, &gameinfo.FileUploadPartResponse{}
}
