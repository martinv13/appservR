package appproxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/martinv13/go-shiny/models"
)

type AppProxy struct {
	Instances []*shinyAppInstance
}

func InitAppProxy(app models.ShinyApp) (AppProxy, error) {
	appProxy := AppProxy{}
	for w := 0; w < app.Workers; w++ {
		inst, err := SpawnApp(app.ID, app.AppDir)
		if err != nil {
			return AppProxy{}, err
		}
		appProxy.Instances = append(appProxy.Instances, inst)
	}
	return appProxy, nil
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

var runningApps = make(map[string]AppProxy)

func StartApps() error {
	appData := new(models.ShinyApp)
	appData.Init()

	apps := appData.GetAll()
	for i := range apps {
		app, err := InitAppProxy(apps[i])
		runningApps[apps[i].ID] = app
		if err != nil {
			return err
		}
	}
	return nil
}

func MatchApp(req *http.Request) (AppProxy, error) {

	appData := new(models.ShinyApp)
	apps := appData.GetAll()

	reqURI, _ := url.Parse(req.RequestURI)

	fmt.Println(reqURI.Path)

	for i := range apps {
		if apps[i].Path == reqURI.Path {
			return runningApps[apps[i].ID], nil
		}
	}
	return AppProxy{}, errors.New("No app found")
}
