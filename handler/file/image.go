package file

import (
	"fmt"
	"retromanager/constants"
	"retromanager/handler/utils"
	"retromanager/model"

	"github.com/gin-gonic/gin"
)

func typeKey(typ uint32, hash string) string {
	return fmt.Sprintf("%d:%s", typ, hash)
}

var ImageUpload = postUploader(constants.MaxPostUploadImageSize,
	func(ctx *gin.Context) string {
		return utils.MustGetConfig(ctx).BucketInfo.ImageBucket
	}, model.FileTypeImage)
