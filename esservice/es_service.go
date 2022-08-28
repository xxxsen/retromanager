package esservice

import (
	"context"
	"retromanager/es"

	"github.com/xxxsen/errs"

	"github.com/olivere/elastic/v7"
)

type IDGetter func(v interface{}) string
type SearchCallback func(v interface{}) error

func UpsertRecord(ctx context.Context, client *es.EsClient, table string, record interface{}, idGetter IDGetter) error {
	return UpsertRecords(ctx, client, table, []interface{}{record}, idGetter)
}

func UpsertRecords(ctx context.Context, client *es.EsClient, table string, records []interface{}, idGetter IDGetter) error {
	_, alias := es.Index(table, es.DefaultVersion)
	svr := elastic.NewBulkService(client.Client)
	for _, record := range records {
		req := elastic.NewBulkUpdateRequest().
			Index(alias).
			RetryOnConflict(2).
			Id(idGetter(record)).
			DocAsUpsert(true).
			Doc(record)
		svr.Add(req)
	}
	rsp, err := svr.Do(ctx)
	if err != nil {
		return errs.Wrap(errs.ErrES, "call bulk fail", err)
	}
	if fail := rsp.Failed(); len(fail) > 0 {
		return errs.New(errs.ErrES, "part request fail, err:%+v", fail[0])
	}
	return nil
}

func TryCreateIndex(ctx context.Context, client *es.EsClient, table string, mapping string) error {
	index, alias := es.Index(table, es.DefaultVersion)
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return errs.Wrap(errs.ErrES, "check index fail", err)
	}
	if exists {
		return nil
	}
	if _, err := client.CreateIndex(index).BodyString(mapping).Do(ctx); err != nil {
		return errs.Wrap(errs.ErrES, "create index fail", err)
	}

	if _, err := client.Alias().Add(index, alias).Do(ctx); err != nil {
		return errs.Wrap(errs.ErrES, "map alias fail", err)
	}
	return nil
}
