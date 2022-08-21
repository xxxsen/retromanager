package file

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/handler/utils"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/server"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

type bucketGetterFunc func(ctx *gin.Context) string
type postUploadCallback func(ctx *gin.Context, hash string, meta *multipart.FileHeader) error
type preuploadCheckFunc func(ctx *gin.Context, hash string, meta *multipart.FileHeader) (bool, error)

func postUploader(maxsize uint64, bucketGetter bucketGetterFunc, precheck preuploadCheckFunc, cb postUploadCallback) server.ProcessFunc {
	return func(ctx *gin.Context, req interface{}) (int, errs.IError, interface{}) {
		ctx.Request.ParseMultipartForm(int64(maxsize))
		file, header, err := ctx.Request.FormFile("file")
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrParam, "get file fail", err), nil
		}
		raw, err := utils.ReadAndCloseFile(file)
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrIO, "read file", err), nil
		}
		filename := utils.CalcMd5(raw)
		bucket := bucketGetter(ctx)

		exist, err := precheck(ctx, filename, header)
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrServiceInternal, "upload precheck fail", err), nil
		}
		if exist {
			return http.StatusOK, nil, createPostUploadRsp(filename)
		}
		if err := s3.Client.Upload(ctx, bucket, filename, bytes.NewReader(raw), int64(len(raw))); err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrS3, "upload fail", err), nil
		}
		if err := cb(ctx, filename, header); err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrServiceInternal, "internal service err", err).WithDebugMsg("bucket:%s", bucket), nil
		}
		return http.StatusOK, nil, createPostUploadRsp(filename)
	}
}

func createPostUploadRsp(fileid string) *gameinfo.ImageUploadResponse {
	return &gameinfo.ImageUploadResponse{
		FileId: proto.String(fileid),
	}
}
