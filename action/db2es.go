package action

import (
	"context"
	"retromanager/model"
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
	//TODO: finish it
	panic(1)
}
