package file

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"retromanager/cache"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/handler/utils"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/server"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

type bucketGetterFunc func(ctx *gin.Context) string

func checkExistInDB(ctx context.Context, typ uint32, hash string) (bool, error) {
	if _, exist, _ := cache.Default().Get(ctx, typeKey(typ, hash)); exist {
		return true, nil
	}
	_, exist, err := dao.MediaInfoDao.GetMedia(ctx, &model.GetMediaRequest{
		Hash:     hash,
		FileType: typ,
	})
	if err != nil {
		return false, errs.Wrap(constants.ErrDatabase, "check image db", err)
	}
	if exist {
		_ = cache.Default().Set(ctx, typeKey(typ, hash), true, 0)
	}
	return exist, nil
}

func postUploader(maxsize uint64, bucketGetter bucketGetterFunc, typ uint32) server.ProcessFunc {
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

		exist, err := checkExistInDB(ctx, typ, filename)
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrServiceInternal, "upload precheck fail", err), nil
		}
		if exist {
			return http.StatusOK, nil, createPostUploadRsp(filename)
		}
		if err := s3.Client.Upload(ctx, bucket, filename, bytes.NewReader(raw), int64(len(raw))); err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrS3, "upload fail", err), nil
		}
		if err := writeTypeRecordToDB(ctx, typ, filename, header); err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrServiceInternal, "internal service err", err).WithDebugMsg("bucket:%s", bucket), nil
		}
		return http.StatusOK, nil, createPostUploadRsp(filename)
	}
}

func writeTypeRecordToDB(ctx context.Context, typ uint32, hash string, meta *multipart.FileHeader) error {
	if _, err := dao.MediaInfoDao.CreateMedia(ctx, &model.CreateMediaRequest{
		Item: &model.MediaItem{
			FileName:   meta.Filename,
			Hash:       hash,
			FileSize:   uint64(meta.Size),
			CreateTime: uint64(time.Now().UnixMilli()),
			FileType:   typ,
		},
	}); err != nil {
		return errs.Wrap(constants.ErrDatabase, "create record fail", err).
			WithDebugMsg("typ:%d", typ).WithDebugMsg("hash:%s", hash)
	}
	_ = cache.Default().Set(ctx, typeKey(typ, hash), true, 0)
	return nil
}

func createPostUploadRsp(fileid string) *gameinfo.ImageUploadResponse {
	return &gameinfo.ImageUploadResponse{
		FileId: proto.String(fileid),
	}
}
