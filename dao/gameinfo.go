package dao

import (
	"context"
	"database/sql"
	"fmt"
	"retromanager/constants"
	"retromanager/db"
	"retromanager/errs"
	"retromanager/model"
	"time"

	"github.com/didi/gendry/builder"
)

var gameinfoFields = []string{
	"id", "platform", "display_name", "file_size", "desc", "create_time", "update_time", "hash", "extinfo", "down_key",
}

var GameInfoDao = NewGameInfoDao()

type GameInfoService interface {
	GetGame(ctx context.Context, req *model.GetGameRequest) (*model.GetGameResponse, bool, error)
	ListGame(ctx context.Context, req *model.ListGameRequest) (*model.ListGameResponse, error)
	CreateGame(ctx context.Context, req *model.CreateGameRequest) (*model.CreateGameResponse, error)
	ModifyGame(ctx context.Context, req *model.ModifyGameRequest) (*model.ModifyGameResponse, error)
	DeleteGame(ctx context.Context, req *model.DeleteGameRequest) (*model.DeleteGameResponse, error)
}

type gameinfoImpl struct {
}

func NewGameInfoDao() *gameinfoImpl {
	return &gameinfoImpl{}
}

func (d *gameinfoImpl) Client() *sql.DB {
	return db.GetGameDB()
}

func (d *gameinfoImpl) Table() string {
	return "game_info_tab"
}

func (d *gameinfoImpl) Fields() []string {
	return gameinfoFields
}

func (d *gameinfoImpl) buildTotal(ctx context.Context, where map[string]interface{}) (uint32, error) {
	delete(where, "_limit")
	delete(where, "_orderby")
	sql, args, err := builder.BuildSelect(d.Table(), where, []string{"count(*)"})
	if err != nil {
		return 0, errs.Wrap(constants.ErrParam, "build", err)
	}
	rows, err := d.Client().QueryContext(ctx, sql, args...)
	if err != nil {
		return 0, errs.Wrap(constants.ErrDatabase, "query total", err)
	}
	defer rows.Close()
	var total uint32
	for rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, errs.Wrap(constants.ErrDatabase, "scan total", err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, errs.Wrap(constants.ErrDatabase, "scan total", err)
	}
	return total, nil
}

func (d *gameinfoImpl) GetGame(ctx context.Context, req *model.GetGameRequest) (*model.GetGameResponse, bool, error) {
	subReq := &model.ListGameRequest{
		Query:     &model.ListQuery{ID: &req.GameId},
		NeedTotal: false,
		Offset:    0,
		Limit:     1,
	}
	subRsp, err := d.ListGame(ctx, subReq)
	if err != nil {
		return nil, false, err
	}
	if len(subRsp.List) == 0 {
		return nil, false, nil
	}
	return &model.GetGameResponse{Item: subRsp.List[0]}, true, nil
}

func (d *gameinfoImpl) ListGame(ctx context.Context, req *model.ListGameRequest) (*model.ListGameResponse, error) {
	where := map[string]interface{}{
		"_limit": []uint{uint(req.Offset), uint(req.Limit)},
	}
	if req.Query != nil {
		if req.Query.ID != nil {
			where["id"] = *req.Query.ID
		}
		if req.Query.Platform != nil {
			where["platform"] = *req.Query.Platform
		}
	}
	if req.Order != nil {
		order := "asc"
		if !req.Order.Asc {
			order = "desc"
		}
		where["_orderby"] = fmt.Sprintf("%s %s", req.Order.Field, order)
	}

	sql, args, err := builder.BuildSelect(d.Table(), where, d.Fields())
	if err != nil {
		return nil, errs.Wrap(constants.ErrParam, "build select", err)
	}
	client := d.Client()
	rows, err := client.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "query ctx fail", err)
	}
	defer rows.Close()
	lst := make([]*model.GameItem, 0, req.Limit)
	for rows.Next() {
		item := &model.GameItem{}
		if err := rows.Scan(&item.ID, &item.Platform, &item.DisplayName, &item.FileSize, &item.Desc, &item.CreateTime,
			&item.UpdateTime, &item.Hash, &item.ExtInfo, &item.DownKey); err != nil {

			return nil, errs.Wrap(constants.ErrDatabase, "scan", err)
		}
		lst = append(lst, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "scan", err)
	}
	var total uint32
	if req.NeedTotal {
		total, err = d.buildTotal(ctx, where)
		if err != nil {
			return nil, errs.Wrap(constants.ErrDatabase, "build total", err)
		}
	}
	return &model.ListGameResponse{List: lst, Total: total}, nil
}

func (d *gameinfoImpl) CreateGame(ctx context.Context, req *model.CreateGameRequest) (*model.CreateGameResponse, error) {
	item := req.Item
	data := []map[string]interface{}{
		{
			"id":           item.ID,
			"platform":     item.Platform,
			"display_name": item.DisplayName,
			"desc":         item.Desc,
			"create_time":  item.CreateTime,
			"update_time":  item.UpdateTime,
			"hash":         item.Hash,
			"extinfo":      item.ExtInfo,
			"down_key":     item.DownKey,
		},
	}
	sql, args, err := builder.BuildInsert(d.Table(), data)
	if err != nil {
		return nil, errs.Wrap(constants.ErrParam, "build insert", err)
	}
	rs, err := d.Client().ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "exec insert", err)
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "get insert id", err)
	}
	return &model.CreateGameResponse{GameId: uint64(id)}, nil
}

func (d *gameinfoImpl) ModifyGame(ctx context.Context, req *model.ModifyGameRequest) (*model.ModifyGameResponse, error) {
	if req.Modify == nil {
		return nil, errs.New(constants.ErrParam, "nil modify")
	}
	where := map[string]interface{}{
		"id": req.GameID,
	}
	update := map[string]interface{}{
		"update_time": time.Now().UnixNano() / int64(time.Millisecond),
	}
	if req.Modify.Desc != nil {
		update["desc"] = *req.Modify.Desc
	}
	if req.Modify.DisplayName != nil {
		update["display_name"] = *req.Modify.DisplayName
	}
	if req.Modify.ExtInfo != nil {
		update["extinfo"] = req.Modify.ExtInfo
	}
	if req.Modify.FileSize != nil {
		update["file_size"] = *req.Modify.FileSize
	}
	if req.Modify.Hash != nil {
		update["hash"] = *req.Modify.Hash
	}
	if req.Modify.Platform != nil {
		update["platform"] = *req.Modify.Platform
	}
	if req.Modify.DownKey != nil {
		update["down_key"] = *req.Modify.DownKey
	}
	sql, args, err := builder.BuildUpdate(d.Table(), where, update)
	if err != nil {
		return nil, errs.Wrap(constants.ErrParam, "build update", err)
	}
	if _, err = d.Client().ExecContext(ctx, sql, args...); err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "exec update", err)
	}
	return &model.ModifyGameResponse{}, nil
}

func (d *gameinfoImpl) DeleteGame(ctx context.Context, req *model.DeleteGameRequest) (*model.DeleteGameResponse, error) {
	if req.Query == nil {
		return nil, errs.New(constants.ErrParam, "nil query")
	}
	where := map[string]interface{}{}
	if req.Query.ID != nil {
		where["id"] = *req.Query.ID
	}
	sql, args, err := builder.BuildDelete(d.Table(), where)
	if err != nil {
		return nil, errs.Wrap(constants.ErrParam, "build delete", err)
	}
	_, err = d.Client().ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "exec delete", err)
	}
	return &model.DeleteGameResponse{}, nil
}
