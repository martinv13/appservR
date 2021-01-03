package models

import (
	"sort"
	"strings"

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

type ShinyAppUpdate struct {
	ID      string `form:"id" binding:"required"`
	Name    string `form:"name" binding:"required"`
	Path    string `form:"path" binding:"required"`
	AppDir  string `form:"app_dir" binding:"required"`
	Workers int    `form:"workers" binding:"required"`
	Active  bool   `form:"active" binding:"required"`
}

var shinyApps = make(map[string]ShinyApp)

func (h ShinyApp) Update(shinyAppPayload ShinyAppUpdate) (*ShinyApp, error) {
	//	db := db.GetDB()

	id := uuid.NewV4()
	appId := id.String()
	if _, ok := shinyApps[shinyAppPayload.ID]; ok {
		appId = shinyAppPayload.ID
	}

	app := ShinyApp{
		ID:      appId,
		Name:    shinyAppPayload.Name,
		Path:    shinyAppPayload.Path,
		AppDir:  shinyAppPayload.AppDir,
		Workers: shinyAppPayload.Workers,
		Active:  shinyAppPayload.Active,
	}

	shinyApps[appId] = app

	return &app, nil
}

func (h ShinyApp) Init() {

	shinyApps["feaj66DHS_hdf"] = ShinyApp{
		ID:      "feaj66DHS_hdf",
		Name:    "Main app",
		Path:    "/",
		AppDir:  "C:/Users/marti/code/shiny-apps/shiny-apps/test-app",
		Workers: 2,
		Active:  true,
	}

	shinyApps["trf76dzd-DEz"] = ShinyApp{
		ID:      "trf76dzd-DEz",
		Name:    "Sub app",
		Path:    "/testapp",
		AppDir:  "C:/Users/marti/code/shiny-apps/shiny-apps/test-app2",
		Workers: 2,
		Active:  true,
	}

}

type byPath []ShinyApp

func (a byPath) Len() int {
	return len(a)
}
func (a byPath) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byPath) Less(i, j int) bool {
	return !strings.HasPrefix(a[i].Path, a[j].Path)
}

func (h ShinyApp) GetAll() []ShinyApp {
	apps := make(byPath, len(shinyApps))
	i := 0
	for k := range shinyApps {
		apps[i] = shinyApps[k]
		i++
	}
	sort.Sort(apps)
	appSlice := ([]ShinyApp)(apps)

	return appSlice
}
