package s3

import (
	"context"
	"io"
	"net/http"
	"retromanager/constants"
	"retromanager/errs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"google.golang.org/protobuf/proto"
)

var Client *s3Client

type s3Client struct {
	c      *config
	sess   *session.Session
	client *s3.S3
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

	credit := credentials.NewStaticCredentials(c.secretId, c.secretKey, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: credit,
		Endpoint:    aws.String(c.endpoint),
		DisableSSL:  aws.Bool(!c.ssl),
		HTTPClient:  &http.Client{},
		Region:      proto.String("cn"),
	})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "init session fail", err)
	}
	client := s3.New(sess)
	return &s3Client{c: c, client: client, sess: sess}, nil
}

func (c *s3Client) Download(ctx context.Context, fileid string) (io.ReadCloser, error) {
	output, err := c.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.c.bucket),
		Key:    aws.String(fileid),
	})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "get obj fail", err)
	}
	return output.Body, nil
}

func (c *s3Client) Upload(ctx context.Context, fileid string, r io.ReadSeeker, sz int64) error {
	_, err := c.client.PutObject(&s3.PutObjectInput{
		Body:   r,
		Bucket: aws.String(c.c.bucket),
		Key:    aws.String(fileid),
	})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "write obj fail", err)
	}
	return nil
	// uploader := s3manager.NewUploader(c.sess, func(u *s3manager.Uploader) {
	// 	u.PartSize = 2 * 1024 * 1024 // The minimum/default allowed part size is 5MB
	// 	u.Concurrency = 5            // default is 5
	// })
	// _, err := uploader.Upload(&s3manager.UploadInput{
	// 	Bucket: aws.String(c.c.bucket),
	// 	Key:    aws.String(fileid),
	// 	Body:   r,
	// })
	// if err != nil {
	// 	return errs.Wrap(constants.ErrS3, "write obj fail", err)
	// }
	// return nil
}

func (c *s3Client) Remove(ctx context.Context, fileid string) error {
	_, err := c.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(c.c.bucket),
		Key:    aws.String(fileid),
	})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "delete fail", err)
	}
	return nil
}

func (c *s3Client) BeginUpload(ctx context.Context, fileid string) (string, error) {
	output, err := c.client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(c.c.bucket),
		Key:    aws.String(fileid),
	})
	if err != nil {
		return "", errs.Wrap(constants.ErrS3, "create multi part upload fail", err)
	}
	return *output.UploadId, nil
}

func (c *s3Client) UploadPart(ctx context.Context, fileid string, uploadid string, partid int, file io.ReadSeeker) error {
	_, err := c.client.UploadPart(&s3.UploadPartInput{
		Body:       file,
		Bucket:     aws.String(c.c.bucket),
		Key:        aws.String(fileid),
		PartNumber: aws.Int64(int64(partid)),
		UploadId:   aws.String(uploadid),
	})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "put part fail", err)
	}
	return nil
}

func (c *s3Client) listParts(ctx context.Context, fileid string, uploadid string) ([]*s3.Part, error) {
	output, err := c.client.ListParts(&s3.ListPartsInput{
		Bucket:              aws.String(c.c.bucket),
		ExpectedBucketOwner: new(string),
		Key:                 aws.String(fileid),
		UploadId:            aws.String(uploadid),
	})
	if err != nil {
		return nil, errs.Wrap(constants.ErrS3, "list part fail", err)
	}
	return output.Parts, nil
}

func (c *s3Client) parts2completeparts(src []*s3.Part) []*s3.CompletedPart {
	out := make([]*s3.CompletedPart, 0, len(src))
	for _, p := range src {
		out = append(out, &s3.CompletedPart{
			ChecksumCRC32:  p.ChecksumCRC32,
			ChecksumCRC32C: p.ChecksumCRC32C,
			ChecksumSHA1:   p.ChecksumSHA1,
			ChecksumSHA256: p.ChecksumSHA256,
			ETag:           p.ETag,
			PartNumber:     p.PartNumber,
		})
	}
	return out
}

func (c *s3Client) EndUpload(ctx context.Context, fileid string, uploadid string, partcount int) error {
	parts, err := c.listParts(ctx, fileid, uploadid)
	if err != nil {
		return err
	}
	if len(parts) != partcount {
		return errs.New(constants.ErrParam, "part count not match, need:%d, get:%d", partcount, len(parts))
	}
	_, err = c.client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket: aws.String(c.c.bucket),
		Key:    aws.String(fileid),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: c.parts2completeparts(parts),
		},
		UploadId: aws.String(uploadid),
	})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "finish upload fail", err)
	}
	return nil
}

func (c *s3Client) DiscardMultiPartUpload(ctx context.Context, fileid string, uploadid string) error {
	_, err := c.client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(c.c.bucket),
		Key:      aws.String(fileid),
		UploadId: aws.String(uploadid),
	})
	if err != nil {
		return errs.Wrap(constants.ErrS3, "abort multipart upload fail", err)
	}
	return nil
}
