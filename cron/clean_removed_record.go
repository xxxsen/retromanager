package cron

import (
	"context"
	"fmt"
	"retromanager/dao"
	"retromanager/es"
	"retromanager/esservice"
	"retromanager/model"
	"strings"

	"github.com/xxxsen/errs"
	"github.com/xxxsen/naivesvr/log"
	"github.com/xxxsen/runner"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type cleanRemovedRecordCron struct {
}

func newCleanRemovedRecordCron() *cleanRemovedRecordCron {
	return &cleanRemovedRecordCron{}
}

func (c *cleanRemovedRecordCron) Name() string {
	return "clean_removed_record"
}

func (c *cleanRemovedRecordCron) Run(ctx context.Context) error {
	rsp, err := dao.GameInfoDao.ListGame(ctx, &model.ListGameRequest{
		Query: &model.ListQuery{
			State: proto.Uint32(model.GameStateDelete),
		},
		NeedTotal: false,
		Offset:    0,
		Limit:     100,
	})
	if err != nil {
		return errs.Wrap(errs.ErrDatabase, "list game fail", err)
	}
	if len(rsp.List) == 0 {
		return nil
	}
	if err := c.doClean(ctx, rsp.List); err != nil {
		return errs.Wrap(errs.ErrStorage, "do clean fail", err)
	}
	return nil
}

func (c *cleanRemovedRecordCron) doClean(ctx context.Context, lst []*model.GameItem) error {
	run := runner.New(5)
	for _, item := range lst {
		item := item
		run.Add(fmt.Sprintf("clean_record_%d", item.ID), func(ctx context.Context) error {
			return c.cleanOneItem(ctx, item)
		})
	}
	return run.Run(ctx)
}

func (c *cleanRemovedRecordCron) cleanOneItem(ctx context.Context, item *model.GameItem) error {
	if err := esservice.RemoveRecord(ctx, es.Client, dao.GameInfoDao.Table(), item.ID); err != nil && !strings.Contains(err.Error(), "elastic: Error 404") {
		return err
	}
	if _, err := dao.GameInfoDao.DeleteGame(ctx, &model.DeleteGameRequest{
		GameID: item.ID,
	}); err != nil {
		return err
	}
	log.GetLogger(ctx).With(
		zap.String("name", "cleanRemovedRecordCron"),
		zap.Uint64("id", item.ID),
	).Info("clean record succ")
	return nil
}
