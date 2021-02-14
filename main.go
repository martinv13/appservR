package main

//go:generate go run -tags=dev generate_assets.go

import (
	"fmt"

	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/server"
	"github.com/martinv13/go-shiny/services/appproxy"
)

func main() {

	loadConfig()

	db := models.InitDB()

	app := models.ShinyApp{}
	app.Init(db)

	err := appproxy.StartApps()
	if err != nil {
		fmt.Println(err)
	}

	server.Init(db)

}
