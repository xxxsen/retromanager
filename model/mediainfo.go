package model

type FileType uint8

const (
	FileTypeVideo = 1
	FileTypeImage = 2
)

type CreateMediaRequest struct {
	Item *MediaItem
}

type CreateMediaResponse struct {
}

type GetMediaRequest struct {
	FileType uint32
	Hash     string
}

type MediaItem struct {
	Id         uint64
	FileName   string
	Hash       string
	FileSize   uint64
	CreateTime uint64
	FileType   uint32
}

type GetMediaResponse struct {
	Item *MediaItem
}
