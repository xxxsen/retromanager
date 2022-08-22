package file

import (
	"retromanager/cache"
	"retromanager/model"
)

var videoCache, _ = cache.New(10000)

var VideoUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeVideo), videoCache))
var VideoDownload = CommonFileDownload(NewFileDownloader(uint32(model.FileTypeVideo), videoCache))
