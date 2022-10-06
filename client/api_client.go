package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"retromanager/proto/retromanager/gameinfo"
	"retromanager/utils"
	"strings"
	"time"

	"github.com/xxxsen/common/errs"
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
	if len(c.filesvr) == 0 || len(c.apisvr) == 0 {
		return nil, errs.New(errs.ErrParam, "filesvr/apisvr not found")
	}
	if strings.HasSuffix(c.filesvr, "/") || strings.HasSuffix(c.apisvr, "/") {
		return nil, errs.New(errs.ErrParam, "host should not end with '/'")
	}
	client := &http.Client{}

	return &Client{c: c, client: client}, nil
}

func (c *Client) buildFileSvrAPI(api string) string {
	return c.c.filesvr + api
}

func (c *Client) buildAPISvrAPI(api string) string {
	return c.c.apisvr + api
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
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "copy stream fail", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "close writer fail", err)
	}

	req, err := http.NewRequest("POST", c.buildFileSvrAPI(api), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ts := time.Now().Add(1 * time.Minute).Unix()
	utils.CreateCodeAuthRequest(req, c.c.ak, c.c.sk, uint64(ts))
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
	req, err := c.createMultiPartRequest(api, map[string]string{"md5": md5v}, "file", path.Base(file), f)
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
	req, err := http.NewRequest(http.MethodPost, c.buildAPISvrAPI(api), bytes.NewReader(raw))
	if err != nil {
		return nil, errs.Wrap(errs.ErrServiceInternal, "make request fail", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	st, err := os.Stat(req.File)
	if err != nil {
		return nil, errs.Wrap(errs.ErrIO, "stat file fail", err)
	}
	caller := c.uploadSmallFile
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
