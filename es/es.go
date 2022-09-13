package es

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/xxxsen/common/errs"

	"github.com/olivere/elastic/v7"
)

var Client *EsClient

const (
	DefaultVersion = "v1"
)

func Init(opts ...Option) error {
	client, err := New(opts...)
	if err != nil {
		return err
	}
	Client = client
	return nil
}

type EsClient struct {
	*elastic.Client
}

func New(opts ...Option) (*EsClient, error) {
	c := &config{
		user:    "elastic",
		timeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.urls) == 0 {
		return nil, errs.New(errs.ErrParam, "no es host found")
	}
	if len(c.user) == 0 || len(c.password) == 0 {
		return nil, errs.New(errs.ErrParam, "should set user/password")
	}
	httpClient := &http.Client{
		Timeout: c.timeout,
	}
	client, err := elastic.NewClient(
		elastic.SetBasicAuth(c.user, c.password),
		elastic.SetURL(c.urls...),
		elastic.SetHttpClient(httpClient),
	)
	if err != nil {
		return nil, errs.Wrap(errs.ErrES, "init es client fail", err)
	}
	for _, host := range c.urls {
		_, _, err := client.Ping(host).Do(context.Background())
		if err != nil {
			return nil, errs.Wrap(errs.ErrES, "ping host fail", err).WithDebugMsg("host:%s", host)
		}
	}
	return &EsClient{Client: client}, nil
}

func GetSearchResult(ctx context.Context, client *EsClient, searcher ISearcher) ([]interface{}, uint32, error) {
	ss := client.Search().Index(searcher.Index()).RestTotalHitsAsInt(true)
	esQuery := searcher.BuildQuery()
	ss = ss.Query(esQuery)
	countRes, err := ss.From(0).Size(0).Do(ctx)
	if err != nil {
		return nil, 0, errs.Wrap(errs.ErrES, "search for total fail", err)
	}
	total := uint32(countRes.TotalHits())
	if sorter := searcher.BuildSorter(); len(sorter) > 0 {
		ss.SortBy(sorter...)
	}
	results, err := ss.From(0).Size(searcher.Offset() + searcher.Limit()).TrackTotalHits(false).Do(ctx)
	if err != nil {
		return nil, 0, errs.Wrap(errs.ErrES, "search for data fail", err)
	}
	elemType := reflect.TypeOf(searcher.ObjectPtr()).Elem()
	hits := results.Hits.Hits
	if len(hits) < searcher.Offset() {
		return nil, total, nil
	}
	hits = hits[searcher.Offset():]
	rs := make([]interface{}, 0, len(hits))
	for _, hit := range hits {
		item := reflect.New(elemType).Interface()
		jsonData, _ := hit.Source.MarshalJSON()
		err := json.Unmarshal(jsonData, item)
		if err != nil {
			return nil, 0, errs.Wrap(errs.ErrES, "decode json fail", err)
		}
		rs = append(rs, item)
	}
	return rs, total, nil
}
