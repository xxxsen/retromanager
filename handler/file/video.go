package file

import (
	"retromanager/constants"
	"retromanager/handler/utils"
	"retromanager/model"

	"github.com/gin-gonic/gin"
)

var VideoUpload = postUploader(constants.MaxPostUploadVideoSize,
	func(ctx *gin.Context) string {
		return utils.MustGetConfig(ctx).BucketInfo.VideoBucket
	}, model.FileTypeVideo)
