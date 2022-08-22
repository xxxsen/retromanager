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
	fileDBFields = []string{
		"id", "file_name", "hash", "file_size", "create_time", "down_key",
	}
)

var FileInfoDao FileInfoService = &fileInfoDaoImpl{}

type FileInfoService interface {
	GetFile(ctx context.Context, req *model.GetFileRequest) (*model.GetFileResponse, bool, error)
	CreateFile(ctx context.Context, req *model.CreateFileRequest) (*model.CreateFileResponse, error)
}

type fileInfoDaoImpl struct {
}

func (d *fileInfoDaoImpl) Table() string {
	return "file_info_tab"
}

func (d *fileInfoDaoImpl) Client() *sql.DB {
	return db.GetMediaDB()
}

func (d *fileInfoDaoImpl) Fields() []string {
	return fileDBFields
}

func (d *fileInfoDaoImpl) GetFile(ctx context.Context, req *model.GetFileRequest) (*model.GetFileResponse, bool, error) {
	where := map[string]interface{}{
		"down_key": req.DownKey,
		"_limit":   []uint{0, 1},
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
	var item *model.FileItem
	for rows.Next() {
		item = &model.FileItem{}
		if err := rows.Scan(&item.Id, &item.FileName,
			&item.Hash, &item.FileSize, &item.CreateTime,
			&item.DownKey); err != nil {

			return nil, false, errs.Wrap(constants.ErrDatabase, "scan fail", err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, errs.Wrap(constants.ErrDatabase, "scan fail", err)
	}
	if item == nil {
		return nil, false, nil
	}
	return &model.GetFileResponse{Item: item}, true, nil
}

func (d *fileInfoDaoImpl) CreateFile(ctx context.Context, req *model.CreateFileRequest) (*model.CreateFileResponse, error) {
	data := []map[string]interface{}{
		{
			"file_name":   req.Item.FileName,
			"hash":        req.Item.Hash,
			"file_size":   req.Item.FileSize,
			"create_time": req.Item.CreateTime,
			"down_key":    req.Item.DownKey,
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
	return &model.CreateFileResponse{}, nil
}
