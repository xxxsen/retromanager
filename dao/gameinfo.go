package dao

import (
	"context"
	"database/sql"
	"fmt"
	"retromanager/db"
	"retromanager/model"
	"time"

	"github.com/xxxsen/errs"

	"github.com/didi/gendry/builder"
	"google.golang.org/protobuf/proto"
)

var gameinfoFields = []string{
	"id", "platform", "display_name", "file_size", "detail", "create_time", "update_time", "hash", "extinfo", "down_key",
}

var GameInfoDao GameInfoService = NewGameInfoDao()

type GameInfoService interface {
	IWatcher
	Table() string
	GetGame(ctx context.Context, req *model.GetGameRequest) (*model.GetGameResponse, bool, error)
	ListGame(ctx context.Context, req *model.ListGameRequest) (*model.ListGameResponse, error)
	CreateGame(ctx context.Context, req *model.CreateGameRequest) (*model.CreateGameResponse, error)
	ModifyGame(ctx context.Context, req *model.ModifyGameRequest) (*model.ModifyGameResponse, error)
	DeleteGame(ctx context.Context, req *model.DeleteGameRequest) (*model.DeleteGameResponse, error)
}

type gameinfoImpl struct {
	notifyers []IDataNotifyer
}

func NewGameInfoDao() *gameinfoImpl {
	return &gameinfoImpl{}
}

func (d *gameinfoImpl) Watch(notifyer IDataNotifyer) {
	d.notifyers = append(d.notifyers, notifyer)
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
		return 0, errs.Wrap(errs.ErrParam, "build", err)
	}
	rows, err := d.Client().QueryContext(ctx, sql, args...)
	if err != nil {
		return 0, errs.Wrap(errs.ErrDatabase, "query total", err)
	}
	defer rows.Close()
	var total uint32
	for rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, errs.Wrap(errs.ErrDatabase, "scan total", err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, errs.Wrap(errs.ErrDatabase, "scan total", err)
	}
	return total, nil
}

func (d *gameinfoImpl) GetGame(ctx context.Context, req *model.GetGameRequest) (*model.GetGameResponse, bool, error) {
	subReq := &model.ListGameRequest{
		Query:     &model.ListQuery{ID: &req.GameId, State: proto.Uint32(model.GameStateNormal)},
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
		if req.Query.State != nil {
			where["state"] = *req.Query.State
		}
		if len(req.Query.UpdateTime) > 0 {
			where["update_time >="] = req.Query.UpdateTime[0]
			if len(req.Query.UpdateTime) > 1 {
				where["update_time <="] = req.Query.UpdateTime[1]
			}
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
		return nil, errs.Wrap(errs.ErrParam, "build select", err)
	}
	client := d.Client()
	rows, err := client.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(errs.ErrDatabase, "query ctx fail", err)
	}
	defer rows.Close()
	lst := make([]*model.GameItem, 0, req.Limit)
	for rows.Next() {
		item := &model.GameItem{}
		if err := rows.Scan(&item.ID, &item.Platform, &item.DisplayName, &item.FileSize, &item.Desc, &item.CreateTime,
			&item.UpdateTime, &item.Hash, &item.ExtInfo, &item.DownKey); err != nil {

			return nil, errs.Wrap(errs.ErrDatabase, "scan", err)
		}
		lst = append(lst, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.Wrap(errs.ErrDatabase, "scan", err)
	}
	var total uint32
	if req.NeedTotal {
		total, err = d.buildTotal(ctx, where)
		if err != nil {
			return nil, errs.Wrap(errs.ErrDatabase, "build total", err)
		}
	}
	return &model.ListGameResponse{List: lst, Total: total}, nil
}

func (d *gameinfoImpl) CreateGame(ctx context.Context, req *model.CreateGameRequest) (*model.CreateGameResponse, error) {
	item := req.Item
	data := []map[string]interface{}{
		{
			"platform":     item.Platform,
			"display_name": item.DisplayName,
			"file_size":    item.FileSize,
			"detail":       item.Desc,
			"create_time":  item.CreateTime,
			"update_time":  item.UpdateTime,
			"hash":         item.Hash,
			"extinfo":      item.ExtInfo,
			"down_key":     item.DownKey,
			"state":        model.GameStateNormal,
		},
	}
	sql, args, err := builder.BuildInsert(d.Table(), data)
	if err != nil {
		return nil, errs.Wrap(errs.ErrParam, "build insert", err)
	}
	rs, err := d.Client().ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(errs.ErrDatabase, "exec insert", err)
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return nil, errs.Wrap(errs.ErrDatabase, "get insert id", err)
	}
	AsyncNotify(ctx, d.Table(), model.ActionCreate, uint64(id), d.notifyers...)
	return &model.CreateGameResponse{GameId: uint64(id)}, nil
}

func (d *gameinfoImpl) ModifyGame(ctx context.Context, req *model.ModifyGameRequest) (*model.ModifyGameResponse, error) {
	if req.Modify == nil {
		return nil, errs.New(errs.ErrParam, "nil modify")
	}
	where := map[string]interface{}{
		"id": req.GameID,
	}
	if req.State != nil {
		where["state"] = *req.State
	}
	update := map[string]interface{}{
		"update_time": time.Now().UnixNano() / int64(time.Millisecond),
	}
	if req.Modify.Desc != nil {
		update["detail"] = *req.Modify.Desc
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
	if req.Modify.State != nil {
		update["state"] = *req.Modify.State
	}
	sql, args, err := builder.BuildUpdate(d.Table(), where, update)
	if err != nil {
		return nil, errs.Wrap(errs.ErrParam, "build update", err)
	}
	rs, err := d.Client().ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(errs.ErrDatabase, "exec update", err)
	}
	cnt, err := rs.RowsAffected()
	if err != nil {
		return nil, errs.Wrap(errs.ErrDatabase, "get affect rows fail", err)
	}
	AsyncNotify(ctx, d.Table(), model.ActionModify, req.GameID, d.notifyers...)
	return &model.ModifyGameResponse{AffectRows: cnt}, nil
}

func (d *gameinfoImpl) DeleteGame(ctx context.Context, req *model.DeleteGameRequest) (*model.DeleteGameResponse, error) {
	daoReq := &model.ModifyGameRequest{
		GameID: req.GameID,
		State:  proto.Uint32(model.GameStateNormal),
		Modify: &model.ModifyInfo{
			State: proto.Uint32(model.GameStateDelete),
		},
	}
	daoRsp, err := d.ModifyGame(ctx, daoReq)
	if err != nil {
		return nil, errs.Wrap(errs.ErrDatabase, "exec delete", err)
	}
	return &model.DeleteGameResponse{AffectRows: daoRsp.AffectRows}, nil
}
