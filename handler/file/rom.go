package file

import (
	"retromanager/cache"
	"retromanager/model"
)

var gameCache, _ = cache.New(10000)
var RomUpload = CommonFilePostUpload(NewFileUploader(uint32(model.FileTypeRom), gameCache))
var RomDownload = CommonFileDownload(NewFileDownloader(uint32(model.FileTypeRom), gameCache))
