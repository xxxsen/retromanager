package esservice

import (
	"context"
	"fmt"
	"retromanager/es"
	"strings"

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

func TryCreateIndex(ctx context.Context, client *es.EsClient, table string) error {
	index, alias := es.Index(table, es.DefaultVersion)
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return errs.Wrap(errs.ErrES, "check index fail", err)
	}
	if exists {
		return nil
	}
	if _, err := client.CreateIndex(index).Do(ctx); err != nil {
		return errs.Wrap(errs.ErrES, "create index fail", err)
	}

	if _, err := client.Alias().Add(index, alias).Do(ctx); err != nil {
		return errs.Wrap(errs.ErrES, "map alias fail", err)
	}
	return nil
}

func RemoveIndex(ctx context.Context, client *es.EsClient, table string) error {
	index, alias := es.Index(table, es.DefaultVersion)
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return errs.Wrap(errs.ErrES, "check index fail", err)
	}
	if !exists {
		return nil
	}
	if _, err := client.Alias().Remove(index, alias).Do(ctx); err != nil && !strings.Contains(err.Error(), "elastic: Error 404 (Not Found)") {
		return errs.Wrap(errs.ErrES, "delete alias fail", err)
	}
	if _, err := client.DeleteIndex().Index([]string{index}).Do(ctx); err != nil {
		return errs.Wrap(errs.ErrES, "delete index fail", err)
	}
	return nil
}

func RemoveRecord(ctx context.Context, client *es.EsClient, table string, id uint64) error {
	_, alias := es.Index(table, es.DefaultVersion)
	_, err := client.Delete().Index(alias).Id(fmt.Sprintf("%d", id)).Do(ctx)
	if err != nil {
		return errs.Wrap(errs.ErrES, "remove record fail", err)
	}
	return nil
}
