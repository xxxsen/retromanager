package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"retromanager/cache"
	"retromanager/codec"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/handler/utils"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/server"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xxxsen/log"
	"google.golang.org/protobuf/proto"
)

type FileDownloadRequest struct {
	FileId string `schema:"file_id"`
}

type bucketGetterFunc func(ctx *gin.Context) string
type hashGetterFunc func(ctx *gin.Context, request interface{}) (string, error)

func checkExistInDB(ctx context.Context, typ uint32, hash string) (bool, error) {
	_, exist, err := getMediaTypeInfoFromDB(ctx, typ, hash)
	if err != nil {
		return false, err
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
	return nil
}

func createPostUploadRsp(fileid string) *gameinfo.ImageUploadResponse {
	return &gameinfo.ImageUploadResponse{
		FileId: proto.String(fileid),
	}
}

func mediaFileDownload(bucketGetter bucketGetterFunc, typ uint32, hashGetter hashGetterFunc) server.ProcessFunc {
	return func(ctx *gin.Context, req interface{}) (statuscode int, retErr errs.IError, response interface{}) {
		defer func() {
			if errs.IsErrOK(retErr) {
				return
			}
			codec.JsonCodec.Encode(ctx, statuscode, retErr, response)
		}()

		hash, err := hashGetter(ctx, req)
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrParam, "hash not found", err), nil
		}
		meta, exist, err := getMediaTypeInfoFromDB(ctx, typ, hash)
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrDatabase, "get meta fail", err), nil
		}
		if !exist {
			return http.StatusOK, errs.New(constants.ErrNotFound, "not found"), nil
		}
		bucket := bucketGetter(ctx)
		reader, err := s3.Client.Download(ctx, bucket, hash)
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrS3, "read stream fail", err), nil
		}
		defer reader.Close()
		fileToDownload(ctx, reader, hash, meta.FileName, meta.FileSize)
		return http.StatusOK, nil, nil
	}
}

func fileToDownload(ctx *gin.Context, reader io.Reader, fileid string, name string, size uint64) {
	writer := ctx.Writer
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", strconv.Quote(name)))
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))
	sz, err := io.Copy(ctx.Writer, reader)
	if err != nil || sz != int64(size) {
		log.Errorf("write data to remote fail, path:%s, fileid:%s, sz:%d, err:%v", ctx.Request.URL.Path, fileid, sz, err)
	}
}

func getMediaTypeInfoFromDB(ctx context.Context, typ uint32, hash string) (*model.MediaItem, bool, error) {
	if val, exist, _ := cache.Default().Get(ctx, typeKey(typ, hash)); exist {
		return val.(*model.MediaItem), true, nil
	}
	val, exist, err := dao.MediaInfoDao.GetMedia(ctx, &model.GetMediaRequest{
		Hash:     hash,
		FileType: typ,
	})
	if err != nil {
		return nil, false, errs.Wrap(constants.ErrDatabase, "check db", err)
	}
	if !exist {
		return nil, false, nil
	}
	_ = cache.Default().Set(ctx, typeKey(typ, hash), val.Item, 0)
	return val.Item, true, nil
}
