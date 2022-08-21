package s3

import (
	"context"
	"io"
	"retromanager/constants"
	"retromanager/errs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *s3Client

type s3Client struct {
	c      *config
	client *minio.Client
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
	client, err := minio.New(c.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.secretId, c.secretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "init client fail", err)
	}
	return &s3Client{c: c, client: client}, nil
}

func (c *s3Client) Download(ctx context.Context, bucket string, fileid string) (io.ReadCloser, error) {
	reader, err := c.client.GetObject(ctx, bucket, fileid, minio.GetObjectOptions{})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "get obj fail", err)
	}
	return reader, nil
}

func (c *s3Client) Upload(ctx context.Context, bucket string, fileid string, r io.Reader, sz int64) error {
	_, err := c.client.PutObject(ctx, bucket, fileid, r, sz, minio.PutObjectOptions{})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "write obj fail", err)
	}
	return nil
}

func (c *s3Client) Remove(ctx context.Context, bucket string, fileid string) error {
	return c.client.RemoveObject(ctx, bucket, fileid, minio.RemoveObjectOptions{})
}
