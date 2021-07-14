package main

//go:generate go run -tags=dev generate_assets.go

import (
	"log"

	"fmt"

	"runtime/debug"

	"github.com/kardianos/service"
	"github.com/martinv13/go-shiny/modules/config"
	"github.com/spf13/cobra"
)

var (
	address string
	port    string
	mode    string
	logger  service.Logger
)

func startApp() {
	server, err := InitializeServer(config.RunFlags{Address: address, Mode: mode, Port: port})
	if err != nil {
		panic(err)
	}
	server.Start()
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	startApp()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

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

	cmdRoot := &cobra.Command{
		Use:   "go-shiny",
		Short: "A server for R Shiny apps",
		Long:  `Go-Shiny is a program to deploy R Shiny app on Windows and Linux`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Use \"go-shiny help\" for more information about available commands")
			s.Run()
		},
	}
	cmdRoot.Flags().StringVarP(&address, "address", "a", "", "server hostname or ip adress (default \"localhost\")")
	cmdRoot.Flags().StringVarP(&port, "port", "p", "", "server port (default 8080)")
	cmdRoot.Flags().StringVarP(&mode, "mode", "m", "", "prod or debug mode (default \"prod\")")

	cmdServe := &cobra.Command{
		Use:   "serve",
		Short: "Start server",
		Long:  `Start server`,
		Run: func(cmd *cobra.Command, args []string) {
			startApp()
		},
	}
	cmdServe.Flags().StringVarP(&address, "address", "a", "", "server hostname or ip adress (default \"localhost\")")
	cmdServe.Flags().StringVarP(&port, "port", "p", "", "server port (default 8080)")
	cmdServe.Flags().StringVarP(&mode, "mode", "m", "", "prod or debug mode (default \"prod\")")

	cmdService := &cobra.Command{
		Use:   "service",
		Short: "Manage service (install, remove, start, stop, run)",
		Long:  `Manage go-shiny service (install, remove, start, stop, run)`,
	}

	cmdInstall := &cobra.Command{
		Use:   "install",
		Short: "Install go-shiny as a service",
		Long:  `Install go-shiny as a service`,
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Install()
			if err != nil {
				fmt.Printf("Failed to install: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" installed.\n", svcConfig.DisplayName)
		},
	}

	var cmdRemove = &cobra.Command{
		Use:   "remove",
		Short: "Remove go-shiny service",
		Long:  `Remove go-shiny service if previously installed`,
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Uninstall()
			if err != nil {
				fmt.Printf("Failed to remove: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" removed.\n", svcConfig.DisplayName)
		},
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start go-shiny service",
		Long:  `Start go-shiny service if previously installed`,
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Start()
			if err != nil {
				fmt.Printf("Failed to start: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" started.\n", svcConfig.DisplayName)
		},
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop go-shiny service",
		Long:  `Stop go-shiny service`,
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Stop()
			if err != nil {
				fmt.Printf("Failed to stop: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" stopped.\n", svcConfig.DisplayName)
		},
	}

	cmdRoot.AddCommand(cmdServe, cmdService)
	cmdService.AddCommand(cmdInstall, cmdRemove, cmdStart, cmdStop)
	cmdRoot.Execute()
}
