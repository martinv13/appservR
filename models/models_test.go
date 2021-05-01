package models

import (
	"errors"
	"testing"

	"github.com/martinv13/go-shiny/modules/config"
	"gorm.io/gorm"
)

func setUp() (*gorm.DB, error) {

	conf, err := config.NewConfig()
	if err != nil {
		return nil, errors.New("unable to initialize config")
	}

	conf.Set("database.type", "sqlite")
	conf.Set("database.path", "memory")

	db, err := NewDB(conf)

	if err != nil {
		return nil, errors.New("unable to initialize in memory database")
	}

	return db, nil
}

func TestDataModelDB(t *testing.T) {

	db, err := setUp()
	if err != nil {
		t.Error("unable to initialize database")
	}

	groupModel := NewGroupModelDB(db)
	if err != nil {
		t.Error("unable to initialize group model")
	}

	appModel, err := NewAppModelDB(db, groupModel)
	if err != nil {
		t.Error("unable to initialize app model")
	}

	userModel := NewUserModelDB(db, groupModel)
	if err != nil {
		t.Error("unable to initialize user model")
	}

	t.Run("user=lifecycle", func(t *testing.T) {
		err := userModel.Save(&User{Username: "admin", DisplayedName: "John", Password: "test"}, "new")
		if err != nil {
			t.Error("failed to create admin user")
		}
		err = userModel.Save(&User{Username: "user1", DisplayedName: "James", Password: "test"}, "new")
		if err != nil {
			t.Error("failed to create non-admin user")
		}
		user, err := userModel.FindByUsername("admin")
		if err != nil {
			t.Error("failed to find admin user")
		}
		groups := userModel.GroupsMap(user)
		if gr, ok := groups["admins"]; len(groups) != 1 || !ok || !gr || len(user.Groups) != 1 || user.Groups[0].Name != "admins" {
			t.Error("user not registered as admin")
		}
		user1 := User{Username: "user1", Password: "test"}
		err = userModel.Login(&user1)
		if err != nil {
			t.Error("failed to login")
		}
		user2 := User{Username: "user1", Password: "test2"}
		err = userModel.Login(&user2)
		if err == nil {
			t.Error("login should fail")
		}
		groups = userModel.GroupsMap(&user1)
		if gr, ok := groups["admins"]; len(groups) != 1 || !ok || gr || len(user1.Groups) != 0 {
			t.Error("user should not be admin")
		}
	})

	t.Run("app=find", func(t *testing.T) {
		_, err := appModel.FindByName("sample-app")
		if err != nil {
			t.Error("cannot get default app")
		}
	})
	t.Run("app=all", func(t *testing.T) {
		apps := appModel.All()
		if len(apps) != 1 {
			t.Error("did not return exactly one app")
		}
	})
	t.Run("app=create", func(t *testing.T) {
		app := ShinyApp{
			AppName: "test-app",
			Path:    "/test-app",
			AppDir:  "shinyapps/sample-app/",
			Workers: 1,
		}
		err := appModel.Save(app, "new")
		if err != nil {
			t.Error("cannot create app")
		}
	})

}
