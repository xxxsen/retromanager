package es

import (
	"github.com/olivere/elastic/v7"
)

type ISearcher interface {
	BuildQuery() *elastic.BoolQuery
	BuildSorter() []elastic.Sorter
	Limit() int
	Offset() int
	Index() string
	ObjectPtr() interface{}
}

type FilterValue struct {
	Field string
	Value string
}

type RangeValue struct {
	Field string
	Left  string
	Right string
}

type SortValue struct {
	Field string
	Asc   bool
}

type SearchParam struct {
	FilterList []*FilterValue
	RangeList  []*RangeValue
	SortList   []*SortValue
	Offset     uint32
	Limit      uint32
}

type Searcher struct {
	q       *elastic.BoolQuery
	sorters []elastic.Sorter
	limit   int
	offset  int
	index   string
	objptr  interface{}
}

func buildSorter(s *Searcher, sorters []*SortValue) {
	if len(sorters) == 0 {
		return
	}
	lst := make([]elastic.Sorter, 0, len(sorters))
	for _, sorter := range sorters {
		f := elastic.NewFieldSort(sorter.Field).Order(sorter.Asc)
		lst = append(lst, f)
	}
	s.SetSorter(lst...)
}

func buildRange(s *Searcher, ranges []*RangeValue) {
	if len(ranges) == 0 {
		return
	}
	for _, r := range ranges {
		s.q.Must(elastic.NewRangeQuery(r.Field).Lte(r.Right).Gte(r.Left))
	}
}

func buildFilter(s *Searcher, filters []*FilterValue) {
	if len(filters) == 0 {
		return
	}
	for _, filter := range filters {
		s.q.Must(elastic.NewTermQuery(filter.Field, filter.Value))
	}
}

func FromSearchParam(param *SearchParam) *Searcher {
	s := NewSearcher()
	s.SetOffset(int(param.Offset))
	s.SetLimit(int(param.Limit))
	buildSorter(s, param.SortList)
	buildRange(s, param.RangeList)
	buildFilter(s, param.FilterList)
	return s
}

func NewSearcher() *Searcher {
	return &Searcher{q: elastic.NewBoolQuery()}
}

func (s *Searcher) SetQuery(q *elastic.BoolQuery) *Searcher {
	s.q = q
	return s
}

func (s *Searcher) BuildQuery() *elastic.BoolQuery {
	return s.q
}

func (s *Searcher) SetSorter(sorters ...elastic.Sorter) *Searcher {
	s.sorters = sorters
	return s
}

func (s *Searcher) BuildSorter() []elastic.Sorter {
	return s.sorters
}

func (s *Searcher) SetLimit(v int) *Searcher {
	s.limit = v
	return s
}

func (s *Searcher) Limit() int {
	return s.limit
}

func (s *Searcher) SetOffset(v int) *Searcher {
	s.offset = v
	return s
}

func (s *Searcher) Offset() int {
	return s.offset
}

func (s *Searcher) SetIndex(v string) *Searcher {
	s.index = v
	return s
}

func (s *Searcher) Index() string {
	return s.index
}

func (s *Searcher) SetObjectPtr(v interface{}) *Searcher {
	s.objptr = v
	return s
}

func (s *Searcher) ObjectPtr() interface{} {
	return s.objptr
}
