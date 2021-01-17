package appproxy

import (
	"errors"

	"github.com/martinv13/go-shiny/models"
)

type AppProxy struct {
	ShinyApp  *models.ShinyApp
	Instances []*shinyAppInstance
}

func NewAppProxy(app *models.ShinyApp) *AppProxy {
	return &AppProxy{ShinyApp: app, Instances: []*shinyAppInstance{}}
}

func (appProxy *AppProxy) Start() error {
	for w := 0; w < appProxy.ShinyApp.Workers; w++ {
		inst, err := SpawnApp(appProxy.ShinyApp.ID, appProxy.ShinyApp.AppDir)
		if err != nil {
			return err
		}
		appProxy.Instances = append(appProxy.Instances, inst)
	}
	return nil
}

func (appProxy AppProxy) GetPort() (string, error) {

	if len(appProxy.Instances) > 0 {
		ports := []string{}
		for i := range appProxy.Instances {
			if appProxy.Instances[i].State == "RUNNING" {
				ports = append(ports, appProxy.Instances[i].Port)
			}
		}
		if len(ports) > 0 {
			return ports[0], nil
		}
	}
	return "", errors.New("No running instance available")
}
