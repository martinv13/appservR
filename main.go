package main

//go:generate go run -tags=dev generate_assets.go

import (
	"fmt"

	"github.com/martinv13/go-shiny/server"
	"github.com/martinv13/go-shiny/services/appproxy"
)

func main() {

	loadConfig()

	err := appproxy.StartApps()
	if err != nil {
		fmt.Println(err)
	}

	server.Init()

}
