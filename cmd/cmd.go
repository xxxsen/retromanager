package main

import (
	"context"
	"flag"
	"retromanager/action"
	"retromanager/config"
	"retromanager/cron"
	"retromanager/dao"
	"retromanager/db"
	"retromanager/esservice"
	"retromanager/handler"
	"time"

	"github.com/xxxsen/common/es"
	"github.com/xxxsen/common/idgen"
	"github.com/xxxsen/common/logger"
	"github.com/xxxsen/common/naivesvr"
	"go.uber.org/zap"
)

var file = flag.String("config", "./config.json", "config file path")

func main() {
	flag.Parse()

	c, err := config.Parse(*file)
	if err != nil {
		panic(err)
	}
	logitem := c.LogInfo
	logger := logger.Init(logitem.File, logitem.Level, int(logitem.FileCount), int(logitem.FileSize), int(logitem.KeepDays), logitem.Console)

	logger.Info("recv config", zap.Any("config", c))

	if err := db.InitGameDB(&c.GameDBInfo); err != nil {
		logger.With(zap.Error(err)).Fatal("init game db fail")
	}
	dao.GameInfoDao.Watch(action.NewDB2ESAction())
	if err := idgen.Init(c.IDGenInfo.WorkerID); err != nil {
		logger.With(zap.Error(err)).Fatal("init idgen fail")
	}
	if err := es.Init(
		es.WithAuth(c.EsInfo.User, c.EsInfo.Password),
		es.WithHost(c.EsInfo.Host...),
		es.WithTimeout(time.Duration(c.EsInfo.Timeout)*time.Millisecond),
	); err != nil {
		logger.With(zap.Error(err)).Fatal("init es fail")
	}
	if err := esservice.TryCreateIndex(
		context.Background(),
		es.Client,
		dao.GameInfoDao.Table(),
	); err != nil {
		logger.With(zap.Error(err)).Fatal("create index fail")
	}

	//start cronjob
	cron.Start()

	svr, err := naivesvr.NewServer(
		naivesvr.WithAddress(c.ServerInfo.Address),
		naivesvr.WithHandlerRegister(handler.OnRegist),
	)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("init server fail")
	}
	if err := svr.Run(); err != nil {
		logger.With(zap.Error(err)).Fatal("run server fail")
	}
}
