package main

import (
	"context"
	"flag"
	"log"
	"retromanager/cmd/tools/import_from_dir/importer"
)

var dir = flag.String("dir", "", "rom dir")
var system = flag.Int("system", 0, "system type")
var apiSvr = flag.String("apisvr", "http://127.0.0.1:9900", "api server addr")
var fileSvr = flag.String("filesvr", "http://127.0.0.1:9901", "file svr addr")
var cleanBeforeValidate = flag.Bool("clean_before_validate", false, "clean before validate")
var checkOnly = flag.Bool("check_only", true, "check only")
var ak = flag.String("ak", "abc", "access key")
var sk = flag.String("sk", "123456", "secret key")

func main() {
	flag.Parse()

	imp, err := importer.New(
		importer.WithAPISvr(*apiSvr),
		importer.WithDir(*dir),
		importer.WithSystem(*system),
		importer.WithFileSvr(*fileSvr),
		importer.WithSecret(*ak, *sk),
	)
	if err != nil {
		panic(err)
	}
	log.Printf("load gamedata finish, folder:%d, game:%d", imp.GetGameList().FolderSize(), imp.GetGameList().GameSize())
	if *cleanBeforeValidate {
		if err := imp.Clean(); err != nil {
			log.Panicf("clean gamelist fail, err:%v", err)
		}
	}
	log.Printf("load gamedata after clean finish, folder:%d, game:%d", imp.GetGameList().FolderSize(), imp.GetGameList().GameSize())
	if *checkOnly {
		log.Printf("in check mode, skip next")
		return
	}

	if err := imp.Validate(); err != nil {
		panic(err)
	}
	if err := imp.DoImport(context.Background()); err != nil {
		panic(err)
	}
	log.Printf("do import finish")
}
