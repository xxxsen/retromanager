package game

import (
	"context"
	"fmt"
	"net/http"
	"retromanager/dao"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"time"

	"github.com/xxxsen/errs"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

func checkCreate(req *gameinfo.CreateGameRequest) error {
	if req.Item == nil {
		return fmt.Errorf("nil item")
	}
	item := req.Item
	if len(item.GetDesc()) == 0 {
		return fmt.Errorf("nil desc")
	}
	if len(item.GetDisplayName()) == 0 {
		return fmt.Errorf("nil display name")
	}
	if len(item.GetHash()) != 32 { //hex(md5)
		return fmt.Errorf("invalid hash")
	}
	if item.GetFileSize() == 0 {
		return fmt.Errorf("empty file")
	}
	if item.Extinfo == nil {
		return fmt.Errorf("nil extinfo")
	}
	if len(item.GetDownKey()) == 0 {
		return fmt.Errorf("down key not found")
	}
	if len(item.GetFileName()) == 0 {
		return fmt.Errorf("file name empty")
	}
	return nil
}

func checkIsGameExist(ctx context.Context, hash string) (uint64, bool, error) {
	rsp, err := dao.GameInfoDao.ListGame(ctx, &model.ListGameRequest{
		Query: &model.ListQuery{
			Hash: proto.String(hash),
		},
		NeedTotal: false,
		Offset:    0,
		Limit:     1,
	})
	if err != nil {
		return 0, false, err
	}
	if len(rsp.List) == 0 {
		return 0, false, nil
	}
	return rsp.List[0].ID, true, nil
}

func CreateGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.CreateGameRequest)
	rsp := &gameinfo.CreateGameResponse{}

	if err := checkCreate(req); err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrParam, "invalid params", err), nil
	}
	item := req.GetItem()
	gameid, isExist, err := checkIsGameExist(ctx, item.GetHash())
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrDatabase, "check game exist fail", err), nil
	}
	if isExist {
		rsp.IsGameExist = proto.Bool(true)
		rsp.GameId = proto.Uint64(gameid)
		return http.StatusOK, nil, rsp
	}
	now := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	extinfo, err := proto.Marshal(item.GetExtinfo())
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrMarshal, "encode ext info fail", err), nil
	}
	daoReq := &model.CreateGameRequest{
		Item: &model.GameItem{
			Platform:    item.GetPlatform(),
			DisplayName: item.GetDisplayName(),
			FileSize:    item.GetFileSize(),
			Desc:        item.GetDesc(),
			CreateTime:  now,
			UpdateTime:  now,
			Hash:        item.GetHash(),
			DownKey:     item.GetDownKey(),
			ExtInfo:     extinfo,
			FileName:    item.GetFileName(),
		},
	}
	daoRsp, err := dao.GameInfoDao.CreateGame(ctx, daoReq)
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrDatabase, "create game fail", err), nil
	}
	rsp.GameId = proto.Uint64(daoRsp.GameId)
	rsp.IsGameExist = proto.Bool(false)
	return http.StatusOK, nil, rsp
}
