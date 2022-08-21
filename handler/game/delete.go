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
		Query: &model.DeleteQuery{
			ID: req.GameId,
		},
	}
	_, err := dao.GameInfoDao.DeleteGame(ctx, daoReq)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "delete db fail", err), nil
	}
	return http.StatusOK, nil, nil
}
