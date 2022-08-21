package file

import (
	"mime/multipart"
	"retromanager/cache"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/handler/utils"
	"retromanager/model"
	"time"

	"github.com/gin-gonic/gin"
)

//TODO: 合并video和image的代码
var VideoUpload = postUploader(constants.MaxPostUploadVideoSize,
	func(ctx *gin.Context) string {
		return utils.MustGetConfig(ctx).BucketInfo.VideoBucket
	}, func(ctx *gin.Context, hash string, meta *multipart.FileHeader) (bool, error) {
		if _, exist, _ := cache.Default().Get(ctx, typeKey(model.FileTypeVideo, hash)); exist {
			return true, nil
		}
		_, exist, err := dao.MediaInfoDao.GetMedia(ctx, &model.GetMediaRequest{
			Hash:     hash,
			FileType: model.FileTypeVideo,
		})
		if err != nil {
			return false, errs.Wrap(constants.ErrDatabase, "check video db", err)
		}
		return exist, nil
	}, func(ctx *gin.Context, hash string, meta *multipart.FileHeader) error {
		if _, err := dao.MediaInfoDao.CreateMedia(ctx, &model.CreateMediaRequest{
			Item: &model.MediaItem{
				FileName:   meta.Filename,
				Hash:       hash,
				FileSize:   uint64(meta.Size),
				CreateTime: uint64(time.Now().UnixMilli()),
				FileType:   model.FileTypeVideo,
			},
		}); err != nil {
			return errs.Wrap(constants.ErrDatabase, "create video fail", err)
		}
		_ = cache.Default().Set(ctx, typeKey(model.FileTypeVideo, hash), true, 0)
		return nil
	})
