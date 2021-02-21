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

// Returns the status of the apps
func GetAllStatus() map[string]interface{} {

	status := map[string]interface{}{}

	for n, app := range appsByName {
		nb_running := 0
		nb_rollingout := 0
		for _, i := range app.Instances {
			if i.State == "RUNNING" {
				nb_running++
			} else if i.State == "ROLLING_OUT" {
				nb_rollingout++
			}
		}
		status[n] = map[string]interface{}{
			"AppName":        app.ShinyApp.AppName,
			"Active":         app.ShinyApp.Active,
			"RunningInst":    nb_running,
			"RollingOutInst": nb_rollingout,
			"ConnectedUsers": len(app.Sessions),
		}
	}
	return status

}

// Returns the status of a given app
func GetStatus(appName string) (map[string]interface{}, error) {
	app, ok := appsByName[appName]
	if !ok {
		return nil, errors.New("App not found")
	}
	nbRunning := 0
	nbRollingout := 0
	stdErr := []string{}
	for _, i := range app.Instances {
		if i.State == "RUNNING" {
			nbRunning++
		} else if i.State == "ROLLING_OUT" {
			nbRollingout++
		}
		stdErr = append(stdErr, i.StdErr)
	}
	return map[string]interface{}{
		"AppName":        app.ShinyApp.AppName,
		"Active":         app.ShinyApp.Active,
		"RunningInst":    nbRunning,
		"RollingOutInst": nbRollingout,
		"ConnectedUsers": len(app.Sessions),
		"StdErr":         stdErr,
	}, nil
}
