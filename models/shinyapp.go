package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ShinyApp struct {
	gorm.Model
	AppName        string `gorm:"unique"`
	Path           string
	AppDir         string
	Workers        int
	Active         bool
	RestrictAccess bool
	AllowedGroups  []*Group `gorm:"many2many:app_allowed_groups;"`
}

var shinyApps = make(map[string]*ShinyApp)

func (h ShinyApp) Init(db *gorm.DB) error {

	apps := []*ShinyApp{}

	err := db.Find(&apps).Error
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		defaultApp := ShinyApp{
			AppName:        "demo-app",
			Path:           "/",
			AppDir:         "C:/Users/marti/code/shiny-apps/shiny-apps/test-app",
			Workers:        2,
			Active:         true,
			RestrictAccess: false,
		}
		db.Create(&defaultApp)
		apps = append(apps, &defaultApp)
	}

	for _, app := range apps {
		shinyApps[app.AppName] = app
	}

	return nil

}

func (app *ShinyApp) Get() error {
	res, ok := shinyApps[app.AppName]
	if ok {
		*app = *res
		return nil
	}
	return errors.New("App not found")
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

func (h ShinyApp) GetAllMapSlice(db *gorm.DB) []map[string]interface{} {
	res := make([]map[string]interface{}, len(shinyApps), len(shinyApps))
	j := 0
	for i := range shinyApps {
		app := shinyApps[i]
		res[j] = map[string]interface{}{
			"AppName":        app.AppName,
			"Path":           app.Path,
			"AppDir":         app.AppDir,
			"Workers":        app.Workers,
			"Active":         app.Active,
			"RestrictAccess": app.RestrictAccess,
			"AllowedGroups":  app.GroupsMap(db),
		}
		j++
	}
	return res
}

func (app *ShinyApp) Update(db *gorm.DB, oldAppName string) error {

	if app.RestrictAccess {
		groupNames := make([]string, len(app.AllowedGroups), len(app.AllowedGroups))
		for i, g := range app.AllowedGroups {
			groupNames[i] = g.Name
		}
		groups := []*Group{}
		err := db.Where("name IN ?", groupNames).Find(&groups).Error
		if err != nil {
			return fmt.Errorf("Specifying non existing groups for app: %s", oldAppName)
		}
		app.AllowedGroups = groups
	}

	if app.AppName == "new" {
		return errors.New("App name cannot be 'new'")
	}

	if oldAppName == "new" {
		err := db.Create(&app).Error
		if err != nil {
			return errors.New("Failed to create new app")
		}
		return nil
	}

	var currentApp ShinyApp

	err := db.First(&currentApp, "app_name=?", oldAppName).Error
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

	tx := db.Begin()
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

	if _, ok := shinyApps[oldAppName]; ok {
		delete(shinyApps, oldAppName)
	}
	shinyApps[app.AppName] = app

	return nil
}

func (app *ShinyApp) Delete(db *gorm.DB) error {
	if _, ok := shinyApps[app.AppName]; ok {
		delete(shinyApps, app.AppName)
		return nil
	} else {
		return errors.New("App not found")
	}
}

// Function to retrieve groups as a map of boolean for the current app
func (app *ShinyApp) GroupsMap(db *gorm.DB) map[string]bool {
	groupsMap := map[string]bool{}
	groups, err := GetAllGroupNames(db)
	if err != nil {
		fmt.Println("Unable to retrieve groups")
	}
	for i := range groups {
		groupsMap[groups[i]] = false
	}
	for i := range app.AllowedGroups {
		groupsMap[app.AllowedGroups[i].Name] = true
	}
	return groupsMap
}
