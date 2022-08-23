package server

import (
	"context"
	"retromanager/constants"
	"retromanager/errs"
	"retromanager/server/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type server struct {
	c *Config
}

func NewServer(opts ...Option) (*server, error) {
	c := &Config{
		attach: make(map[string]interface{}),
	}
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
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	s.registDefault(engine)
	s.c.registerFn(engine)
	if err := engine.Run(s.c.addresses...); err != nil {
		return err
	}
	return nil
}

func (s *server) registDefault(engine *gin.Engine) {
	engine.Use(
		PanicRecoverMiddleware(s),
		SupportAttachMiddleware(s),
		EnableServerTrace(s),
	)
}

func GetLogger(ctx context.Context, name string) *zap.Logger {
	return log.GetLogger(ctx).With(zap.String("name", name))
}
