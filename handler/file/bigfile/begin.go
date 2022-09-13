package bigfile

import (
	"fmt"
	"net/http"
	"retromanager/constants"
	"retromanager/idgen"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/utils"

	"github.com/xxxsen/common/s3"

	"github.com/xxxsen/common/errs"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

func Begin(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.FileUploadBeginRequest)
	if req.GetFileSize() == 0 {
		return http.StatusOK, errs.New(errs.ErrParam, "zero size file"), nil
	}
	if req.GetFileSize() > constants.MaxFileSize {
		return http.StatusOK, errs.New(errs.ErrParam, "file size out of limit"), nil
	}
	downkey := fmt.Sprintf("%d_%s", model.FileTypeAny, utils.EncodeFileId(idgen.NextId()))
	key, err := s3.Client.BeginUpload(ctx, downkey)
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrS3, "begin upload fail", err), nil
	}
	uploadidctx := &gameinfo.UploadIdCtx{
		FileSize: req.FileSize,
		UploadId: proto.String(key),
		DownKey:  proto.String(downkey),
	}
	uploadctx, err := utils.EncodeUploadID(uploadidctx)
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrServiceInternal, "build upload id fail", err), nil
	}
	return http.StatusOK, nil, &gameinfo.FileUploadBeginResponse{
		UploadCtx: proto.String(uploadctx),
		BlockSize: proto.Uint32(uint32(constants.BlockSize)),
	}
}
