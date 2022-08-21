package dao

import (
	"context"
	"database/sql"
	"retromanager/constants"
	"retromanager/db"
	"retromanager/errs"
	"retromanager/model"

	"github.com/didi/gendry/builder"
)

var (
	mediaDBFields = []string{
		"id", "file_name", "hash", "file_size", "create_time", "file_type",
	}
)

var MediaInfoDao = &mediaInfoDaoImpl{}

type MediaInfoSerice interface {
	GetMedia(ctx context.Context, req *model.GetMediaRequest) (*model.GetMediaResponse, bool, error)
	CreateMedia(ctx context.Context, req *model.CreateMediaRequest) (*model.CreateMediaResponse, error)
}

type mediaInfoDaoImpl struct {
}

func (d *mediaInfoDaoImpl) Table() string {
	return "media_info_tab"
}

func (d *mediaInfoDaoImpl) Client() *sql.DB {
	return db.GetMediaDB()
}

func (d *mediaInfoDaoImpl) Fields() []string {
	return mediaDBFields
}

func (d *mediaInfoDaoImpl) GetMedia(ctx context.Context, req *model.GetMediaRequest) (*model.GetMediaResponse, bool, error) {
	where := map[string]interface{}{
		"hash":      req.Hash,
		"file_type": req.FileType,
		"_limit":    []uint{0, 1},
	}
	sql, args, err := builder.BuildSelect(d.Table(), where, d.Fields())
	if err != nil {
		return nil, false, errs.Wrap(constants.ErrParam, "build select", err)
	}
	rows, err := d.Client().QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, false, errs.Wrap(constants.ErrDatabase, "select fail", err)
	}
	defer rows.Close()
	var item *model.MediaItem
	for rows.Next() {
		item = &model.MediaItem{}
		if err := rows.Scan(&item.Id, &item.FileName,
			&item.Hash, &item.FileSize, &item.CreateTime,
			&item.FileType); err != nil {

			return nil, false, errs.Wrap(constants.ErrDatabase, "scan fail", err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, errs.Wrap(constants.ErrDatabase, "scan fail", err)
	}
	if item == nil {
		return nil, false, nil
	}
	return &model.GetMediaResponse{Item: item}, true, nil
}

func (d *mediaInfoDaoImpl) CreateMedia(ctx context.Context, req *model.CreateMediaRequest) (*model.CreateMediaResponse, error) {
	data := []map[string]interface{}{
		{
			"file_name":   req.Item.FileName,
			"hash":        req.Item.Hash,
			"file_size":   req.Item.FileSize,
			"create_time": req.Item.CreateTime,
			"file_type":   req.Item.FileType,
		},
	}
	sql, args, err := builder.BuildInsertIgnore(d.Table(), data)
	if err != nil {
		return nil, errs.Wrap(constants.ErrParam, "build insert", err)
	}
	_, err = d.Client().ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, errs.Wrap(constants.ErrDatabase, "insert fail", err)
	}
	return &model.CreateMediaResponse{}, nil
}
