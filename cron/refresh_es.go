package cron

import (
	"context"
	"fmt"
	"retromanager/cache"
	"retromanager/dao"
	"retromanager/es"
	"retromanager/esservice"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"time"

	"github.com/xxxsen/common/logutil"

	"github.com/xxxsen/common/errs"

	"go.uber.org/zap"
)

const (
	maxScanPerBatch    = 200
	maxWriteESPerBatch = 20
)

type refreshESCron struct {
	c *cache.Cache
}

func newRefreshESCron() *refreshESCron {
	c, _ := cache.New(10000)
	return &refreshESCron{c: c}
}

func (c *refreshESCron) Name() string {
	return "db2es_refresher"
}

func (c *refreshESCron) Run(ctx context.Context) error {
	left := uint64(time.Now().Add(-30 * time.Minute).UnixMilli())
	right := uint64(time.Now().Add(-1 * time.Minute).UnixMilli())
	rsp, err := dao.GameInfoDao.ListGame(ctx, &model.ListGameRequest{
		Query: &model.ListQuery{
			UpdateTime: []uint64{left, right},
		},
		Order:     &model.OrderBy{Field: model.OrderByUpdateTime, Asc: false},
		NeedTotal: false,
		Offset:    0,
		Limit:     maxScanPerBatch,
	})
	if err != nil {
		return errs.Wrap(errs.ErrDatabase, "list db fail", err)
	}
	if err := c.doBusi(ctx, rsp.List); err != nil {
		return errs.Wrap(errs.ErrServiceInternal, "do refresh logic fail", err)
	}
	return nil
}

func (c *refreshESCron) makeKey(id uint64, ts uint64) string {
	return fmt.Sprintf("%d_%d", id, ts)
}

func (c *refreshESCron) doBusi(ctx context.Context, items []*model.GameItem) error {
	logger := logutil.GetLogger(ctx).With(zap.String("name", "refreshESCron.doBusi"))
	needWriteList := make([]*gameinfo.GameInfo, 0, len(items))
	for _, item := range items {
		key := c.makeKey(item.ID, item.UpdateTime)
		_, exist, _ := c.c.Get(ctx, key)
		if exist {
			continue
		}
		pbitem, err := item.ToPBItem()
		if err != nil {
			return errs.Wrap(errs.ErrUnknown, "translate to pb item fail", err)
		}
		needWriteList = append(needWriteList, pbitem)
	}
	for i := 0; i < len(needWriteList); i += maxWriteESPerBatch {
		sub := make([]interface{}, 0, maxWriteESPerBatch)
		for j := i; j < i+maxWriteESPerBatch && j < len(needWriteList); j++ {
			sub = append(sub, needWriteList[j])
		}
		if err := esservice.UpsertRecords(ctx, es.Client, dao.GameInfoDao.Table(), sub, func(v interface{}) string {
			return fmt.Sprintf("%d", v.(*gameinfo.GameInfo).GetId())
		}); err != nil {
			logger.With(zap.Error(err)).Error("write es fail")
			continue
		}
		for _, vitem := range sub {
			item := vitem.(*gameinfo.GameInfo)
			_ = c.c.Set(ctx, c.makeKey(item.GetId(), item.GetUpdateTime()), true, 0)
		}
		logger.With(zap.Int("size", len(sub))).Info("refresh db item to es succ")
	}
	return nil
}
