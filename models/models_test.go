package models

import (
	"errors"
	"testing"

	"github.com/appservR/appservR/modules/config"
	"gorm.io/gorm"
)

type MockConfig struct {
	keys   map[string]string
	logger config.Logger
}

func (c *MockConfig) ExecutableFolder() string {
	return "."
}

func (c *MockConfig) Logger() *config.Logger {
	return &c.logger
}

func (c *MockConfig) GetString(key string) string {
	res, ok := c.keys[key]
	if ok {
		return res
	}
	return ""
}

func setUp() (*gorm.DB, error) {

	conf := &MockConfig{
		keys: map[string]string{
			"database.type": "sqlite",
			"database.path": "memory",
		},
		logger: config.NewLogger(0),
	}

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
		err := userModel.Save(User{Username: "admin", DisplayedName: "John", Password: "test"}, "new")
		if err != nil {
			t.Error("failed to create admin user")
		}
		err = userModel.Save(User{Username: "user1", DisplayedName: "James", Password: "test"}, "new")
		if err != nil {
			t.Error("failed to create non-admin user")
		}
		user, err := userModel.Find("admin")
		if err != nil {
			t.Error("failed to find admin user")
		}
		userData, err := userModel.AsMap(user)
		groups := userData["Groups"].(map[string]bool)
		gr, ok := groups["admins"]
		if err != nil || len(groups) != 1 || !ok || !gr || len(user.Groups) != 1 || user.Groups[0].Name != "admins" {
			t.Error("user not registered as admin")
		}
		user1 := User{Username: "user1", Password: "test"}
		user1, err = userModel.Login(user1)
		if err != nil {
			t.Error("failed to login")
		}
		user2 := User{Username: "user1", Password: "test2"}
		user2, err = userModel.Login(user2)
		if err == nil {
			t.Error("login should fail")
		}
		userData, err = userModel.AsMap(user1)
		groups = userData["Groups"].(map[string]bool)
		gr, ok = groups["admins"]
		if err != nil || len(groups) != 1 || !ok || gr || len(user1.Groups) != 0 {
			t.Error("user should not be admin")
		}
	})

	t.Run("app=find", func(t *testing.T) {
		_, err := appModel.Find("sample-app")
		if err != nil {
			t.Error("cannot get default app")
		}
	})
	t.Run("app=all", func(t *testing.T) {
		apps, err := appModel.All()
		if err != nil || len(apps) != 1 {
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
