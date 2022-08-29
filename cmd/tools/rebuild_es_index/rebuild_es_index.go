package main

import (
	"context"
	"flag"
	"retromanager/cmd/tools/rebuild_es_index/rebuilder"
	"retromanager/config"
	"retromanager/dao"
	"retromanager/db"
	"retromanager/es"
	"time"

	"github.com/xxxsen/naivesvr/log"
	"go.uber.org/zap"
)

var file = flag.String("config", "./config.json", "config file")

//main read db data and write to es
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
	if err := es.Init(
		es.WithAuth(c.EsInfo.User, c.EsInfo.Password),
		es.WithHost(c.EsInfo.Host...),
		es.WithTimeout(time.Duration(c.EsInfo.Timeout)*time.Millisecond),
	); err != nil {
		logger.With(zap.Error(err)).Fatal("init es fail")
	}
	rb := rebuilder.NewRebuilder(dao.GameInfoDao, es.Client)
	if err := rb.Rebuild(context.Background()); err != nil {
		logger.With(zap.Error(err)).Fatal("rebuild fail")
	}
	logger.Info("rebuild finish")
}
