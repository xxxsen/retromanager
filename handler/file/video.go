package file

import (
	"retromanager/constants"
	"retromanager/handler/utils"
	"retromanager/model"

	"github.com/gin-gonic/gin"
)

var videoBucketGetter = func(ctx *gin.Context) string {
	return utils.MustGetConfig(ctx).BucketInfo.VideoBucket
}

var VideoUpload = postUploader(constants.MaxPostUploadVideoSize,
	videoBucketGetter, uint32(model.FileTypeVideo))

var VideoDownload = mediaFileDownload(videoBucketGetter, uint32(model.FileTypeVideo), fileDownloadRequestToHash)
