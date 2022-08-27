package main

import (
	"flag"
	"fmt"
	"log"
	"retromanager/cmd/tools/gamelist"
)

var dir = flag.String("dir", "", "rom dir")
var system = flag.Int("system", 0, "system type")
var apiSvr = flag.String("apisvr", "http://127.0.0.1:9900", "api server addr")

func main() {
	flag.Parse()

	gamelistFile := fmt.Sprintf("%s/gamelist.xml", *dir)

	gl, err := gamelist.Parse(gamelistFile)
	if err != nil {
		log.Panicf("parse gamelist fail, err:%v", err)
	}
	log.Printf("parse finish, games:%d, folder:%d", len(gl.Games), len(gl.Folders))
	for _, game := range gl.Games {
		log.Printf("read game:%+v", *game)
	}
	for _, folder := range gl.Folders {
		log.Printf("read folder:%+v", *folder)
	}
	if err := gamelist.Validate(*dir, gl); err != nil {
		log.Panicf("game validate fail, err:%v", err)
	}
}
