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

const (
	maxListLimit = 20
)

func ListGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.ListGameRequest)
	rsp := &gameinfo.ListGameResponse{}

	if req.GetLimit() > maxListLimit {
		return http.StatusOK, errs.New(constants.ErrParam, "invalid params").WithDebugMsg("limit invalid"), nil
	}
	listRsp, err := dao.GameInfoDao.ListGame(ctx, &model.ListGameRequest{
		Query:     &model.ListQuery{Platform: req.Platform},
		Order:     &model.OrderBy{Field: model.OrderByCreateTime, Asc: true},
		NeedTotal: true,
		Offset:    req.GetOffset(),
		Limit:     req.GetLimit(),
	})
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "list game fail", err), nil
	}
	rsp.Total = proto.Uint32(listRsp.Total)
	for _, item := range listRsp.List {
		info, err := item.ToPBItem()
		if err != nil {
			return http.StatusOK, errs.Wrap(constants.ErrUnmarshal, "decode item", err), nil
		}
		rsp.List = append(rsp.List, info)
	}
	return http.StatusOK, nil, rsp
}
