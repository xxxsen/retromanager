package game

import (
	"net/http"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/es"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/utils"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

const (
	maxPageSize = 2000
)

func SearchGame(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.SearchGameRequest)
	rsp := &gameinfo.SearchGameResponse{}
	if req.GetSearcher() == nil {
		return http.StatusOK, errs.New(constants.ErrParam, "invalid params"), nil
	}
	s := req.GetSearcher()
	if s.Page == nil {
		s.Page = &gameinfo.Page{
			Id:   proto.Uint32(0),
			Size: proto.Uint32(maxListLimit),
		}
	}
	if s.OrderBy == nil {
		s.OrderBy = &gameinfo.OrderBy{
			Field: proto.String(string(model.OrderByCreateTime)),
			Order: proto.String("asc"),
		}
	}
	if s.GetPage().GetSize() > maxListLimit || s.GetPage().GetId() > maxPageSize {
		return http.StatusOK, errs.New(constants.ErrParam, "invalid params").WithDebugMsg("size out of limit"), nil
	}
	searcher := utils.ConvertPBSearcherToEsSearcher(req.GetSearcher())
	cb := func(ptr interface{}) error {
		item := ptr.(*gameinfo.GameInfo)
		rsp.List = append(rsp.List, item)
		return nil
	}
	total, err := es.NewSearcher(searcher).WithTable("game_info_tab", "v1").
		WithObjectPtr(&gameinfo.GameInfo{}).GetSearchResult(cb)

	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrES, "search fail", err), nil
	}
	rsp.Total = proto.Uint32(total)
	return http.StatusOK, nil, rsp
}
