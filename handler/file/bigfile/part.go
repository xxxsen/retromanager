package bigfile

import (
	"io/ioutil"
	"net/http"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

func Part(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	if err := ctx.Request.ParseMultipartForm(constants.BlockSize); err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "parse form fail", err), nil
	}
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
	if header.Size < constants.BlockSize && partid+1 != uint64(maxpartid) {
		return http.StatusOK, errs.New(constants.ErrParam, "part size invalid, should eq to block size"), nil
	}

	raw, err := ioutil.ReadAll(file)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrIO, "read file fail", err), nil
	}
	part, err := s3.Client.UploadPart(ctx, uploadctx.GetDownKey(), uploadctx.GetUploadId(), int(partid), raw)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrS3, "upload part fail", err), nil
	}
	partidctx := &gameinfo.PartIdCtx{
		Idx:  proto.Int32(int32(part.PartNumber)),
		Etag: proto.String(part.ETag),
	}
	partctx, err := utils.EncodePartID(partidctx)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrMarshal, "build partid fail", err), nil
	}
	return http.StatusOK, nil, &gameinfo.FileUploadPartResponse{
		PartCtx: proto.String(partctx),
	}
}
