package appsource

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/config"
)

// A generic application source interface
type AppSource interface {
	Path() string   // get the path to the folder containing R app scripts
	Cleanup() error // cleanup if any when app is deleted
	Error() error   // get the error status if any
}

// A simple app source based on a local folder
type AppSourceDir struct {
	AppDir string
	err    error
}

func NewAppSource(app models.App, conf config.Config, checkOnly bool) AppSource {
	if app.AppSource == "sample-app" {
		return NewAppSourceSampleApp(app, conf)
	} else if app.AppSource == "directory" {
		return NewAppSourceDir(app, conf)
	}
	return nil
}

func NewAppSourceDir(app models.App, conf config.Config) *AppSourceDir {
	path := app.AppDir
	if !filepath.IsAbs(path) {
		path = filepath.Join(conf.ExecutableFolder(), path)
	}
	_, err := os.Stat(path)
	if err != nil {
		return &AppSourceDir{AppDir: "", err: errors.New("app directory path does not exist")}
	}
	_, err_app := os.Stat(filepath.Join(path, "app.R"))
	_, err_server := os.Stat(filepath.Join(path, "server.R"))
	_, err_ui := os.Stat(filepath.Join(path, "ui.R"))
	if err_app != nil && (err_server != nil || err_ui != nil) {
		return &AppSourceDir{AppDir: "", err: errors.New("app directory does not contain app.R or server.R and ui.R files")}
	}
	return &AppSourceDir{AppDir: path}
}

func NewAppSourceSampleApp(app models.App, conf config.Config) *AppSourceDir {
	path := conf.ExecutableFolder() + "/apps"
	_, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, 0700)
		if err != nil {
			return &AppSourceDir{AppDir: "", err: fmt.Errorf("unable to create directory %s", path)}
		}
	}
	path = path + "/sample-app"
	_, err = os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, 0700)
		if err != nil {
			return &AppSourceDir{AppDir: "", err: fmt.Errorf("unable to create directory %s", path)}
		}
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return &AppSourceDir{AppDir: "", err: fmt.Errorf("unable to get absolute path")}
	}
	_, err = os.Stat(path + "/app.R")
	if err != nil {
		f, err := os.Create(path + "/app.R")
		if err != nil {
			return &AppSourceDir{AppDir: "", err: errors.New("unable to write file app.R")}
		}
		defer f.Close()
		_, err = f.WriteString(sampleApp)
		if err != nil {
			return &AppSourceDir{AppDir: "", err: errors.New("unable to write file app.R")}
		}
	}
	return &AppSourceDir{AppDir: path}
}

// Get R app directory path
func (s *AppSourceDir) Path() string {
	return s.AppDir
}

// Get R app source status
func (s *AppSourceDir) Error() error {
	return s.err
}

// Nothing to do for cleanup: won't delete any files
func (s *AppSourceDir) Cleanup() error {
	return nil
}
