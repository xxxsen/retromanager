package utils

import (
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
)

func ConvertPBSearcherToEsSearcher(searcher *gameinfo.CommonSearch) *model.Searcher {
	rs := &model.Searcher{}
	if searcher.Page != nil {
		rs.Page = &model.PageInfo{
			PageId:   searcher.GetPage().GetId(),
			PageSize: searcher.GetPage().GetSize(),
		}
	}
	if searcher.OrderBy != nil {
		rs.OrderBy = &model.SearchOrderBy{
			Field: searcher.GetOrderBy().GetField(),
			Order: searcher.GetOrderBy().GetOrder(),
		}
	}
	for _, item := range searcher.Ranges {
		rs.Range = append(rs.Range, model.RangeSearch{
			Gte: item.GetGte(),
			Lte: item.GetLte(),
		})
	}
	for _, item := range searcher.Filters {
		rs.Filter = append(rs.Filter, model.FilterSearch{
			Field: item.GetField(),
			Value: item.GetValue(),
		})
	}
	return rs
}
