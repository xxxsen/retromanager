package model

import (
	"retromanager/proto/retromanager/gameinfo"

	"github.com/xxxsen/errs"

	"google.golang.org/protobuf/proto"
)

type ListQuery struct {
	ID         *uint64
	Platform   *uint32
	State      *uint32
	UpdateTime []uint64
}

type OrderByField string

const (
	OrderByCreateTime OrderByField = "create_time"
	OrderByUpdateTime OrderByField = "update_time"
)

type OrderBy struct {
	Field OrderByField
	Asc   bool
}

type GetGameRequest struct {
	GameId uint64
}

type GetGameResponse struct {
	Item *GameItem
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
	FileSize    uint64
	Desc        string
	CreateTime  uint64
	UpdateTime  uint64
	Hash        string
	DownKey     string
	ExtInfo     []byte
	FileName    string
}

func (item *GameItem) ToPBItem() (*gameinfo.GameInfo, error) {
	info := &gameinfo.GameInfo{
		Id:          proto.Uint64(item.ID),
		Platform:    proto.Uint32(item.Platform),
		DisplayName: proto.String(item.DisplayName),
		FileSize:    proto.Uint64(item.FileSize),
		Desc:        proto.String(item.Desc),
		CreateTime:  proto.Uint64(item.CreateTime),
		UpdateTime:  proto.Uint64(item.UpdateTime),
		Hash:        proto.String(item.Hash),
		DownKey:     proto.String(item.DownKey),
		Extinfo:     &gameinfo.GameExtInfo{},
		FileName:    proto.String(item.FileName),
	}
	if err := proto.Unmarshal(item.ExtInfo, info.Extinfo); err != nil {
		return nil, errs.Wrap(errs.ErrUnmarshal, "decode game extinfo", err)
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
	FileSize    *uint64
	Hash        *string
	Desc        *string
	ExtInfo     []byte
	DownKey     *string
	State       *uint32
	FileName    *string
}

type ModifyGameRequest struct {
	GameID uint64
	State  *uint32
	Modify *ModifyInfo
}

type ModifyGameResponse struct {
	AffectRows int64
}

type DeleteGameRequest struct {
	GameID uint64
}

type DeleteGameResponse struct {
	AffectRows int64
}
