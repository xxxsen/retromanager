package utils

import (
	"retromanager/es"
	"retromanager/proto/retromanager/gameinfo"

	"github.com/xxxsen/errs"
)

func PBSearchParamsToEsSearchParams(s *gameinfo.SearchParam) (*es.SearchParam, error) {
	p := &es.SearchParam{
		FilterList: []*es.FilterValue{},
		RangeList:  []*es.RangeValue{},
		SortList:   []*es.SortValue{},
		Offset:     s.GetOffset(),
		Limit:      s.GetLimit(),
	}
	for _, item := range s.FilterList {
		if !es.IsFieldValid(item.GetField()) {
			return nil, errs.New(errs.ErrParam, "filter field not allow, field:%s", item.GetField())
		}
		p.FilterList = append(p.FilterList, &es.FilterValue{
			Field: item.GetField(),
			Value: item.GetValue(),
		})
	}
	for _, item := range s.RangeList {
		if !es.IsFieldValid(item.GetField()) {
			return nil, errs.New(errs.ErrParam, "range field not allow, field:%s", item.GetField())
		}
		p.RangeList = append(p.RangeList, &es.RangeValue{
			Field: item.GetField(),
			Left:  item.GetLeft(),
			Right: item.GetRight(),
		})
	}
	for _, item := range s.SortList {
		if !es.IsFieldValid(item.GetField()) {
			return nil, errs.New(errs.ErrParam, "sort field not allow, field:%s", item.GetField())
		}
		p.SortList = append(p.SortList, &es.SortValue{
			Field: item.GetField(),
			Asc:   item.GetAsc(),
		})
	}
	return p, nil
}
