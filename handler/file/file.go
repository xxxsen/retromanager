package file

import (
	"context"
	"fmt"
	"io"
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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xxxsen/log"
	"github.com/yitter/idgenerator-go/idgen"
	"google.golang.org/protobuf/proto"
)

var fileCache, _ = cache.New(20000)

var ImageUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeImage), fileCache, ImageExtChecker))
var VideoUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeVideo), fileCache, VideoExtChecker))
var FileUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeFile), fileCache, nil))
var FileDownload = CommonFileDownload(NewFileDownloader(fileCache))

type TypeCheckFunc func(meta *FileUploadMeta) error

func ExtNameChecker(exts ...string) TypeCheckFunc {
	valid := map[string]interface{}{}
	for _, ext := range exts {
		valid[strings.ToLower(ext)] = true
	}
	return func(meta *FileUploadMeta) error {
		ext := strings.ToLower(filepath.Ext(meta.FileName))
		if _, ok := valid[ext]; ok {
			return nil
		}
		return errs.New(constants.ErrParam, "not support ext:%s", ext)
	}
}

var ImageExtChecker = ExtNameChecker(".jpg", ".png")
var VideoExtChecker = ExtNameChecker(".mp4")

type FileUploadMeta struct {
	Reader   io.ReadSeekCloser
	Hash     string
	FileName string
	DownKey  string
	FileSize int64
	MD5      string
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
	md5, exist := ctx.GetPostForm("md5")
	if !exist {
		return nil, false, errs.New(constants.ErrParam, "md5 not found")
	}
	if header.Size > constants.MaxPostUploadSize {
		file.Close()
		return nil, false, errs.New(constants.ErrParam, "file size out of limit")
	}
	return &FileUploadMeta{
		Reader:   file,
		FileName: header.Filename,
		MD5:      md5,
		DownKey:  utils.EncodeFileId(uint64(idgen.NextId())),
		FileSize: header.Size,
	}, true, nil
}

func (f *S3SmallFileUploader) OnUpload(ctx *gin.Context, meta *FileUploadMeta) error {
	if err := s3.Client.Upload(ctx, meta.DownKey, meta.Reader, meta.FileSize, meta.MD5); err != nil {
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
			defer func() {
				if meta != nil && meta.Reader != nil {
					meta.Reader.Close()
				}
			}()
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
	typ  uint32
	c    *cache.Cache
	ckfn TypeCheckFunc
}

func NewFileUploader(typ uint32, c *cache.Cache, ckfn TypeCheckFunc) *FileUploader {
	if ckfn == nil {
		ckfn = func(meta *FileUploadMeta) error { return nil }
	}
	return &FileUploader{
		typ:  typ,
		c:    c,
		ckfn: ckfn,
	}
}

func (uploader *FileUploader) BeforeUpload(ctx *gin.Context, request interface{}) (*FileUploadMeta, bool, error) {
	meta, _, err := uploader.S3SmallFileUploader.BeforeUpload(ctx, request)
	if err != nil {
		return nil, false, err
	}
	meta.DownKey = fmt.Sprintf("%d_%s", uploader.typ, meta.DownKey)
	if err := uploader.ckfn(meta); err != nil {
		return nil, false, errs.Wrap(constants.ErrParam, "meta check not pass", err)
	}

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
			FileSize:   uint64(meta.FileSize),
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
