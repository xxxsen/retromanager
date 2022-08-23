package cron

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/xxxsen/log"
)

var Default = New()

type ICron interface {
	Name() string
	Run(ctx context.Context) error
}

type crontab struct {
	cr *cron.Cron
}

func New() *crontab {
	cr := cron.New()

	return &crontab{cr: cr}
}

func (c *crontab) Regist(expr string, runner ICron) {
	c.cr.AddFunc(expr, c.wrapTask(runner))
}

func (c *crontab) wrapTask(runner ICron) func() {
	return func() {
		ctx := context.Background()
		if err := runner.Run(ctx); err != nil {
			log.Errorf("run task:%s fail, err:%v", runner.Name(), err)
		}
	}
}

func (c *crontab) Start() {
	c.cr.Start()
}

func Regist(expr string, runner ICron) {
	Default.Regist(expr, runner)
}

func Start() {
	Default.Start()
}
