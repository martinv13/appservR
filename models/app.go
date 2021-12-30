package models

import (
	"errors"
	"fmt"

	"github.com/appservR/appservR/modules/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type App struct {
	gorm.Model
	Name            string `gorm:"unique"`
	Path            string
	AppSource       string
	AppDir          string
	GitSourceUrl    string
	GitSourceBranch string
	GitSourceToken  string
	Workers         int
	IsActive        bool
	RestrictAccess  int
	AllowedGroups   []Group `gorm:"many2many:app_allowed_groups;"`
}

type AppModel interface {
	All() ([]App, error)
	Find(name string) (App, error)
	Save(app App, oldName string) error
	Delete(name string) error
	AsMap(app App) (map[string]interface{}, error)
	AsMapSlice(apps []App) ([]map[string]interface{}, error)
}

type AppModelDB struct {
	DB         *gorm.DB
	groupModel *GroupModelDB
}

func NewAppModelDB(db *gorm.DB, groupModel *GroupModelDB) (*AppModelDB, error) {

	appModel := AppModelDB{
		DB:         db,
		groupModel: groupModel,
	}

	app := App{}

	err := db.First(&app).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultApp := App{
				Name:           "sample-app",
				Path:           "/",
				AppSource:      "sample-app",
				AppDir:         "apps/sample-app/",
				Workers:        2,
				IsActive:       true,
				RestrictAccess: config.AccessLevels.PUBLIC,
			}
			err = db.Create(&defaultApp).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &appModel, nil
}

// Get all apps
func (m *AppModelDB) All() ([]App, error) {
	var apps []App
	err := m.DB.Preload("AllowedGroups").Find(&apps).Error
	if err != nil {
		return []App{}, errors.New("unable to retrieve apps")
	}
	return apps, nil
}

// Find a specific app by app name
func (m *AppModelDB) Find(name string) (App, error) {
	var app App
	err := m.DB.Preload(clause.Associations).First(&app, "name = ?", name).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return App{}, errors.New("unable to retrieve app from db")
		} else {
			return App{}, fmt.Errorf("app %s does not exist", name)
		}
	}
	return app, nil
}

// Create or update a R app to the database
func (m *AppModelDB) Save(app App, oldName string) error {

	groupNames := make([]string, len(app.AllowedGroups))
	for i, g := range app.AllowedGroups {
		groupNames[i] = g.Name
	}
	var groups []Group
	err := m.DB.Where("name IN ?", groupNames).Find(&groups).Error
	if err != nil {
		return fmt.Errorf("specifying non existing groups")
	}

	if app.Name == "new" {
		return errors.New("app name cannot be 'new'")
	}

	if oldName == "new" {
		err := m.DB.Create(&app).Error
		if err != nil {
			return errors.New("failed to create new app")
		}
		return nil
	}

	var currentApp App

	err = m.DB.First(&currentApp, "name=?", oldName).Error
	if err != nil {
		return fmt.Errorf("update failed; could not find app: %s", oldName)
	}
	updateMap := map[string]interface{}{
		"Name":            app.Name,
		"Path":            app.Path,
		"AppDir":          app.AppDir,
		"GitSourceUrl":    app.GitSourceUrl,
		"GitSourceBranch": app.GitSourceBranch,
		"GitSourceToken":  app.GitSourceToken,
		"Workers":         app.Workers,
		"IsActive":        app.IsActive,
		"RestrictAccess":  app.RestrictAccess,
	}

	tx := m.DB.Begin()
	err = tx.Model(&currentApp).Updates(updateMap).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while updating app: %s", oldName)
	}
	err = tx.Model(&currentApp).Association("AllowedGroups").Replace(groups)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while updating allowed groups for app: %s", oldName)
	}
	tx.Commit()

	return nil
}

// Delete an app
func (m *AppModelDB) Delete(name string) error {
	var app App
	err := m.DB.Unscoped().Where("name = ?", name).Delete(&app).Error
	if err != nil {
		return fmt.Errorf("error while deleting app: %s", name)
	}
	return nil
}

// Get a map of boolean values, representing allowed groups
func (m *AppModelDB) groupsMap(allowedGroups []Group, allGroups []string) map[string]bool {
	groupsMap := make(map[string]bool)
	for _, g := range allGroups {
		groupsMap[g] = false
	}
	for _, g := range allowedGroups {
		groupsMap[g.Name] = true
	}
	return groupsMap
}

// Get an app as a map, directly usable in template
func (m *AppModelDB) AsMap(app App) (map[string]interface{}, error) {
	allGroups, err := m.groupModel.AllNames()
	if err != nil {
		return nil, errors.New("unable to retrieve groups")
	}
	return map[string]interface{}{
		"Name":            app.Name,
		"Path":            app.Path,
		"AppDir":          app.AppDir,
		"GitSourceUrl":    app.GitSourceUrl,
		"GitSourceBranch": app.GitSourceBranch,
		"GitSourceToken":  app.GitSourceToken,
		"Workers":         app.Workers,
		"IsActive":        app.IsActive,
		"RestrictAccess":  app.RestrictAccess,
		"AllowedGroups":   m.groupsMap(app.AllowedGroups, allGroups),
	}, nil
}

// Get a slice of apps as a slice of maps, directly usable in template
func (m *AppModelDB) AsMapSlice(apps []App) ([]map[string]interface{}, error) {
	allGroups, err := m.groupModel.AllNames()
	if err != nil {
		return nil, errors.New("unable to retrieve groups")
	}
	appsMap := make([]map[string]interface{}, len(apps))
	for i, app := range apps {
		appsMap[i] = map[string]interface{}{
			"Name":           app.Name,
			"Path":           app.Path,
			"IsActive":       app.IsActive,
			"RestrictAccess": app.RestrictAccess,
			"AllowedGroups":  m.groupsMap(app.AllowedGroups, allGroups),
		}
	}
	return appsMap, nil
}
