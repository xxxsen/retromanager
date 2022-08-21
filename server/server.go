package server

import (
	"retromanager/constants"
	"retromanager/errs"

	"github.com/gin-gonic/gin"
)

type server struct {
	c *Config
}

func NewServer(opts ...Option) (*server, error) {
	c := &Config{}
	for _, opt := range opts {
		opt(c)
	}
	s := &server{c: c}
	if err := s.initServer(); err != nil {
		return nil, errs.Wrap(constants.ErrParam, "init server fail", err)
	}
	return s, nil
}

func (s *server) initServer() error {
	if len(s.c.addresses) == 0 {
		return errs.New(constants.ErrParam, "no bind address found")
	}
	return nil
}

func (s *server) Run() error {
	engine := gin.New()
	s.c.registerFn(engine)
	if err := engine.Run(s.c.addresses...); err != nil {
		return err
	}
	return nil
}
