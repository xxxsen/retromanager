package game

import (
	"net/http"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/gin-gonic/gin"
)

func DeleteGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.DeleteGameRequest)
	daoReq := &model.DeleteGameRequest{
		GameID: req.GetGameId(),
	}
	rs, err := dao.GameInfoDao.DeleteGame(ctx, daoReq)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "delete db fail", err), nil
	}
	if rs.AffectRows != 1 {
		return http.StatusOK, errs.New(constants.ErrParam, "gameid not found"), nil
	}
	return http.StatusOK, nil, nil
}
