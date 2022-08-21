package es

type esSearcher struct {
	searcher ISearcher
	table    string
	version  string
	ptr      interface{}
}

type SearchCallback func(item interface{}) error

func NewSearcher(searcher ISearcher) *esSearcher {
	return &esSearcher{
		searcher: searcher,
	}
}

func (s *esSearcher) WithTable(name string, ver string) *esSearcher {
	s.table = name
	s.version = ver
	return s
}

func (s *esSearcher) WithObjectPtr(ptr interface{}) *esSearcher {
	s.ptr = ptr
	return s
}

func (s *esSearcher) GetSearchResult(cb SearchCallback) (uint32, error) {
	//TODO: finish it
	panic(1)
}
