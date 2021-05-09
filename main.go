package main

//go:generate go run -tags=dev generate_assets.go

import (
	"log"

	"fmt"

	"runtime/debug"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

var (
	host   string
	port   string
	mode   string
	logger service.Logger
)

func startApp(cmd *cobra.Command) {
	server, err := InitializeServer(cmd)
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
	startApp(nil)
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
			startApp(cmd)
		},
	}
	cmdRoot.Flags().StringP("address", "a", "localhost", "server hostname or ip adress")
	cmdRoot.Flags().StringP("port", "p", "8080", "server port")
	cmdRoot.Flags().StringP("mode", "m", "prod", "prod or debug mode")

	cmdServe := &cobra.Command{
		Use:   "serve",
		Short: "Start server",
		Long:  `Start server`,
		Run: func(cmd *cobra.Command, args []string) {
			startApp(cmd)
		},
	}
	cmdServe.Flags().StringP("address", "a", "localhost", "server hostname or ip adress")
	cmdServe.Flags().StringP("port", "p", "8080", "server port")
	cmdServe.Flags().StringP("mode", "m", "prod", "prod or debug mode")

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
			err = s.Install()
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
			err = s.Uninstall()
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
			err = s.Start()
			if err != nil {
				fmt.Printf("Failed to start: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" started.\n", svcConfig.DisplayName)
		},
	}

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Run go-shiny service",
		Long:  `Run go-shiny service`,
		Run: func(cmd *cobra.Command, args []string) {
			err = s.Run()
			if err != nil {
				fmt.Printf("Failed to run: %s\n", err)
			}
		},
	}

	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stop go-shiny service",
		Long:  `Stop go-shiny service`,
		Run: func(cmd *cobra.Command, args []string) {
			err = s.Stop()
			if err != nil {
				fmt.Printf("Failed to stop: %s\n", err)
				return
			}
			fmt.Printf("Service \"%s\" stopped.\n", svcConfig.DisplayName)
		},
	}

	cmdRoot.AddCommand(cmdServe, cmdService)
	cmdService.AddCommand(cmdRun, cmdInstall, cmdRemove, cmdStart, cmdStop)
	cmdRoot.Execute()
}
