package file

import (
	"fmt"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/handler/utils"
	"retromanager/model"

	"github.com/gin-gonic/gin"
)

func typeKey(typ uint32, hash string) string {
	return fmt.Sprintf("%d:%s", typ, hash)
}

var imageBucketGetter = func(ctx *gin.Context) string {
	return utils.MustGetConfig(ctx).BucketInfo.ImageBucket
}

var ImageUpload = postUploader(constants.MaxPostUploadImageSize, imageBucketGetter, model.FileTypeImage)

func fileDownloadRequestToHash(ctx *gin.Context, request interface{}) (string, error) {
	req, ok := request.(*FileDownloadRequest)
	if !ok {
		return "", errs.New(constants.ErrServiceInternal, "invalid request type")
	}
	return req.FileId, nil
}

var ImageDownload = mediaFileDownload(imageBucketGetter, model.FileTypeImage, fileDownloadRequestToHash)
