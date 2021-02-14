package appproxy

import (
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/martinv13/go-shiny/models"
)

var appsByName = make(map[string]*AppProxy)

var byPath []*AppProxy

func StartApps() error {
	appData := models.ShinyApp{}

	apps := appData.GetAll()
	for i := range apps {
		app := NewAppProxy(apps[i])
		appsByName[apps[i].AppName] = app
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

func GetSession(r *http.Request) (*Session, error) {

	reqURI, _ := url.Parse(r.RequestURI)

	if reqURI.Path != "/" {
		reqURI.Path = strings.TrimSuffix(reqURI.Path, "/")
	}

	for i := range byPath {
		if byPath[i].ShinyApp.Path == reqURI.Path {
			session, err := byPath[i].GetSession("")
			if err != nil {
				return nil, err
			}
			return session, nil
		}
	}

	appCookie, err := r.Cookie("go_shiny_appid")
	if err == nil {
		if app, ok := appsByName[appCookie.Value]; ok {
			sessCookie, _ := r.Cookie("go_shiny_session")
			session, err := app.GetSession(sessCookie.Value)
			if err != nil {
				return nil, err
			}
			return session, nil
		}
	}

	return nil, errors.New("No matching app found")
}
