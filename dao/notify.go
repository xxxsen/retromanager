package dao

import (
	"context"
	"log"
	"runtime/debug"
)

type ActionType int

const (
	ActionCreate ActionType = 1
	ActionModify ActionType = 2
	ActionDelete ActionType = 3
)

type IDataNotifyer interface {
	Name() string
	OnChange(ctx context.Context, table string, action ActionType, id uint64)
}

func AsyncNotify(ctx context.Context, table string, action ActionType, id uint64, notifyers ...IDataNotifyer) {
	for _, notifyer := range notifyers {
		notifyer := notifyer
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("do async notify panic, name:%s, err:%v\nstack:%s", notifyer.Name(), err, string(debug.Stack()))
				}
			}()
			notifyer.OnChange(ctx, table, action, id)
		}()
	}
}
