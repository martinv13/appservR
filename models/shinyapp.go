package models

import (
	"errors"

	uuid "github.com/satori/go.uuid"
)

type ShinyApp struct {
	ID      string
	Name    string
	Path    string
	AppDir  string
	Workers int
	Active  bool
}

type ShinyAppPayload struct {
	ID      string `form:"id" binding:"required"`
	Name    string `form:"name" binding:"required"`
	Path    string `form:"path" binding:"required"`
	AppDir  string `form:"app_dir" binding:"required"`
	Workers int    `form:"workers" binding:"required"`
	Active  bool   `form:"active" binding:"required"`
}

var shinyApps = make(map[string]*ShinyApp)

func (h ShinyApp) Update(shinyAppPayload ShinyAppPayload) (*ShinyApp, error) {
	id := uuid.NewV4()
	appId := id.String()
	if _, ok := shinyApps[shinyAppPayload.ID]; ok {
		appId = shinyAppPayload.ID
	}

	app := &ShinyApp{
		ID:      appId,
		Name:    shinyAppPayload.Name,
		Path:    shinyAppPayload.Path,
		AppDir:  shinyAppPayload.AppDir,
		Workers: shinyAppPayload.Workers,
		Active:  shinyAppPayload.Active,
	}

	shinyApps[appId] = app

	return app, nil
}

func (h ShinyApp) Delete(appID string) error {
	if _, ok := shinyApps[appID]; ok {
		delete(shinyApps, appID)
		return nil
	} else {
		return errors.New("App not found")
	}
}

func (h ShinyApp) GetAll() []*ShinyApp {
	apps := make([]*ShinyApp, len(shinyApps), len(shinyApps))
	j := 0
	for i := range shinyApps {
		apps[j] = shinyApps[i]
		j++
	}
	return apps
}

func (h ShinyApp) Init() {

	shinyApps["feaj66DHS_hdf"] = &ShinyApp{
		ID:      "feaj66DHS_hdf",
		Name:    "Main app",
		Path:    "/",
		AppDir:  "C:/Users/marti/code/shiny-apps/shiny-apps/test-app",
		Workers: 2,
		Active:  true,
	}

	shinyApps["trf76dzd-DEz"] = &ShinyApp{
		ID:      "trf76dzd-DEz",
		Name:    "Sub app",
		Path:    "/testapp",
		AppDir:  "C:/Users/marti/code/shiny-apps/shiny-apps/test-app2",
		Workers: 2,
		Active:  true,
	}

}
