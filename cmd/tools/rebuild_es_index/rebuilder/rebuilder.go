package rebuilder

import (
	"context"
	"fmt"
	"retromanager/dao"
	"retromanager/es"
	"retromanager/esservice"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/retry"
	"time"

	"github.com/xxxsen/common/errs"
)

type Rebuilder struct {
	gameService dao.GameInfoService
	esClient    *es.EsClient
}

func NewRebuilder(gameService dao.GameInfoService, esClient *es.EsClient) *Rebuilder {
	return &Rebuilder{
		gameService: gameService,
		esClient:    esClient,
	}
}

func (r *Rebuilder) Rebuild(ctx context.Context) error {
	if err := esservice.RemoveIndex(ctx, r.esClient, r.gameService.Table()); err != nil {
		return errs.Wrap(errs.ErrES, "remove es index fail", err)
	}
	if err := esservice.TryCreateIndex(ctx, r.esClient, r.gameService.Table()); err != nil {
		return errs.Wrap(errs.ErrES, "build es index fail", err)
	}
	if err := r.gameService.IterRows(ctx, dao.GameInfoDao.Table(), 2000, r.rebuild); err != nil {
		return errs.Wrap(errs.ErrDatabase, "iter table fail", err)
	}
	return nil
}

func (r *Rebuilder) rebuild(ctx context.Context, rows []interface{}) (bool, error) {
	for i := 0; i < len(rows); i += 20 {
		sub := make([]interface{}, 0, 20)
		for j := i; j < i+20 && j < len(rows); j++ {
			row := rows[j].(*model.GameItem)
			gameitem, err := row.ToPBItem()
			if err != nil {
				return false, errs.Wrap(errs.ErrUnknown, "model to pb fail", err)
			}
			sub = append(sub, gameitem)
		}
		if err := retry.RetryDo(ctx, 2, 500*time.Millisecond, func(ctx context.Context) error {
			return esservice.UpsertRecords(ctx, r.esClient, r.gameService.Table(), sub, func(v interface{}) string {
				return fmt.Sprintf("%d", v.(*gameinfo.GameInfo).GetId())
			})
		}); err != nil {
			return false, errs.Wrap(errs.ErrES, "write es fail", err)
		}
	}
	return true, nil
}
