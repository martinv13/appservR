package models

import (
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

type ShinyApp struct {
	gorm.Model
	AppName        string `gorm:"unique"`
	Path           string
	AppSource      string
	AppDir         string
	Workers        int
	Active         bool
	RestrictAccess bool
	AllowedGroups  []*Group `gorm:"many2many:app_allowed_groups;"`
}

type AppModel interface {
	All() []*ShinyApp
	FindByName(appName string) (*ShinyApp, error)
	Save(app ShinyApp, oldAppName string) error
	Delete(appName string) error
	AsMap(app ShinyApp) (map[string]interface{}, error)
}

type AppModelDB struct {
	sync.RWMutex
	DB         *gorm.DB
	groupModel *GroupModelDB
	apps       map[string]*ShinyApp
}

func NewAppModelDB(db *gorm.DB, groupModel *GroupModelDB) (*AppModelDB, error) {

	appModel := AppModelDB{
		DB:         db,
		apps:       map[string]*ShinyApp{},
		groupModel: groupModel,
	}

	apps := []*ShinyApp{}

	err := db.Find(&apps).Error
	if err != nil {
		return nil, err
	}

	if len(apps) == 0 {
		defaultApp := ShinyApp{
			AppName:        "sample-app",
			Path:           "/",
			AppSource:      "sample-app",
			Workers:        2,
			Active:         true,
			RestrictAccess: false,
		}
		err = db.Create(&defaultApp).Error
		if err != nil {
			return nil, err
		}
		apps = append(apps, &defaultApp)
	}

	appModel.Lock()
	for _, app := range apps {
		appModel.apps[app.AppName] = app
	}
	appModel.Unlock()

	return &appModel, nil

}

// Get all apps
func (m *AppModelDB) All() []*ShinyApp {
	m.RLock()
	defer m.RUnlock()
	apps := make([]*ShinyApp, len(m.apps), len(m.apps))
	j := 0
	for i := range m.apps {
		apps[j] = m.apps[i]
		j++
	}
	return apps
}

// Find a specific app by app name
func (m *AppModelDB) FindByName(appName string) (*ShinyApp, error) {
	m.RLock()
	defer m.RUnlock()
	res, ok := m.apps[appName]
	if ok {
		return res, nil
	}
	return nil, errors.New("App not found")
}

// Create or update a shinyapp to the database
func (m *AppModelDB) Save(app ShinyApp, oldAppName string) error {

	if app.RestrictAccess {
		allGroups, err := m.groupModel.AllNames()
		if err != nil {
			return errors.New("Unable to retrieve groups")
		}
		allGroupsMap := make(map[string]bool)
		for _, g := range allGroups {
			allGroupsMap[g] = true
		}
		for _, g := range app.AllowedGroups {
			if _, ok := allGroupsMap[g.Name]; !ok {
				return fmt.Errorf("Specifying non existing groups '%s' for app: %s", g.Name, app.AppName)
			}
		}
	}

	if app.AppName == "new" {
		return errors.New("App name cannot be 'new'")
	}

	if oldAppName == "new" {
		err := m.DB.Create(&app).Error
		if err != nil {
			return errors.New("Failed to create new app")
		}
		m.Lock()
		m.apps[app.AppName] = &app
		m.Unlock()
		return nil
	}

	var currentApp ShinyApp

	err := m.DB.First(&currentApp, "app_name=?", oldAppName).Error
	if err != nil {
		return fmt.Errorf("Update failed. Could not find app: %s", oldAppName)
	}
	updateMap := map[string]interface{}{
		"AppName":        app.AppName,
		"Path":           app.Path,
		"AppDir":         app.AppDir,
		"Workers":        app.Workers,
		"Active":         app.Active,
		"RestrictAccess": app.RestrictAccess,
	}

	tx := m.DB.Begin()
	err = tx.Model(&currentApp).Updates(updateMap).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error while updating app: %s", oldAppName)
	}
	err = tx.Model(&currentApp).Association("AllowedGroups").Replace(app.AllowedGroups)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Error while updating allowed groups for app: %s", oldAppName)
	}
	tx.Commit()

	m.Lock()
	if _, ok := m.apps[oldAppName]; ok {
		delete(m.apps, oldAppName)
	}
	m.apps[app.AppName] = &app
	m.Unlock()

	return nil
}

// Delete an app
func (m *AppModelDB) Delete(appName string) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.apps[appName]; ok {
		app := ShinyApp{}
		err := m.DB.Unscoped().Where("app_name = ?", appName).Delete(&app).Error
		if err != nil {
			return fmt.Errorf("Error while deleting app: %s", appName)
		}
		delete(m.apps, appName)
		return nil
	} else {
		return errors.New("App not found")
	}
}

// Get an app as a map, directly usable in template
func (m *AppModelDB) AsMap(app ShinyApp) (map[string]interface{}, error) {
	allGroups, err := m.groupModel.AllNames()
	if err != nil {
		return nil, errors.New("Unable to retrieve groups")
	}
	groups := make(map[string]bool)
	for _, g := range allGroups {
		groups[g] = false
	}
	allowedGroups := app.AllowedGroups
	if allowedGroups == nil {
		allowedGroups = make([]*Group, 0)
	}
	for _, g := range allowedGroups {
		groups[g.Name] = true
	}
	res := map[string]interface{}{
		"AppName":        app.AppName,
		"Path":           app.Path,
		"AppDir":         app.AppDir,
		"Workers":        app.Workers,
		"Active":         app.Active,
		"RestrictAccess": app.RestrictAccess,
		"AllowedGroups":  groups,
	}
	return res, nil
}
