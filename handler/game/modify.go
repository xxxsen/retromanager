package game

import (
	"net/http"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

func ModifyGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.ModifyGameRequest)
	if req.Item == nil {
		return http.StatusOK, errs.New(constants.ErrParam, "invalid item"), nil
	}
	daoReq := &model.ModifyGameRequest{
		GameID: req.GetGameId(),
		State:  proto.Uint32(dao.GameStateNormal),
		Modify: &model.ModifyInfo{
			Platform:    req.Item.Platform,
			DisplayName: req.Item.DisplayName,
			FileSize:    req.Item.FileSize,
			Hash:        req.Item.Hash,
			Desc:        req.Item.Desc,
			ExtInfo:     nil,
			DownKey:     req.Item.DownKey,
		},
	}
	if req.Item.Extinfo != nil {
		raw, err := proto.Marshal(req.Item.Extinfo)
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrMarshal, "encode extinfo", err), nil
		}
		daoReq.Modify.ExtInfo = raw
	}
	rs, err := dao.GameInfoDao.ModifyGame(ctx, daoReq)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "modify db fail", err), nil
	}
	if rs.AffectRows == 0 {
		return http.StatusOK, errs.New(constants.ErrParam, "gameid not found"), nil
	}
	return http.StatusOK, nil, nil
}
