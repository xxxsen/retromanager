package client

import "retromanager/proto/retromanager/gameinfo"

type FileMeta struct {
	Path    string
	Name    string
	Size    int64
	MD5     string
	DownKey string
}

type UploadImageRequest struct {
	File string
}

type UploadImageResponse struct {
	Meta *FileMeta
}

type UploadVideoRequest struct {
	File string
}

type UploadVideoResponse struct {
	Meta *FileMeta
}

type UploadFileRequest struct {
	File string
}

type UploadFileResponse struct {
	Meta *FileMeta
}

type CreateGameRequest = gameinfo.CreateGameRequest
type CreateGameResponse = gameinfo.CreateGameResponse
