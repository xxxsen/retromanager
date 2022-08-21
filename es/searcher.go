package es

import (
	"github.com/olivere/elastic/v7"
)

type ISearcher interface {
	BuildQuery() *elastic.BoolQuery
}
