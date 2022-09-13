package action

import (
	"context"
	"fmt"
	"retromanager/dao"
	"retromanager/es"
	"retromanager/esservice"
	"retromanager/model"
	"retromanager/retry"
	"time"

	"github.com/xxxsen/common/logutil"

	"go.uber.org/zap"
)

type DB2ESAction struct {
}

func NewDB2ESAction() *DB2ESAction {
	return &DB2ESAction{}
}

func (act *DB2ESAction) Name() string {
	return "db2es"
}

func (act *DB2ESAction) OnChange(ctx context.Context, table string, action model.ActionType, id uint64) {
	var caller = act.nop
	logger := logutil.GetLogger(ctx).With(
		zap.Int("action", int(action)), zap.Uint64("id", id), zap.String("table", table),
	)
	switch action {
	case model.ActionCreate:
		caller = act.onCreate
	case model.ActionDelete:
		caller = act.onDelete
	case model.ActionModify:
		caller = act.onChange
	default:
		logger.Error("unsupport action found")
	}
	caller(ctx, logger, table, action, id)
}

func (act *DB2ESAction) nop(ctx context.Context, logger *zap.Logger, table string, action model.ActionType, id uint64) {
}

func (act *DB2ESAction) onDelete(ctx context.Context, logger *zap.Logger, table string, action model.ActionType, id uint64) {
	logger.Info("recv delete event, skip")
}

func (act *DB2ESAction) onChange(ctx context.Context, logger *zap.Logger, table string, action model.ActionType, id uint64) {
	var rsp *model.GetGameResponse
	if err := retry.RetryDo(ctx, 2, 100*time.Millisecond, func(ctx context.Context) error {
		var err error
		var exist bool
		rsp, exist, err = dao.GameInfoDao.GetGame(ctx, &model.GetGameRequest{
			GameId: id,
		})
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf("not found gameid")
		}
		return nil
	}); err != nil {
		logger.With(zap.Error(err)).Error("call get game info fail")
		return
	}
	gameinfo, err := rsp.Item.ToPBItem()
	if err != nil {
		logger.With(zap.Error(err), zap.Any("item", rsp.Item)).Error("translate db item to pb item fail")
		return
	}
	if err := retry.RetryDo(ctx, 2, 100*time.Millisecond, func(ctx context.Context) error {
		if err := esservice.UpsertRecord(ctx, es.Client, dao.GameInfoDao.Table(), gameinfo, func(v interface{}) string {
			return fmt.Sprintf("%d", gameinfo.GetId())
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.With(zap.Error(err)).Error("call write es fail")
		return
	}
}

func (act *DB2ESAction) onCreate(ctx context.Context, logger *zap.Logger, table string, action model.ActionType, id uint64) {
	act.onChange(ctx, logger, table, action, id)
}
