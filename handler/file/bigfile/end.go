package bigfile

import (
	"net/http"
	"retromanager/constants"
	"retromanager/dao"
	"retromanager/errs"
	"retromanager/model"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/s3"
	"retromanager/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"google.golang.org/protobuf/proto"
)

func buildPartList(parts []string) ([]minio.CompletePart, error) {
	lst := make([]minio.CompletePart, 0, len(parts))
	for idx, part := range parts {
		partctx, err := utils.DecodePartID(part)
		if err != nil {
			return nil, err
		}
		if partctx.GetIdx() != int32(idx) {
			return nil, errs.New(constants.ErrParam, "idx not match, index:%d, real:%d", idx, partctx.GetIdx())
		}
		lst = append(lst, minio.CompletePart{
			PartNumber: int(partctx.GetIdx()),
			ETag:       partctx.GetEtag(),
		})
	}
	return lst, nil
}

func End(ctx *gin.Context, request interface{}) (int, errs.IError, interface{}) {
	req := request.(*gameinfo.FileUploadEndRequest)
	if len(req.GetHash()) == 0 || len(req.GetFileName()) == 0 ||
		len(req.GetPartCtx()) == 0 || len(req.GetUploadCtx()) == 0 {
		return http.StatusOK, errs.New(constants.ErrParam, "invalid hash/filename/partctx/uploadctx"), nil
	}
	uploadctx, err := utils.DecodeUploadID(req.GetUploadCtx())
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "decode upload ctx fail", err), nil
	}
	maxpartsz := utils.CalcFileBlockCount(uploadctx.GetFileSize(), constants.BlockSize)
	if len(req.GetPartCtx())+1 != maxpartsz {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "part count invalid", err), nil
	}
	parts, err := buildPartList(req.GetPartCtx())
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrParam, "decode part ctx fail", err), nil
	}
	err = s3.Client.EndUpload(ctx, *uploadctx.DownKey, uploadctx.GetUploadId(), parts)
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrS3, "complete upload fail", err), nil
	}
	_, err = dao.FileInfoDao.CreateFile(ctx, &model.CreateFileRequest{
		Item: &model.FileItem{
			FileName:   req.GetFileName(),
			Hash:       req.GetHash(),
			FileSize:   uploadctx.GetFileSize(),
			CreateTime: uint64(time.Now().UnixMilli()),
			DownKey:    uploadctx.GetDownKey(),
		},
	})
	if err != nil {
		return http.StatusOK, errs.Wrap(constants.ErrDatabase, "write file record fail", err), nil
	}
	return http.StatusOK, nil, &gameinfo.FileUploadEndResponse{
		DownKey: proto.String(uploadctx.GetDownKey()),
	}
}
