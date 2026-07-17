package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/appservR/appservR/controllers"
	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/appserver"
	"github.com/appservR/appservR/modules/auth"
	"github.com/appservR/appservR/modules/config"
	"github.com/appservR/appservR/modules/ssehandler"
	"github.com/appservR/appservR/modules/vfsdata"
	"github.com/gin-gonic/gin"
)

// fakeConfig is a minimal config.Config for router tests. It never spawns a
// real Rscript process: tests delete the auto-seeded "sample-app" before
// building the AppServer, so no Instance ever needs to start.
type fakeConfig struct {
	mode       string
	execFolder string
	dbName     string
	logger     config.Logger
}

func (c *fakeConfig) ExecutableFolder() string { return c.execFolder }
func (c *fakeConfig) GetString(key string) string {
	if key == "mode" {
		return c.mode
	}
	return ""
}
func (c *fakeConfig) Logger() *config.Logger { return &c.logger }

func newFakeConfig(t testing.TB) *fakeConfig {
	return &fakeConfig{mode: "debug", execFolder: t.TempDir(), dbName: t.Name(), logger: config.NewLogger(3)}
}

// repoStaticPaths points a StaticPaths directly at this repo's real
// templates/ and assets/ directories (absolute paths), sidestepping the
// vfsgen-generated placeholders in modules/vfsdata (which only resolve
// correctly after `go generate ./...` has bundled them, and even then
// default to CWD-relative lookups before that).
func repoStaticPaths(t *testing.T) *vfsdata.StaticPaths {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(wd)
	templatesDir := filepath.Join(root, "templates")
	assetsDir := filepath.Join(root, "assets")
	if _, err := os.Stat(templatesDir); err != nil {
		t.Fatalf("expected repo templates dir at %s: %v", templatesDir, err)
	}
	return &vfsdata.StaticPaths{
		Templates: vfsdata.HybridFileSystem{
			LocalFS:   http.Dir(templatesDir),
			BundledFS: http.Dir(templatesDir),
		},
		Assets: vfsdata.HybridFileSystem{
			LocalFS:   http.Dir(assetsDir),
			BundledFS: http.Dir(assetsDir),
		},
	}
}

func newTestRouter(t *testing.T) *AppRouter {
	t.Helper()
	gin.SetMode(gin.TestMode)

	conf := newFakeConfig(t)
	db, err := models.NewDB(&testDBConfig{conf})
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	groupModel := models.NewGroupModelDB(db)
	appModel, err := models.NewAppModelDB(db, groupModel)
	if err != nil {
		t.Fatalf("NewAppModelDB failed: %v", err)
	}
	// Drop the auto-seeded default app so AppServer starts with zero apps -
	// this test is about route wiring/access control, not process
	// management (already covered by modules/appserver's own tests).
	if err := appModel.Delete("sample-app"); err != nil {
		t.Fatalf("failed to remove seeded sample-app: %v", err)
	}
	userModel := models.NewUserModelDB(db, groupModel)

	msgBroker := ssehandler.NewMessageBroker()
	appServer, err := appserver.NewAppServer(appModel, msgBroker, conf)
	if err != nil {
		t.Fatalf("NewAppServer failed: %v", err)
	}

	appsCtl := controllers.NewAppController(appModel, appServer, conf)
	usersCtl := controllers.NewUserController(userModel)
	groupsCtl := controllers.NewGroupController(groupModel)
	authCtl := controllers.NewAuthController(userModel)

	staticPaths := repoStaticPaths(t)

	router, err := NewAppRouter(conf, staticPaths, appServer, msgBroker, appsCtl, usersCtl, groupsCtl, authCtl)
	if err != nil {
		t.Fatalf("NewAppRouter failed: %v", err)
	}
	return router
}

// testDBConfig adapts a config.Config to also satisfy models.NewDB's needs
// with an isolated in-memory sqlite DB unique to this test.
type testDBConfig struct {
	*fakeConfig
}

func (c *testDBConfig) GetString(key string) string {
	switch key {
	case "database.type":
		return "sqlite"
	case "database.path":
		return "file:" + c.dbName + "?mode=memory&cache=shared"
	case "mode":
		return c.mode
	}
	return ""
}

func doRequest(router *AppRouter, method, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	router.router.ServeHTTP(w, req)
	return w
}

func TestRouterServesLoginPage(t *testing.T) {
	router := newTestRouter(t)
	w := doRequest(router, http.MethodGet, "/auth/login")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for login page, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestRouterServesStaticAssets(t *testing.T) {
	router := newTestRouter(t)
	w := doRequest(router, http.MethodGet, "/assets/css/bootstrap.min.css")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for static asset, got %d", w.Code)
	}
}

func TestRouterAdminRoutesRejectAnonymous(t *testing.T) {
	router := newTestRouter(t)
	w := doRequest(router, http.MethodGet, "/admin/apps")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for anonymous admin access, got %d", w.Code)
	}
}

func TestRouterAdminRoutesAllowAdminUser(t *testing.T) {
	router := newTestRouter(t)
	user := models.User{
		Username:      "admin",
		DisplayedName: "Admin",
		Groups:        []models.Group{{Name: "admins"}},
	}
	token := auth.GenerateToken(user)
	w := doRequest(router, http.MethodGet, "/admin/apps", &http.Cookie{Name: "token", Value: token})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for admin user, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestRouterAdminRoutesRejectNonAdminUser(t *testing.T) {
	router := newTestRouter(t)
	user := models.User{
		Username:      "bob",
		DisplayedName: "Bob",
		Groups:        []models.Group{{Name: "viewers"}},
	}
	token := auth.GenerateToken(user)
	w := doRequest(router, http.MethodGet, "/admin/apps", &http.Cookie{Name: "token", Value: token})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-admin user, got %d", w.Code)
	}
}

func TestRouterCatchAllReturns404ForUnknownApp(t *testing.T) {
	router := newTestRouter(t)
	w := doRequest(router, http.MethodGet, "/no-such-app")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unmatched app path, got %d", w.Code)
	}
}
