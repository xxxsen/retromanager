package model

import (
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/proto/retromanager/gameinfo"

	"google.golang.org/protobuf/proto"
)

type ListQuery struct {
	ID       *uint64
	Platform *uint32
}

type DeleteQuery struct {
	ID *uint64
}

type OrderByField string

const (
	OrderByCreateTime OrderByField = "create_time"
)

type OrderBy struct {
	Field OrderByField
	Asc   bool
}

type ListGameRequest struct {
	Query     *ListQuery
	Order     *OrderBy
	NeedTotal bool
	Offset    uint32
	Limit     uint32
}

type ListGameResponse struct {
	List  []*GameItem
	Total uint32
}

type GameItem struct {
	ID          uint64
	Platform    uint32
	DisplayName string
	FileName    string
	FileSize    uint64
	Desc        string
	CreateTime  uint64
	UpdateTime  uint64
	Hash        string
	ExtInfo     []byte
}

func (item *GameItem) ToPBItem() (*gameinfo.GameInfo, error) {
	info := &gameinfo.GameInfo{
		Id:          proto.Uint64(item.ID),
		Platform:    proto.Uint32(item.Platform),
		DisplayName: proto.String(item.DisplayName),
		FileName:    proto.String(item.FileName),
		FileSize:    proto.Uint64(item.FileSize),
		Desc:        proto.String(item.Desc),
		CreateTime:  proto.Uint64(item.CreateTime),
		UpdateTime:  proto.Uint64(item.UpdateTime),
		Hash:        proto.String(item.Hash),
		Extinfo:     &gameinfo.GameExtInfo{},
	}
	if err := proto.Unmarshal(item.ExtInfo, info.Extinfo); err != nil {
		return nil, errs.Wrap(constants.ErrUnmarshal, "decode game extinfo", err)
	}
	return info, nil
}

type CreateGameRequest struct {
	Item *GameItem
}

type CreateGameResponse struct {
	GameId uint64
}

type ModifyInfo struct {
	Platform    *uint32
	DisplayName *string
	FileName    *string
	FileSize    *uint64
	Hash        *string
	Desc        *string
	ExtInfo     []byte
}

type ModifyGameRequest struct {
	GameID uint64
	Modify *ModifyInfo
}

type ModifyGameResponse struct{}

type DeleteGameRequest struct {
	Query *DeleteQuery
}

type DeleteGameResponse struct {
}
