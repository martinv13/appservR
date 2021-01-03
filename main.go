package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/martinv13/go-shiny/server"
	"github.com/martinv13/go-shiny/services/appproxy"
	"github.com/martinv13/go-shiny/services/db"
)

func main() {
	environment := flag.String("e", "development", "")
	flag.Usage = func() {
		fmt.Println("Usage: server -e {mode}")
		os.Exit(1)
	}
	flag.Parse()
	fmt.Println(*environment)
	db.Init()
	err := appproxy.StartApps()
	if err != nil {
		fmt.Println(err)
	}
	server.Init()
}
