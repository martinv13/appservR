package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ShinyApp struct {
	gorm.Model
	AppName         string `gorm:"unique"`
	Path            string
	AppSource       string
	AppDir          string
	GitSourceUrl    string
	GitSourceBranch string
	GitSourceToken  string
	Workers         int
	Active          bool
	RestrictAccess  bool
	AllowedGroups   []Group `gorm:"many2many:app_allowed_groups;"`
}

type AppModel interface {
	All() ([]ShinyApp, error)
	Find(appName string) (ShinyApp, error)
	Save(app ShinyApp, oldAppName string) error
	Delete(appName string) error
	AsMap(app ShinyApp) (map[string]interface{}, error)
	AsMapSlice(apps []ShinyApp) ([]map[string]interface{}, error)
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

	app := ShinyApp{}

	err := db.First(&app).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		} else {
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
		}
	}
	return &appModel, nil
}

// Get all apps
func (m *AppModelDB) All() ([]ShinyApp, error) {
	var apps []ShinyApp
	err := m.DB.Find(&apps).Error
	if err != nil {
		return []ShinyApp{}, errors.New("Unable to retrieve apps")
	}
	return apps, nil
}

// Find a specific app by app name
func (m *AppModelDB) Find(appName string) (ShinyApp, error) {
	var app ShinyApp
	err := m.DB.First(&app, "app_name = ?", appName).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ShinyApp{}, errors.New("Unable to retrieve app from db")
		} else {
			return ShinyApp{}, fmt.Errorf("App %s does not exist.", appName)
		}
	}
	return app, nil
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
		return nil
	}

	var currentApp ShinyApp

	err := m.DB.First(&currentApp, "app_name=?", oldAppName).Error
	if err != nil {
		return fmt.Errorf("Update failed. Could not find app: %s", oldAppName)
	}
	updateMap := map[string]interface{}{
		"AppName":         app.AppName,
		"Path":            app.Path,
		"AppDir":          app.AppDir,
		"GitSourceUrl":    app.GitSourceUrl,
		"GitSourceBranch": app.GitSourceBranch,
		"GitSourceToken":  app.GitSourceToken,
		"Workers":         app.Workers,
		"Active":          app.Active,
		"RestrictAccess":  app.RestrictAccess,
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

	return nil
}

// Delete an app
func (m *AppModelDB) Delete(appName string) error {
	var app ShinyApp
	err := m.DB.Unscoped().Where("app_name = ?", appName).Delete(&app).Error
	if err != nil {
		return fmt.Errorf("Error while deleting app: %s", appName)
	}
	return nil
}

// Get an app as a map, directly usable in template
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
func (m *AppModelDB) AsMap(app ShinyApp) (map[string]interface{}, error) {
	allGroups, err := m.groupModel.AllNames()
	if err != nil {
		return nil, errors.New("Unable to retrieve groups")
	}
	return map[string]interface{}{
		"AppName":         app.AppName,
		"Path":            app.Path,
		"AppDir":          app.AppDir,
		"GitSourceUrl":    app.GitSourceUrl,
		"GitSourceBranch": app.GitSourceBranch,
		"GitSourceToken":  app.GitSourceToken,
		"Workers":         app.Workers,
		"Active":          app.Active,
		"RestrictAccess":  app.RestrictAccess,
		"AllowedGroups":   m.groupsMap(app.AllowedGroups, allGroups),
	}, nil
}

// Get a slice of apps as a slice of maps, directly usable in template
func (m *AppModelDB) AsMapSlice(apps []ShinyApp) ([]map[string]interface{}, error) {
	allGroups, err := m.groupModel.AllNames()
	if err != nil {
		return nil, errors.New("Unable to retrieve groups")
	}
	appsMap := make([]map[string]interface{}, len(apps), len(apps))
	for i, app := range apps {
		appsMap[i] = map[string]interface{}{
			"AppName":        app.AppName,
			"Path":           app.Path,
			"Active":         app.Active,
			"RestrictAccess": app.RestrictAccess,
			"AllowedGroups":  m.groupsMap(app.AllowedGroups, allGroups),
		}
	}
	return appsMap, nil
}
