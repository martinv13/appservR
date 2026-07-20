package controllers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/appserver"
	"github.com/appservR/appservR/modules/ssehandler"
	"github.com/gin-gonic/gin"
)

var errTestDeleteFailed = errors.New("delete failed")

func newTestAppController(t *testing.T, apps ...models.App) (*AppController, *fakeAppModel) {
	t.Helper()
	appModel := newFakeAppModel(apps...)
	appServer, err := appserver.NewAppServer(appModel, ssehandler.NewMessageBroker(), newFakeConfig())
	if err != nil {
		t.Fatalf("NewAppServer failed: %v", err)
	}
	return NewAppController(appModel, appServer, newFakeConfig()), appModel
}

func newFormRequest(method, path string, form url.Values, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := newTestContext(method, path, nil)
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req
	c.Params = params
	return c, w
}

func TestAppControllerGetApps(t *testing.T) {
	ctl, _ := newTestAppController(t)
	c, w := newTestContext(http.MethodGet, "/admin/apps", nil)
	invoke(c, ctl.GetApps())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAppControllerGetAppNew(t *testing.T) {
	ctl, _ := newTestAppController(t)
	c, w := newTestContext(http.MethodGet, "/admin/apps/new", nil)
	c.Params = gin.Params{{Key: "appname", Value: "new"}}
	invoke(c, ctl.GetApp())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAppControllerGetAppUnknownFallsBackToNewForm(t *testing.T) {
	ctl, _ := newTestAppController(t)
	c, w := newTestContext(http.MethodGet, "/admin/apps/ghost", nil)
	c.Params = gin.Params{{Key: "appname", Value: "ghost"}}
	invoke(c, ctl.GetApp())
	// Current behavior: an unknown app name falls back to the "new app" form
	// with a 200, rather than a 404.
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (fallback to new-app form), got %d", w.Code)
	}
}

func TestAppControllerGetAppExisting(t *testing.T) {
	appDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(appDir, "app.R"), []byte("# app"), 0644); err != nil {
		t.Fatal(err)
	}
	existing := models.App{Name: "myapp", Path: "/myapp", AppSource: "directory", AppDir: appDir, IsActive: false}
	ctl, _ := newTestAppController(t, existing)

	c, w := newTestContext(http.MethodGet, "/admin/apps/myapp", nil)
	c.Params = gin.Params{{Key: "appname", Value: "myapp"}}
	invoke(c, ctl.GetApp())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAppControllerUpdateAppCreatesNewApp(t *testing.T) {
	appDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(appDir, "app.R"), []byte("# app"), 0644); err != nil {
		t.Fatal(err)
	}
	ctl, appModel := newTestAppController(t)

	form := url.Values{}
	form.Set("appname", "newapp")
	form.Set("path", "/newapp")
	form.Set("appsource", "directory")
	form.Set("appdir", appDir)
	form.Set("workers", "0")
	form.Set("restrictaccess", "0")
	// "active" omitted from properties[] => IsActive stays false, so Update()
	// won't try to spawn any real instances.

	c, w := newFormRequest(http.MethodPost, "/admin/apps/new", form, gin.Params{{Key: "appname", Value: "new"}})
	invoke(c, ctl.UpdateApp())

	if w.Code != http.StatusOK {
		body, _ := io.ReadAll(w.Body)
		t.Fatalf("expected 200, got %d, body=%s", w.Code, body)
	}
	if _, ok := appModel.apps["newapp"]; !ok {
		t.Error("expected new app to be persisted via appModel.Save")
	}
	if _, err := ctl.appServer.GetStatus("newapp"); err != nil {
		t.Errorf("expected appServer to know about the new app: %v", err)
	}
}

func TestAppControllerUpdateAppMissingRequiredFieldFails(t *testing.T) {
	ctl, _ := newTestAppController(t)

	form := url.Values{}
	form.Set("appname", "newapp")
	// "path" is required and intentionally omitted.

	c, w := newFormRequest(http.MethodPost, "/admin/apps/new", form, gin.Params{{Key: "appname", Value: "new"}})
	invoke(c, ctl.UpdateApp())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing required field, got %d", w.Code)
	}
}

func TestAppControllerUpdateAppInvalidSourceDirFails(t *testing.T) {
	ctl, appModel := newTestAppController(t)

	form := url.Values{}
	form.Set("appname", "newapp")
	form.Set("path", "/newapp")
	form.Set("appsource", "directory")
	form.Set("appdir", "/does/not/exist")
	form.Set("workers", "0")

	c, w := newFormRequest(http.MethodPost, "/admin/apps/new", form, gin.Params{{Key: "appname", Value: "new"}})
	invoke(c, ctl.UpdateApp())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid app source directory, got %d", w.Code)
	}
	if _, ok := appModel.apps["newapp"]; ok {
		t.Error("expected app not to be persisted when source validation fails")
	}
}

func TestAppControllerDeleteApp(t *testing.T) {
	appDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(appDir, "app.R"), []byte("# app"), 0644); err != nil {
		t.Fatal(err)
	}
	existing := models.App{Name: "myapp", Path: "/myapp", AppSource: "directory", AppDir: appDir, IsActive: false}
	ctl, appModel := newTestAppController(t, existing)

	c, w := newTestContext(http.MethodGet, "/admin/apps/myapp/delete", nil)
	c.Params = gin.Params{{Key: "appname", Value: "myapp"}}
	invoke(c, ctl.DeleteApp())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, ok := appModel.apps["myapp"]; ok {
		t.Error("expected app to be removed from the model")
	}
}

func TestAppControllerDeleteAppModelError(t *testing.T) {
	ctl, appModel := newTestAppController(t)
	appModel.delErr = errTestDeleteFailed

	c, w := newTestContext(http.MethodGet, "/admin/apps/myapp/delete", nil)
	c.Params = gin.Params{{Key: "appname", Value: "myapp"}}
	invoke(c, ctl.DeleteApp())

	// Current behavior: even a failed delete responds 200 (differentiated
	// only by an errorMessage in the body, not by status code).
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
