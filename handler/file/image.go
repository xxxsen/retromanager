package file

import (
	"retromanager/cache"
	"retromanager/model"
)

var imageCache, _ = cache.New(10000)

var ImageUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeImage), imageCache))
var ImageDownload = CommonFileDownload(NewFileDownloader(uint32(model.FileTypeImage), imageCache))
