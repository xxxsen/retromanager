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

func main() {
	flag.Parse()

	imp, err := importer.New(
		importer.WithAPISvr(*apiSvr),
		importer.WithDir(*dir),
		importer.WithSystem(*system),
	)
	if err != nil {
		panic(err)
	}
	if err := imp.Validate(); err != nil {
		panic(err)
	}
	if err := imp.DoImport(context.Background()); err != nil {
		panic(err)
	}
	log.Printf("do import finish")
}
