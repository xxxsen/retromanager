package model

type CreateFileRequest struct {
	Item *FileItem
}

type CreateFileResponse struct {
}

type GetFileRequest struct {
	DownKey string
}

type FileItem struct {
	Id         uint64
	FileName   string
	Hash       string
	FileSize   uint64
	CreateTime uint64
	DownKey    string
}

type GetFileResponse struct {
	Item *FileItem
}
