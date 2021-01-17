package appproxy

import (
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/martinv13/go-shiny/models"
)

var appsByID = make(map[string]*AppProxy)

var byPath []*AppProxy

func StartApps() error {
	appData := new(models.ShinyApp)
	appData.Init()

	apps := appData.GetAll()
	for i := range apps {
		app := NewAppProxy(apps[i])
		appsByID[apps[i].ID] = app
		err := app.Start()
		if err != nil {
			return err
		}
		byPath = append(byPath, app)
	}
	sort.SliceStable(byPath, func(i, j int) bool {
		return !strings.HasPrefix(byPath[i].ShinyApp.Path, byPath[j].ShinyApp.Path)
	})
	return nil
}

func MatchApp(r *http.Request) (*AppProxy, error) {

	reqURI, _ := url.Parse(r.RequestURI)

	if reqURI.Path != "/" {
		reqURI.Path = strings.TrimSuffix(reqURI.Path, "/")
	}

	for i := range byPath {
		if byPath[i].ShinyApp.Path == reqURI.Path {
			return byPath[i], nil
		}
	}

	cookie, err := r.Cookie("GO_SHINY_APP_ID")
	if err == nil {
		if app, ok := appsByID[cookie.Value]; ok {
			return app, nil
		}
	}

	return &AppProxy{}, errors.New("No app found")
}
