package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"retromanager/cache"
	"retromanager/codec"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/server"
	rutils "retromanager/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xxxsen/log"
	"google.golang.org/protobuf/proto"
)

type FileUploadMeta struct {
	Data      []byte
	FileName  string
	Dir       string
	StoreName string
}

type ISmallFileUploader interface {
	BeforeUpload(ctx *gin.Context, request interface{}) (*FileUploadMeta, bool, error)
	OnUpload(ctx *gin.Context, meta *FileUploadMeta) error
	AfterUpload(ctx *gin.Context, realUpload bool, meta *FileUploadMeta) (interface{}, error)
}

type S3SmallFileUploader struct {
}

func (f *S3SmallFileUploader) BeforeUpload(ctx *gin.Context, request interface{}) (*FileUploadMeta, bool, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, false, errs.Wrap(constants.ErrParam, "get form file fail", err)
	}
	defer file.Close()
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, false, errs.Wrap(constants.ErrIO, "read file data fail", err)
	}
	return &FileUploadMeta{
		Data:      raw,
		FileName:  header.Filename,
		StoreName: rutils.CalcMd5(raw),
	}, true, nil

	return nil, false, nil
}

func (f *S3SmallFileUploader) OnUpload(ctx *gin.Context, meta *FileUploadMeta) error {
	if len(meta.Data) == 0 {
		return errs.New(constants.ErrParam, "no file data")
	}
	if err := s3.Client.Upload(ctx, meta.StoreName, bytes.NewReader(meta.Data), int64(len(meta.Data))); err != nil {
		return errs.Wrap(constants.ErrS3, "upload to s3 fail", err)
	}
	return nil
}

func (f *S3SmallFileUploader) AfterUpload(ctx *gin.Context, realUpload bool, meta *FileUploadMeta) (interface{}, error) {
	return nil, nil
}

type FileDownloadMeta struct {
	FileId      string
	FileName    string
	FileSize    int64
	ContentType string
}

type IFileDownloader interface {
	BeforeDownload(ctx *gin.Context, request interface{}) (*FileDownloadMeta, error)
	OnDownload(ctx *gin.Context, meta *FileDownloadMeta) error
	AfterDownload(ctx *gin.Context, meta *FileDownloadMeta, err error)
}

type S3FileDownloader struct {
}

func (f *S3FileDownloader) BeforeDownload(ctx *gin.Context, request interface{}) (*FileDownloadMeta, error) {
	return nil, fmt.Errorf("need impl")
}

func (f *S3FileDownloader) OnDownload(ctx *gin.Context, meta *FileDownloadMeta) error {
	reader, err := s3.Client.Download(ctx, meta.FileId)
	if err != nil {
		return errs.Wrap(constants.ErrS3, "create download stream fail", err)
	}
	defer reader.Close()
	contentType := meta.ContentType
	if len(contentType) == 0 {
		contentType = mime.TypeByExtension(filepath.Ext(meta.FileName))
	}
	writer := ctx.Writer
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", strconv.Quote(meta.FileName)))
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", meta.FileSize))
	writer.Header().Set("Content-Type", contentType)
	sz, err := io.Copy(ctx.Writer, reader)
	if err != nil {
		return errs.Wrap(constants.ErrIO, "copy stream fail", err)
	}
	if sz != int64(meta.FileSize) {
		return errs.New(constants.ErrIO, "io size not match, need %d, write:%d", meta.FileSize, sz)
	}
	return nil
}

func (f *S3FileDownloader) AfterDownload(ctx *gin.Context, meta *FileDownloadMeta, err error) {
	if err == nil {
		return
	}
	log.Errorf("file download fail, path:%s, err:%v", ctx.Request.URL.Path, err)
}

func CommonFilePostUpload(uploader ISmallFileUploader) server.ProcessFunc {
	return func(ctx *gin.Context, req interface{}) (int, errs.IError, interface{}) {
		caller := func() (interface{}, errs.IError) {
			meta, needUpload, err := uploader.BeforeUpload(ctx, req)
			if err != nil {
				return nil, errs.Wrap(constants.ErrServiceInternal, "before post upload fail", err)
			}
			if needUpload {
				if err := uploader.OnUpload(ctx, meta); err != nil {
					return nil, errs.Wrap(constants.ErrStorage, "on upload fail", err)
				}
			}
			rsp, err := uploader.AfterUpload(ctx, needUpload, meta)
			if err != nil {
				return nil, errs.Wrap(constants.ErrServiceInternal, "after upload fail", err)
			}
			return rsp, nil
		}
		rsp, err := caller()
		return http.StatusOK, err, rsp
	}
}

func CommonFileDownload(downloader IFileDownloader) server.ProcessFunc {
	return func(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
		caller := func() error {
			meta, err := downloader.BeforeDownload(ctx, request)
			if err == nil {
				err = downloader.OnDownload(ctx, meta)
			}
			downloader.AfterDownload(ctx, meta, err)
			if err != nil {
				return err
			}
			return nil
		}
		if err := caller(); err != nil {
			e := errs.FromError(err)
			log.Errorf("call file download fail, path:%s, err:%v", e)
			codec.JsonCodec.Encode(ctx, http.StatusOK, e, nil)
		}
		return http.StatusOK, nil, nil
	}
}

func cacheGetFileMeta(ctx context.Context, c *cache.Cache, key interface{},
	cb func() (interface{}, bool, error)) (interface{}, bool, error) {

	ival, exist, _ := c.Get(ctx, key)
	if exist {
		return ival, true, nil
	}
	val, exist, err := cb()
	if err != nil {
		return nil, false, err
	}
	if exist {
		c.Set(ctx, key, val, 10*time.Minute)
	}
	return val, exist, nil
}

type FileUploader struct {
	S3SmallFileUploader
	typ uint32
	c   *cache.Cache
}

func NewFileUploader(typ uint32, c *cache.Cache) *FileUploader {
	return &FileUploader{
		typ: typ,
		c:   c,
	}
}

func (uploader *FileUploader) BeforeUpload(ctx *gin.Context, request interface{}) (*FileUploadMeta, bool, error) {
	meta, _, err := uploader.S3SmallFileUploader.BeforeUpload(ctx, request)
	if err != nil {
		return nil, false, err
	}
	meta.StoreName = fmt.Sprintf("%d_%s", uploader.typ, meta.StoreName)
	_, exist, _ := uploader.c.Get(ctx, meta.StoreName)
	if exist {
		return meta, false, nil
	}
	return meta, true, nil
}

func (uploader *FileUploader) AfterUpload(ctx *gin.Context, realUpload bool, meta *FileUploadMeta) (interface{}, error) {
	if _, err := dao.MediaInfoDao.CreateMedia(ctx, &model.CreateMediaRequest{
		Item: &model.MediaItem{
			FileName:   meta.FileName,
			Hash:       meta.StoreName,
			FileSize:   uint64(len(meta.Data)),
			CreateTime: uint64(time.Now().UnixMilli()),
			FileType:   uploader.typ,
		},
	}); err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "insert image to db fail", err)
	}
	return &gameinfo.ImageUploadResponse{
		FileId: proto.String(meta.StoreName),
	}, nil
}

type FileDownloader struct {
	S3FileDownloader
	typ uint32
	c   *cache.Cache
}

func NewFileDownloader(typ uint32, c *cache.Cache) *FileDownloader {
	return &FileDownloader{typ: typ, c: c}
}

func (d *FileDownloader) BeforeDownload(ctx *gin.Context, request interface{}) (*FileDownloadMeta, error) {
	fileid := ctx.Request.URL.Query().Get("file_id")
	if len(fileid) == 0 {
		return nil, errs.New(constants.ErrParam, "no fileid found")
	}

	ifileinfo, exist, err := cacheGetFileMeta(ctx, d.c, fileid, func() (interface{}, bool, error) {
		daoRsp, exist, err := dao.MediaInfoDao.GetMedia(ctx, &model.GetMediaRequest{
			FileType: uint32(d.typ),
			Hash:     fileid,
		})
		if err != nil {
			return nil, false, err
		}
		if !exist {
			return nil, false, nil
		}
		return daoRsp.Item, true, nil
	})
	if err != nil {
		return nil, errs.Wrap(constants.ErrStorage, "cache get file meta fail", err).WithDebugMsg("type:%d", d.typ)
	}
	if !exist {
		return nil, errs.New(constants.ErrNotFound, "not found file meta").WithDebugMsg("filetype:%d", d.typ)
	}
	fileinfo := ifileinfo.(*model.MediaItem)
	return &FileDownloadMeta{
		FileId:   fileid,
		FileName: fileinfo.FileName,
		FileSize: int64(fileinfo.FileSize),
	}, nil
}
