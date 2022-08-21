package model

import "github.com/olivere/elastic/v7"

type RangeSearch struct {
	Gte string
	Lte string
}

type FilterSearch struct {
	Field string
	Value string
}

type SearchOrderBy struct {
	Field string
	Order string
}

type PageInfo struct {
	PageId   uint32
	PageSize uint32
}

type Searcher struct {
	Range   []RangeSearch
	Filter  []FilterSearch
	Page    *PageInfo
	OrderBy *SearchOrderBy
}

func (s *Searcher) BuildQuery() *elastic.BoolQuery {
	//TODO: impl it
	panic(1)
}
