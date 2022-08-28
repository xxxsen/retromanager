package es

import (
	"strings"

	"github.com/olivere/elastic/v7"
)

const (
	keySuffixWildcard = "wildcard"
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
	q        *elastic.BoolQuery
	sorters  []elastic.Sorter
	limit    int
	offset   int
	index    string
	objptr   interface{}
	keyword  map[string]bool
	wildcard map[string]bool
	match    map[string]bool
	rename   map[string]string
}

func (s *Searcher) list2boolMap(m map[string]bool, fields []string) {
	for _, field := range fields {
		m[field] = true
	}
}

func (s *Searcher) setRenameField(rename map[string]string) *Searcher {
	s.rename = rename
	return s
}

func (s *Searcher) setKeywordField(fields ...string) *Searcher {
	s.list2boolMap(s.keyword, fields)
	return s
}

func (s *Searcher) setWildcardField(fields ...string) *Searcher {
	s.list2boolMap(s.wildcard, fields)
	return s
}

func (s *Searcher) setMatchField(fields ...string) *Searcher {
	s.list2boolMap(s.match, fields)
	return s
}

func (s *Searcher) realname(field string) string {
	if v, ok := s.rename[field]; ok {
		return v
	}
	return field
}

func (s *Searcher) buildSorter(sorters []*SortValue) {
	if len(sorters) == 0 {
		return
	}
	lst := make([]elastic.Sorter, 0, len(sorters))
	for _, sorter := range sorters {
		f := elastic.NewFieldSort(s.realname(sorter.Field)).Order(sorter.Asc)
		lst = append(lst, f)
	}
	s.SetSorter(lst...)
}

func (s *Searcher) buildRange(ranges []*RangeValue) {
	if len(ranges) == 0 {
		return
	}
	for _, r := range ranges {
		s.q.Must(elastic.NewRangeQuery(s.realname(r.Field)).Lte(r.Right).Gte(r.Left))
	}
}

func (s *Searcher) buildFilter(filters []*FilterValue) {
	if len(filters) == 0 {
		return
	}
	for _, filter := range filters {
		if _, ok := s.keyword[filter.Field]; ok {
			s.q.Must(elastic.NewTermQuery(s.realname(filter.Field)+".keyword", filter.Value))
			continue
		}
		if _, ok := s.wildcard[filter.Field]; ok {
			s.q.Must(elastic.NewWildcardQuery(s.realname(filter.Field), "*"+filter.Value+"*"))
			continue
		}
		if _, ok := s.match[filter.Field]; ok {
			s.q.Must(elastic.NewMatchQuery(s.realname(filter.Field), filter.Value))
			continue
		}
		s.q.Must(elastic.NewTermQuery(s.realname(filter.Field), filter.Value))
	}
}

func IsFieldValid(field string) bool {
	return !strings.Contains(field, "#")
}

type SearchOption func(s *Searcher)

func WithRenameField(r map[string]string) SearchOption {
	return func(s *Searcher) {
		s.setRenameField(r)
	}
}

func WithKeywordField(fields ...string) SearchOption {
	return func(s *Searcher) {
		s.setKeywordField(fields...)
	}
}

func WithMatchField(fields ...string) SearchOption {
	return func(s *Searcher) {
		s.setMatchField(fields...)
	}
}

func WithWildcardField(fields ...string) SearchOption {
	return func(s *Searcher) {
		s.setWildcardField(fields...)
	}
}

func FromSearchParam(param *SearchParam, opts ...SearchOption) *Searcher {
	s := &Searcher{
		q:        elastic.NewBoolQuery(),
		keyword:  map[string]bool{},
		wildcard: map[string]bool{},
		match:    map[string]bool{},
		rename:   map[string]string{},
	}
	for _, opt := range opts {
		opt(s)
	}
	s.SetOffset(int(param.Offset))
	s.SetLimit(int(param.Limit))
	s.buildSorter(param.SortList)
	s.buildRange(param.RangeList)
	s.buildFilter(param.FilterList)
	return s
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
