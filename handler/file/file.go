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
	"retromanager/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xxxsen/log"
	"github.com/yitter/idgenerator-go/idgen"
	"google.golang.org/protobuf/proto"
)

var fileCache, _ = cache.New(20000)

var ImageUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeImage), fileCache))
var VideoUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeVideo), fileCache))
var RomUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeRom), fileCache))
var FileDownload = CommonFileDownload(NewFileDownloader(fileCache))

type FileUploadMeta struct {
	Data     []byte
	Hash     string
	FileName string
	DownKey  string
}

type ISmallFileUploader interface {
	BeforeUpload(ctx *gin.Context, request interface{}) (*FileUploadMeta, bool, error)
	OnUpload(ctx *gin.Context, meta *FileUploadMeta) error
	AfterUpload(ctx *gin.Context, realUpload bool, meta *FileUploadMeta) (interface{}, error)
}

type S3SmallFileUploader struct {
}

func (f *S3SmallFileUploader) BeforeUpload(ctx *gin.Context, request interface{}) (*FileUploadMeta, bool, error) {
	if err := ctx.Request.ParseMultipartForm(constants.MaxPostUploadSize); err != nil {
		return nil, false, errs.Wrap(constants.ErrIO, "parse form fail", err)
	}
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
		Data:     raw,
		FileName: header.Filename,
		DownKey:  utils.EncodeFileId(uint64(idgen.NextId())),
	}, true, nil
}

func (f *S3SmallFileUploader) OnUpload(ctx *gin.Context, meta *FileUploadMeta) error {
	if len(meta.Data) == 0 {
		return errs.New(constants.ErrParam, "no file data")
	}
	if err := s3.Client.Upload(ctx, meta.DownKey, bytes.NewReader(meta.Data), int64(len(meta.Data))); err != nil {
		return errs.Wrap(constants.ErrS3, "upload to s3 fail", err)
	}
	return nil
}

func (f *S3SmallFileUploader) AfterUpload(ctx *gin.Context, realUpload bool, meta *FileUploadMeta) (interface{}, error) {
	return nil, nil
}

type FileDownloadMeta struct {
	DownKey     string
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
	reader, err := s3.Client.Download(ctx, meta.DownKey)
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
	meta.DownKey = fmt.Sprintf("%d_%s", uploader.typ, meta.DownKey)
	_, exist, _ := uploader.c.Get(ctx, meta.DownKey)
	if exist {
		return meta, false, nil
	}
	return meta, true, nil
}

func (uploader *FileUploader) AfterUpload(ctx *gin.Context, realUpload bool, meta *FileUploadMeta) (interface{}, error) {
	if _, err := dao.FileInfoDao.CreateFile(ctx, &model.CreateFileRequest{
		Item: &model.FileItem{
			FileName:   meta.FileName,
			Hash:       meta.Hash,
			FileSize:   uint64(len(meta.Data)),
			CreateTime: uint64(time.Now().UnixMilli()),
			DownKey:    meta.DownKey,
		},
	}); err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "insert image to db fail", err)
	}
	return &gameinfo.FileUploadResponse{
		DownKey: proto.String(meta.DownKey),
	}, nil
}

type FileDownloader struct {
	S3FileDownloader
	c *cache.Cache
}

func NewFileDownloader(c *cache.Cache) *FileDownloader {
	return &FileDownloader{c: c}
}

func (d *FileDownloader) BeforeDownload(ctx *gin.Context, request interface{}) (*FileDownloadMeta, error) {
	downKey := ctx.Request.URL.Query().Get("down_key")
	if len(downKey) == 0 {
		return nil, errs.New(constants.ErrParam, "no fileid found")
	}

	ifileinfo, exist, err := cacheGetFileMeta(ctx, d.c, downKey, func() (interface{}, bool, error) {
		daoRsp, exist, err := dao.FileInfoDao.GetFile(ctx, &model.GetFileRequest{
			DownKey: downKey,
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
		return nil, errs.Wrap(constants.ErrStorage, "cache get file meta fail", err)
	}
	if !exist {
		return nil, errs.New(constants.ErrNotFound, "not found file meta")
	}
	fileinfo := ifileinfo.(*model.FileItem)
	return &FileDownloadMeta{
		DownKey:  downKey,
		FileName: fileinfo.FileName,
		FileSize: int64(fileinfo.FileSize),
	}, nil
}
