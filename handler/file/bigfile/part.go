package bigfile

import (
	"net/http"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Part(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	suploadid, _ := ctx.GetPostForm("upload_id")
	spartid, _ := ctx.GetPostForm("part_id")
	partid, err := strconv.ParseUint(spartid, 10, 64)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "parse partid fail", err), nil
	}
	uploadctx, err := utils.DecodeUploadID(suploadid)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "parse uploadid fail", err), nil
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "get file fail", err), nil
	}
	defer file.Close()
	maxpartid := utils.CalcFileBlockCount(uint64(header.Size), constants.BlockSize)
	if partid >= uint64(maxpartid) {
		return http.StatusOK, errs.New(constants.ErrParam, "partid out of limit").
			WithDebugMsg("partid:%d", partid).WithDebugMsg("maxid:%d", maxpartid), nil
	}
	if header.Size > constants.BlockSize {
		return http.StatusOK, errs.New(constants.ErrParam, "block size out of limit"), nil
	}
	if header.Size < constants.BlockSize && partid+1 != uint64(maxpartid) {
		return http.StatusOK, errs.New(constants.ErrParam, "part size invalid, should eq to block size"), nil
	}
	err = s3.Client.UploadPart(ctx, uploadctx.GetDownKey(), uploadctx.GetUploadId(), int(partid), file)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrS3, "upload part fail", err), nil
	}
	return http.StatusOK, nil, &gameinfo.FileUploadPartResponse{}
}
