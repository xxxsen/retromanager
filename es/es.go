package es

import (
	"context"
	"net/http"
	"retromanager/constants"
	"retromanager/errs"
	"time"

	"github.com/olivere/elastic/v7"
)

var Client *EsClient

const (
	DefaultVersion = "v1"
)

func Init(opts ...Option) error {
	client, err := New(opts...)
	if err != nil {
		return err
	}
	Client = client
	return nil
}

type EsClient struct {
	*elastic.Client
}

func New(opts ...Option) (*EsClient, error) {
	c := &config{
		user:    "elastic",
		timeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.urls) == 0 {
		return nil, errs.New(constants.ErrParam, "no es host found")
	}
	if len(c.user) == 0 || len(c.password) == 0 {
		return nil, errs.New(constants.ErrParam, "should set user/password")
	}
	httpClient := &http.Client{
		Timeout: c.timeout,
	}
	client, err := elastic.NewClient(
		elastic.SetBasicAuth(c.user, c.password),
		elastic.SetURL(c.urls...),
		elastic.SetHttpClient(httpClient),
	)
	if err != nil {
		return nil, errs.Wrap(constants.ErrES, "init es client fail", err)
	}
	for _, host := range c.urls {
		_, _, err := client.Ping(host).Do(context.Background())
		if err != nil {
			return nil, errs.Wrap(constants.ErrES, "ping host fail", err).WithDebugMsg("host:%s", host)
		}
	}
	return &EsClient{Client: client}, nil
}
