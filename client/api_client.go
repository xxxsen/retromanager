package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"retromanager/constants"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/utils"
	"strconv"

	"github.com/xxxsen/errs"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	c      *config
	client *http.Client
}

func New(opts ...Option) (*Client, error) {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.host) == 0 {
		return nil, errs.New(errs.ErrParam, "no host found")
	}
	client := &http.Client{}

	return &Client{c: c, client: client}, nil
}

func (c *Client) createMultiPartRequest(api string, kv map[string]string, fparam, fname string, f io.Reader) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range kv {
		_ = writer.WriteField(k, v)
	}
	part, err := writer.CreateFormFile(fparam, fname)
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "create form file fail", err)
	}
	_, err = io.Copy(part, f)
	err = writer.Close()
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "close writer fail", err)
	}

	req, err := http.NewRequest("POST", api, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

type msgFrame struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (c *Client) responseByJson(ctx context.Context, req *http.Request, rsp interface{}) error {
	httpRsp, err := c.client.Do(req)
	if err != nil {
		return errs.Wrap(errs.ErrIO, "do http request fail", err)
	}
	defer httpRsp.Body.Close()
	if httpRsp.StatusCode != http.StatusOK {
		return errs.New(errs.ErrServiceInternal, "http status code:%d not ok", httpRsp.StatusCode)
	}
	raw, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return errs.Wrap(errs.ErrIO, "read body fail", err)
	}
	frame := &msgFrame{
		Data: rsp,
	}
	if err := json.Unmarshal(raw, frame); err != nil {
		return errs.Wrap(errs.ErrUnmarshal, "decode json frame fail", err)
	}
	if frame.Code != 0 {
		return errs.New(int64(frame.Code), frame.Message)
	}
	return nil
}

func (c *Client) UploadVideo(ctx context.Context, req *UploadVideoRequest) (*UploadVideoResponse, error) {
	st, err := os.Stat(req.File)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "stat file fail", err)
	}
	api := apiUploadVideo
	meta, err := c.uploadSmallFileByAPI(ctx, api, req.File, st.Size())
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "upload by api fail", err).WithDebugMsg("api:%s", api)
	}
	return &UploadVideoResponse{
		Meta: meta,
	}, nil
}

func (c *Client) UploadImage(ctx context.Context, req *UploadImageRequest) (*UploadImageResponse, error) {
	st, err := os.Stat(req.File)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "stat file fail", err)
	}
	api := apiUploadImage
	meta, err := c.uploadSmallFileByAPI(ctx, api, req.File, st.Size())
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "upload by api fail", err).WithDebugMsg("api:%s", api)
	}
	return &UploadImageResponse{
		Meta: meta,
	}, nil
}

func (c *Client) uploadSmallFileByAPI(ctx context.Context, api string, file string, sz int64) (*FileMeta, error) {
	md5v, err := utils.CalcMd5File(file)
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "calc md5 fail", err)
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "open file fail", err)
	}
	defer f.Close()
	req, err := c.createMultiPartRequest(apiUploadFile, map[string]string{"md5": md5v}, "file", path.Base(file), f)
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "create multipart req fail", err)
	}
	rsp := &gameinfo.FileUploadResponse{}
	if err := c.responseByJson(ctx, req, rsp); err != nil {
		return nil, err
	}
	return &FileMeta{
		Path:    path.Dir(file),
		Name:    path.Base(file),
		Size:    sz,
		MD5:     md5v,
		DownKey: *rsp.DownKey,
	}, nil
}

func (c *Client) uploadSmallFile(ctx context.Context, file string, sz int64) (*FileMeta, error) {
	return c.uploadSmallFileByAPI(ctx, apiUploadFile, file, sz)
}

func (c *Client) createJsonRequest(api string, body interface{}) (*http.Request, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, errs.Wrap(errs.ErrMarshal, "encode json", err)
	}
	req, err := http.NewRequest(http.MethodPost, api, bytes.NewReader(raw))
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "make request fail", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) uploadBigFileBegin(ctx context.Context, sz int64) (string, int64, error) {
	req := &gameinfo.FileUploadBeginRequest{
		FileSize: proto.Uint64(uint64(sz)),
	}
	httpReq, err := c.createJsonRequest(apiUploadBigFileBegin, req)
	if err != nil {
		return "", 0, errs.Wrap(errs.ErrServiceInternal, "create json request fail", err)
	}
	rsp := &gameinfo.FileUploadBeginResponse{}
	if err := c.responseByJson(ctx, httpReq, rsp); err != nil {
		return "", 0, errs.Wrap(errs.ErrIO, "read response fail", err)
	}
	return rsp.GetUploadCtx(), int64(rsp.GetBlockSize()), nil
}

func (c *Client) uploadBigFileEnd(ctx context.Context, uploadid string, filename string, hash string) (string, error) {
	req := &gameinfo.FileUploadEndRequest{
		UploadCtx: proto.String(uploadid),
		FileName:  proto.String(filename),
		Hash:      proto.String(hash),
	}
	httpReq, err := c.createJsonRequest(apiUploadBigFileEnd, req)
	if err != nil {
		return "", errs.Wrap(errs.ErrServiceInternal, "create json request fail", err)
	}
	rsp := &gameinfo.FileUploadEndResponse{}
	if err := c.responseByJson(ctx, httpReq, rsp); err != nil {
		return "", errs.Wrap(errs.ErrIO, "read response fail", err)
	}
	return rsp.GetDownKey(), nil
}

func (c *Client) uploadBigFilePart(ctx context.Context, uploadid string, partid int, hash string, reader io.Reader) error {
	//file, part_id, md5, upload_ctx
	req, err := c.createMultiPartRequest(apiUploadBigFilePart, map[string]string{
		"upload_ctx": uploadid,
		"md5":        hash,
		"part_id":    strconv.FormatInt(int64(partid), 10),
	}, "file", fmt.Sprintf("part_%d", partid), reader)
	if err != nil {
		return errs.Wrap(errs.ErrServiceInternal, "create request fail", err)
	}
	rsp := &gameinfo.FileUploadPartResponse{}
	if err := c.responseByJson(ctx, req, rsp); err != nil {
		return errs.Wrap(errs.ErrIO, "read response fail", err)
	}
	return nil
}

func (c *Client) seekMd5(f io.ReadSeeker, sz int64) (string, error) {
	loc, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", err
	}
	reader := io.LimitReader(f, sz)
	md5v, err := utils.CalcMd5Reader(reader)
	if err != nil {
		return "", err
	}
	if _, err := f.Seek(loc, io.SeekStart); err != nil {
		return "", err
	}
	return md5v, nil
}

func (c *Client) uploadBigFile(ctx context.Context, file string, sz int64) (*FileMeta, error) {
	md5v, err := utils.CalcMd5File(file)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "create md5 fail", err)
	}
	uploadid, partsize, err := c.uploadBigFileBegin(ctx, sz)
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "upload begin fail", err)
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "open file fail", err)
	}
	defer f.Close()
	for i := 0; i < int(sz); i += int(partsize) {
		partid := int(i + 1)
		partmd5, err := c.seekMd5(f, partsize)
		if err != nil {
			return nil, errs.Wrap(errs.ErrIO, "seek and calc md5 fail", err).
				WithDebugMsg("partid:%d", partid).
				WithDebugMsg("partsz:%d", partsize)
		}
		reader := io.LimitReader(f, partsize)
		if err := c.uploadBigFilePart(ctx, uploadid, partid, partmd5, reader); err != nil {
			return nil, errs.Wrap(errs.ErrIO, "upload part fail", err).WithDebugMsg("partid:%d", partid)
		}
	}
	downkey, err := c.uploadBigFileEnd(ctx, uploadid, path.Base(file), md5v)
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "upload end fail", err)
	}
	return &FileMeta{
		Path:    path.Dir(file),
		Name:    path.Base(file),
		Size:    sz,
		MD5:     md5v,
		DownKey: downkey,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	st, err := os.Stat(req.File)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "stat file fail", err)
	}
	caller := c.uploadSmallFile
	if st.Size() > constants.MaxPostUploadSize {
		caller = c.uploadBigFile
	}
	meta, err := caller(ctx, req.File, st.Size())
	if err != nil {
		return nil, err
	}
	return &UploadFileResponse{Meta: meta}, nil
}

func (c *Client) CreateGame(ctx context.Context, req *CreateGameRequest) (*CreateGameResponse, error) {
	rsp := &CreateGameResponse{}
	httpReq, err := c.createJsonRequest(apiCreateGame, req)
	if err != nil {
		return nil, err
	}
	if err := c.responseByJson(ctx, httpReq, rsp); err != nil {
		return nil, err
	}
	return rsp, nil
}
