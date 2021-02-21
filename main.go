package main

//go:generate go run -tags=dev generate_assets.go

import (
	"log"
	"os"

	"fmt"

	"github.com/kardianos/service"

	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/modules/appproxy"
	"github.com/martinv13/go-shiny/modules/config"
	"github.com/martinv13/go-shiny/server"
)

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {

	config.LoadConfig()

	db := models.InitDB()

	app := models.ShinyApp{}
	app.Init(db)

	err := appproxy.StartApps()
	if err != nil {
		fmt.Println(err)
	}

	server.Init(db)
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "go-shiny-server",
		DisplayName: "Go-Shiny-Server",
		Description: "Serving R Shiny apps",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		var err error
		verb := os.Args[1]
		switch verb {
		case "install":
			err = s.Install()
			if err != nil {
				fmt.Printf("Failed to install: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" installed.\n", svcConfig.DisplayName)
		case "remove":
			err = s.Uninstall()
			if err != nil {
				fmt.Printf("Failed to remove: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" removed.\n", svcConfig.DisplayName)
		case "run":
			err = s.Run()
			if err != nil {
				fmt.Printf("Failed to run: %s\n", err)
			}
		case "start":
			err = s.Start()
			if err != nil {
				fmt.Printf("Failed to start: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" started.\n", svcConfig.DisplayName)
		case "stop":
			err = s.Stop()
			if err != nil {
				fmt.Printf("Failed to stop: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" stopped.\n", svcConfig.DisplayName)
		}
		return
	}

	err = s.Run()

	if err != nil {
		logger.Error(err)
	}
}
