package main

//go:generate go run -tags=dev generate_assets.go

import (
	"flag"
	"fmt"
	"os"

	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/server"
	"github.com/martinv13/go-shiny/services/appproxy"
)

func main() {
	environment := flag.String("e", "development", "")
	flag.Usage = func() {
		fmt.Println("Usage: server -e {mode}")
		os.Exit(1)
	}
	flag.Parse()
	fmt.Println(*environment)
	models.InitDB()
	err := appproxy.StartApps()
	if err != nil {
		fmt.Println(err)
	}
	server.Init()
}
