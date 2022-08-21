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
	videoBucketGetter, model.FileTypeVideo)

var VideoDownload = mediaFileDownload(videoBucketGetter, model.FileTypeVideo, fileDownloadRequestToHash)
