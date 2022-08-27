package game

import (
	"net/http"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/es"
	"retromanager/handler/utils"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/xxxsen/errs"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

var validSearchField = map[string]bool{
	"id":           true,
	"platform":     true,
	"display_name": true,
	"file_size":    true,
	"desc":         true,
	"create_time":  true,
	"update_time":  true,
	"hash":         true,
	"rating":       true,
	"players":      true,
	"genre":        true,
}

var renameFieldMap = map[string]string{
	"genre":        "extinfo.genre",
	"players":      "extinfo.players",
	"rating":       "extinfo.rating",
	"display_name": es.WrapWildcard("display_name"),
	"desc":         es.WrapWildcard("desc"),
}

func SearchGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.SearchGameRequest)
	rsp := &gameinfo.SearchGameResponse{}
	if req.Param == nil {
		return http.StatusOK, errs.New(errs.ErrParam, "nil param"), nil
	}
	if req.GetParam().GetOffset() > constants.MaxGameSearchOffset ||
		req.GetParam().GetLimit() > constants.MaxGameSearchLimit {
		return http.StatusOK, errs.New(errs.ErrParam, "size out of limit"), nil
	}
	if req.GetParam().GetLimit() == 0 {
		req.GetParam().Limit = proto.Uint32(constants.MaxGameSearchLimit)
	}

	param, err := utils.PBSearchParamsToEsSearchParams(req.GetParam(), renameFieldMap, &validSearchField)
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrParam, "param translate fail", err), nil
	}
	searcher := es.FromSearchParam(param)
	_, alias := es.Index(dao.GameInfoDao.Table(), es.DefaultVersion)
	searcher.SetIndex(alias)
	searcher.SetObjectPtr(&gameinfo.GameInfo{})
	result, total, err := es.GetSearchResult(ctx, es.Client, searcher)
	if err != nil {
		return http.StatusOK, errs.Wrap(errs.ErrES, "search es fail", err), nil
	}
	rsp.Total = proto.Uint32(total)
	for _, item := range result {
		rsp.List = append(rsp.List, item.(*gameinfo.GameInfo))
	}
	return http.StatusOK, nil, rsp
}
