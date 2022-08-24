package main

import (
	"flag"
	"retromanager/action"
	"retromanager/config"
	"retromanager/constants"
	"retromanager/cron"
	"retromanager/dao"
	"retromanager/db"
	"retromanager/handler"
	hconfig "retromanager/handler/config"
	"retromanager/idgen"
	"retromanager/s3"
	"retromanager/server"

	"retromanager/server/log"

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
	logger := log.Init(logitem.File, logitem.Level, int(logitem.FileCount), int(logitem.FileSize), int(logitem.KeepDays), logitem.Console)

	logger.Info("recv config", zap.Any("config", c))

	if err := db.InitGameDB(&c.GameDBInfo); err != nil {
		logger.With(zap.Error(err)).Fatal("init game db fail")
	}
	dao.GameInfoDao.Watch(action.NewDB2ESAction())
	if err := db.InitFileDB(&c.FileDBInfo); err != nil {
		logger.With(zap.Error(err)).Fatal("init media db fail")
	}
	if err := idgen.Init(c.IDGenInfo.WorkerID); err != nil {
		logger.With(zap.Error(err)).Fatal("init idgen fail")
	}
	if err := s3.InitGlobal(
		s3.WithEndpoint(c.S3Info.Endpoint),
		s3.WithSSL(c.S3Info.UseSSL),
		s3.WithSecret(c.S3Info.SecretId, c.S3Info.SecretKey),
		s3.WithBucket(c.S3Info.Bucket),
	); err != nil {
		logger.With(zap.Error(err)).Fatal("init s3 fail")
	}

	//start cronjob
	cron.Start()

	//TODO: init es
	svr, err := server.NewServer(
		server.WithAddress(c.ServerInfo.Address),
		server.WithHandlerRegister(handler.OnRegist),
		server.WithAttach(constants.KeyConfigAttach, initServiceConfig(c)),
	)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("init server fail")
	}
	if err := svr.Run(); err != nil {
		logger.With(zap.Error(err)).Fatal("run server fail")
	}
}

func initServiceConfig(c *config.Config) *hconfig.Config {
	return &hconfig.Config{}
}
