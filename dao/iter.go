package dao

import "context"

type IterCallbackFunc func(ctx context.Context, rows []interface{}) (bool, error)

type IDBInterator interface {
	IterRows(ctx context.Context, table string, limit int32, cb IterCallbackFunc) error
}
