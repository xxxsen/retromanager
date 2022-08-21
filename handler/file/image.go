package file

import (
	"fmt"
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

func typeKey(typ uint32, hash string) string {
	return fmt.Sprintf("%d:%s", typ, hash)
}

var ImageUpload = postUploader(constants.MaxPostUploadImageSize,
	func(ctx *gin.Context) string {
		return utils.MustGetConfig(ctx).BucketInfo.ImageBucket
	}, func(ctx *gin.Context, hash string, meta *multipart.FileHeader) (bool, error) {
		if _, exist, _ := cache.Default().Get(ctx, typeKey(model.FileTypeImage, hash)); exist {
			return true, nil
		}
		_, exist, err := dao.MediaInfoDao.GetMedia(ctx, &model.GetMediaRequest{
			Hash:     hash,
			FileType: model.FileTypeImage,
		})
		if err != nil {
			return false, errs.Wrap(constants.ErrDatabase, "check image db", err)
		}
		return exist, nil
	}, func(ctx *gin.Context, hash string, meta *multipart.FileHeader) error {
		if _, err := dao.MediaInfoDao.CreateMedia(ctx, &model.CreateMediaRequest{
			Item: &model.MediaItem{
				FileName:   meta.Filename,
				Hash:       hash,
				FileSize:   uint64(meta.Size),
				CreateTime: uint64(time.Now().UnixMilli()),
				FileType:   model.FileTypeImage,
			},
		}); err != nil {
			return errs.Wrap(constants.ErrDatabase, "create image fail", err)
		}
		_ = cache.Default().Set(ctx, typeKey(model.FileTypeImage, hash), true, 0)
		return nil
	})
