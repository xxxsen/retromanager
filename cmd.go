package main

import (
	"flag"
	"retromanager/config"
	"retromanager/db"
	"retromanager/handler"
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

	if err := db.InitGameDB(&c.DBInfo); err != nil {
		log.Fatalf("init db fail, err:%v", err)
	}
	//TODO: init es
	svr, err := server.NewServer(
		server.WithAddress(c.ServerInfo.Address),
		server.WithHandlerRegister(handler.OnRegist),
	)
	if err != nil {
		log.Fatalf("init server fail, err:%v", err)
	}
	if err := svr.Run(); err != nil {
		log.Fatalf("run server fail, err:%w", err)
	}
}
