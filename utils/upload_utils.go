package utils

import (
	"encoding/base64"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/proto/retromanager/gameinfo"

	"google.golang.org/protobuf/proto"
)

func EncodeUploadID(upload *gameinfo.UploadIdCtx) (string, error) {
	return encodeMessage(upload)
}

func DecodeUploadID(id string) (*gameinfo.UploadIdCtx, error) {
	ctx := &gameinfo.UploadIdCtx{}
	if err := decodeMessage(id, ctx); err != nil {
		return nil, err
	}
	return ctx, nil
}

func EncodePartID(part *gameinfo.PartIdCtx) (string, error) {
	return encodeMessage(part)
}

func DecodePartID(id string) (*gameinfo.PartIdCtx, error) {
	part := &gameinfo.PartIdCtx{}
	if err := decodeMessage(id, part); err != nil {
		return nil, err
	}
	return part, nil
}

func encodeMessage(msg proto.Message) (string, error) {
	raw, err := proto.Marshal(msg)
	if err != nil {
		return "", errs.Wrap(constants.ErrMarshal, "pb marshal fail", err)
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

func decodeMessage(id string, dst proto.Message) error {
	raw, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return errs.Wrap(constants.ErrUnmarshal, "base64 decode fail", err)
	}
	if err := proto.Unmarshal(raw, dst); err != nil {
		return errs.Wrap(constants.ErrUnmarshal, "proto decode fail", err)
	}
	return nil
}

func CalcFileBlockCount(sz uint64, blksz uint64) int {
	return int((sz + blksz - 1) / blksz)
}
