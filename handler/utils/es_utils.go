package utils

import (
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/es"
	"retromanager/proto/retromanager/gameinfo"
)

func fieldRename(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return key
}

func isFieldValid(m *map[string]bool, key string) bool {
	if m == nil {
		return true
	}
	if _, ok := (*m)[key]; ok {
		return true
	}
	return false
}

func PBSearchParamsToEsSearchParams(s *gameinfo.SearchParam, renameMap map[string]string, validFieldMap *map[string]bool) (*es.SearchParam, error) {
	p := &es.SearchParam{
		FilterList: []*es.FilterValue{},
		RangeList:  []*es.RangeValue{},
		SortList:   []*es.SortValue{},
		Offset:     s.GetOffset(),
		Limit:      s.GetLimit(),
	}
	for _, item := range s.FilterList {
		if !isFieldValid(validFieldMap, item.GetField()) {
			return nil, errs.New(constants.ErrParam, "filter field not allow, field:%s", item.GetField())
		}
		p.FilterList = append(p.FilterList, &es.FilterValue{
			Field: fieldRename(renameMap, item.GetField()),
			Value: item.GetValue(),
		})
	}
	for _, item := range s.RangeList {
		if !isFieldValid(validFieldMap, item.GetField()) {
			return nil, errs.New(constants.ErrParam, "range field not allow, field:%s", item.GetField())
		}
		p.RangeList = append(p.RangeList, &es.RangeValue{
			Field: fieldRename(renameMap, item.GetField()),
			Left:  item.GetLeft(),
			Right: item.GetRight(),
		})
	}
	for _, item := range s.SortList {
		if !isFieldValid(validFieldMap, item.GetField()) {
			return nil, errs.New(constants.ErrParam, "sort field not allow, field:%s", item.GetField())
		}
		p.SortList = append(p.SortList, &es.SortValue{
			Field: fieldRename(renameMap, item.GetField()),
			Asc:   item.GetAsc(),
		})
	}
	return p, nil
}
