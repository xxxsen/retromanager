package file

import (
	"context"
	"fmt"
	"net/http"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"

	"github.com/gin-gonic/gin"
	"github.com/xxxsen/runner"
	"google.golang.org/protobuf/proto"
)

const (
	maxListMetaSizePerRequest = 20
)

func Meta(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.GetFileMetaRequest)
	if len(req.DownKey) == 0 || len(req.DownKey) > maxListMetaSizePerRequest {
		return http.StatusOK, errs.New(constants.ErrParam, "invalid down key size:%d", len(req.DownKey)), nil
	}
	daoRsp, err := dao.FileInfoDao.ListFile(ctx, &model.ListFileRequest{
		Query: &model.ListFileQuery{
			DownKey: req.GetDownKey(),
		},
		Offset: 0,
		Limit:  uint32(len(req.DownKey)),
	})
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "read file list fail", err), nil
	}
	metalist := fileinfo2pbmeta(req.GetDownKey(), daoRsp.List)
	_ = buildS3Info(ctx, metalist)
	return http.StatusOK, nil, &gameinfo.GetFileMetaResponse{
		List: metalist,
	}
}

func buildS3Info(ctx context.Context, metalist []*gameinfo.FileItem) error {
	run := runner.New(10)
	for _, meta := range metalist {
		meta := meta
		if !meta.GetExist() {
			continue
		}
		run.Add(fmt.Sprintf("fetch_%s", meta.GetDownKey()), func(ctx context.Context) error {
			info, err := s3.Client.GetFileInfo(ctx, meta.GetDownKey())
			if err != nil {
				return errs.Wrap(constants.ErrS3, "read key fail", err).WithDebugMsg("key:%s", meta.GetDownKey())
			}
			meta.LowLevelInfo = &gameinfo.FileMetaLowLevelInfo{
				Etag: info.ETag,
			}
			return nil
		})
	}
	return run.Run(ctx)
}

func fileinfo2pbmeta(order []string, lst []*model.FileItem) []*gameinfo.FileItem {
	mapper := make(map[string]*model.FileItem)
	for _, item := range lst {
		mapper[item.DownKey] = item
	}
	rs := make([]*gameinfo.FileItem, 0, len(lst))
	for _, key := range order {
		src := mapper[key]
		dst := &gameinfo.FileItem{
			DownKey: proto.String(key),
			Exist:   proto.Bool(false),
		}
		if src != nil {
			dst.FileName = proto.String(src.FileName)
			dst.Hash = proto.String(src.Hash)
			dst.FileSize = proto.Uint64(src.FileSize)
			dst.CreateTime = proto.Uint64(src.CreateTime)
			dst.DownKey = proto.String(src.DownKey)
			dst.Exist = proto.Bool(true)
		}
		rs = append(rs, dst)
	}
	return rs
}
