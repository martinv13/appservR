package appsource

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/modules/config"
)

type AppSource interface {
	Path() string
	Cleanup() error
	Error() error
}

type AppSourceDir struct {
	AppDir string
	err    error
}

func NewAppSource(app models.ShinyApp, conf *config.Config) AppSource {
	if app.AppSource == "sample-app" {
		return NewAppSourceSampleApp(app, conf)
	} else {
		return NewAppSourceDir(app, conf)
	}
}

func NewAppSourceDir(app models.ShinyApp, conf *config.Config) *AppSourceDir {
	path := app.AppDir
	if !filepath.IsAbs(path) {
		path = filepath.Join(conf.ExecutableFolder, path)
	}
	_, err := os.Stat(path)
	if err != nil {
		return &AppSourceDir{AppDir: "", err: errors.New("App directory path does not exist")}
	}
	return &AppSourceDir{AppDir: path}
}

func NewAppSourceSampleApp(app models.ShinyApp, conf *config.Config) *AppSourceDir {
	path := conf.ExecutableFolder + "/shinyapps"
	_, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, os.ModeDir)
		if err != nil {
			return &AppSourceDir{AppDir: "", err: fmt.Errorf("Unable to create directory %s", path)}
		}
	}
	path = path + "/sample-app"
	_, err = os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, os.ModeDir)
		if err != nil {
			return &AppSourceDir{AppDir: "", err: fmt.Errorf("Unable to create directory %s", path)}
		}
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return &AppSourceDir{AppDir: "", err: fmt.Errorf("Unable to get absolute path")}
	}
	_, err = os.Stat(path + "/app.R")
	if err != nil {
		f, err := os.Create(path + "/app.R")
		if err != nil {
			return &AppSourceDir{AppDir: "", err: errors.New("Unable to write file app.R")}
		}
		defer f.Close()
		_, err = f.WriteString(sampleApp)
		if err != nil {
			return &AppSourceDir{AppDir: "", err: errors.New("Unable to write file app.R")}
		}
	}
	return &AppSourceDir{AppDir: path}
}

// Get shiny app directory path
func (s *AppSourceDir) Path() string {
	return s.AppDir
}

// Get shiny app source status
func (s *AppSourceDir) Error() error {
	return s.err
}

// Remove shiny app directory
func (s *AppSourceDir) Cleanup() error {
	err := os.RemoveAll(s.AppDir)
	if err != nil {
		return fmt.Errorf("Unable to delete path %s", s.AppDir)
	}
	return nil
}
