package game

import (
	"fmt"
	"net/http"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"time"

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
	if len(item.GetFileName()) == 0 {
		return fmt.Errorf("nil filename")
	}
	if len(item.GetDisplayName()) == 0 {
		return fmt.Errorf("nil display name")
	}
	if len(item.GetHash()) == 0 {
		return fmt.Errorf("nil hash")
	}
	if item.GetFileSize() == 0 {
		return fmt.Errorf("empty file")
	}
	if item.Extinfo == nil {
		return fmt.Errorf("nil extinfo")
	}
	return nil
}

func CreateGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.CreateGameRequest)
	rsp := &gameinfo.CreateGameResponse{}

	if err := checkCreate(req); err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "invalid params", err), nil
	}
	item := req.GetItem()

	now := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	extinfo, err := proto.Marshal(item.GetExtinfo())
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrMarshal, "encode ext info fail", err), nil
	}
	daoReq := &model.CreateGameRequest{
		Item: &model.GameItem{
			Platform:    item.GetPlatform(),
			DisplayName: item.GetDisplayName(),
			FileName:    item.GetFileName(),
			FileSize:    item.GetFileSize(),
			Desc:        item.GetDesc(),
			CreateTime:  now,
			UpdateTime:  now,
			Hash:        item.GetHash(),
			ExtInfo:     extinfo,
		},
	}
	daoRsp, err := dao.GameInfoDao.CreateGame(ctx, daoReq)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "create game fail", err), nil
	}
	rsp.GameId = proto.Uint64(daoRsp.GameId)
	return http.StatusOK, nil, rsp
}
