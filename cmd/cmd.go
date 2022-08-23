package main

import (
	"flag"
	"retromanager/config"
	"retromanager/constants"
	"retromanager/cron"
	"retromanager/db"
	"retromanager/handler"
	hconfig "retromanager/handler/config"
	"retromanager/idgen"
	"retromanager/s3"
	"retromanager/server"

	"github.com/xxxsen/log"
)

var file = flag.String("config", "./config.json", "config file path")

func main() {
	flag.Parse()

	c, err := config.Parse(*file)
	if err != nil {
		panic(err)
	}
	logitem := c.LogInfo
	log.Init(logitem.File, log.StringToLevel(logitem.Level), int(logitem.FileCount), int(logitem.FileSize), int(logitem.KeepDays), logitem.Console)

	log.Infof("recv config:%+v", *c)

	if err := db.InitGameDB(&c.GameDBInfo); err != nil {
		log.Fatalf("init game db fail, err:%v", err)
	}
	if err := db.InitFileDB(&c.FileDBInfo); err != nil {
		log.Fatalf("init media db fail, err:%v", err)
	}
	if err := idgen.Init(c.IDGenInfo.WorkerID); err != nil {
		log.Fatalf("init idgen fail, err:%v", err)
	}
	if err := s3.InitGlobal(
		s3.WithEndpoint(c.S3Info.Endpoint),
		s3.WithSSL(c.S3Info.UseSSL),
		s3.WithSecret(c.S3Info.SecretId, c.S3Info.SecretKey),
		s3.WithBucket(c.S3Info.Bucket),
	); err != nil {
		log.Fatalf("init s3 fail, err:%v", err)
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
		log.Fatalf("init server fail, err:%v", err)
	}
	if err := svr.Run(); err != nil {
		log.Fatalf("run server fail, err:%w", err)
	}
}

func initServiceConfig(c *config.Config) *hconfig.Config {
	return &hconfig.Config{}
}
