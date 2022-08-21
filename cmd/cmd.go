package main

import (
	"flag"
	"retromanager/config"
	"retromanager/constants"
	"retromanager/db"
	"retromanager/handler"
	hconfig "retromanager/handler/config"
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
	if err := db.InitMediaDB(&c.MediaDBInfo); err != nil {
		log.Fatalf("init media db fail, err:%v", err)
	}
	if err := s3.InitGlobal(
		s3.WithEndpoint(c.S3Info.Endpoint),
		s3.WithSSL(c.S3Info.UseSSL),
		s3.WithSecret(c.S3Info.SecretId, c.S3Info.SecretKey),
	); err != nil {
		log.Fatalf("init s3 fail, err:%w", err)
	}

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
	return &hconfig.Config{
		BucketInfo: &hconfig.Bucket{
			RomBucket:   c.S3Info.RomBucket,
			ImageBucket: c.S3Info.ImageBucket,
			VideoBucket: c.S3Info.VideoBucket,
		},
	}
}
