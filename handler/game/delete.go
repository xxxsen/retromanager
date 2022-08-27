package game

import (
	"net/http"
	"retromanager/dao"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/xxxsen/errs"

	"github.com/gin-gonic/gin"
)

func DeleteGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.DeleteGameRequest)
	daoReq := &model.DeleteGameRequest{
		GameID: req.GetGameId(),
	}
	rs, err := dao.GameInfoDao.DeleteGame(ctx, daoReq)
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrDatabase, "delete db fail", err), nil
	}
	if rs.AffectRows != 1 {
		return http.StatusOK, errs.New(errs.ErrParam, "gameid not found"), nil
	}
	return http.StatusOK, nil, nil
}
