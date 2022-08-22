package s3

import (
	"bytes"
	"context"
	"io"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *s3Client

type s3Client struct {
	c      *config
	client *minio.Client
	core   *minio.Core
}

func InitGlobal(opts ...Option) error {
	client, err := New(opts...)
	if err != nil {
		return errs.Wrap(constants.ErrS3, "init s3", err)
	}
	Client = client
	return nil
}

func New(opts ...Option) (*s3Client, error) {
	c := &config{
		ssl: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.bucket) == 0 {
		return nil, errs.New(constants.ErrParam, "nil bucket name")
	}
	client, err := minio.New(c.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.secretId, c.secretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "init client fail", err)
	}

	core, err := minio.NewCore(c.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.secretId, c.secretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "init core fail", err)
	}
	return &s3Client{c: c, client: client, core: core}, nil
}

func (c *s3Client) Download(ctx context.Context, fileid string) (io.ReadCloser, error) {
	reader, err := c.client.GetObject(ctx, c.c.bucket, fileid, minio.GetObjectOptions{})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "get obj fail", err)
	}
	return reader, nil
}

func (c *s3Client) Upload(ctx context.Context, fileid string, r io.Reader, sz int64) error {
	_, err := c.client.PutObject(ctx, c.c.bucket, fileid, r, sz, minio.PutObjectOptions{})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "write obj fail", err)
	}
	return nil
}

func (c *s3Client) Remove(ctx context.Context, fileid string) error {
	return c.client.RemoveObject(ctx, c.c.bucket, fileid, minio.RemoveObjectOptions{})
}

func (c *s3Client) BeginUpload(ctx context.Context, fileid string) (string, error) {
	uploadid, err := c.core.NewMultipartUpload(ctx, c.c.bucket, fileid, minio.PutObjectOptions{})
	if err != nil {
		return "", errs.Wrap(constants.ErrS3, "create multi part upload fail", err)
	}
	return uploadid, nil
}

func (c *s3Client) UploadPart(ctx context.Context, fileid string, uploadid string, partid int, data []byte) (*minio.ObjectPart, error) {
	md5base64 := utils.CalcMd5Base64(data)
	sha256hex := utils.CalcSha256Hex(data)
	op, err := c.core.PutObjectPart(ctx, c.c.bucket, fileid, uploadid, partid, bytes.NewReader(data), int64(len(data)), md5base64, sha256hex, nil)
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "put part fail", err)
	}
	return &op, nil
}

func (c *s3Client) EndUpload(ctx context.Context, fileid string, uploadid string, parts []minio.CompletePart) error {
	_, err := c.core.CompleteMultipartUpload(ctx, c.c.bucket, fileid, uploadid, parts, minio.PutObjectOptions{})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "finish upload fail", err)
	}
	return nil
}

func (c *s3Client) DiscardMultiPartUpload(ctx context.Context, fileid string, uploadid string) error {
	if err := c.core.AbortMultipartUpload(ctx, c.c.bucket, fileid, uploadid); err != nil {
		return errs.Wrap(constants.ErrS3, "abort multipart upload fail", err)
	}
	return nil
}
