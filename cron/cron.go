package cron

import (
	"context"

	"log"

	"github.com/robfig/cron/v3"
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
	c.cr.AddJob(expr, c.wrapTask(runner))
}

func (c *crontab) wrapTask(runner ICron) cron.Job {
	return cron.SkipIfStillRunning(cron.DefaultLogger)(cron.FuncJob(func() {
		ctx := context.Background()
		if err := runner.Run(ctx); err != nil {
			log.Printf("run task:%s fail, err:%v", runner.Name(), err)
		}
	}))
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
