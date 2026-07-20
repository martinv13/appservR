package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/auth"
)

// adminCookie returns a valid auth cookie for a logged-in admin user, for
// hitting routes behind middlewares.AdminAuth().
func adminCookie() *http.Cookie {
	user := models.User{Username: "admin", DisplayedName: "Admin", Groups: []models.Group{{Name: "admins"}}}
	return &http.Cookie{Name: "token", Value: auth.GenerateToken(user)}
}

func doFormRequest(router *AppRouter, method, path string, form url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	router.router.ServeHTTP(w, req)
	return w
}

// TestAdminUICanCreateViewAndDeleteApp drives the real admin app pages
// (app.html / apps.html, with real controller-produced data) end to end:
// new-app form -> create -> appears in the list -> detail page renders ->
// delete -> gone from the list. This is the same flow a person clicking
// through the admin UI would exercise, minus the browser and JS.
func TestAdminUICanCreateViewAndDeleteApp(t *testing.T) {
	router := newTestRouter(t)
	admin := adminCookie()

	appDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(appDir, "app.R"), []byte("# app"), 0644); err != nil {
		t.Fatal(err)
	}

	// New-app form renders.
	w := doRequest(router, http.MethodGet, "/admin/apps/new", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for new-app form, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "New App") {
		t.Errorf("expected new-app form to render its title, got body=%s", w.Body.String())
	}

	// Create the app.
	form := url.Values{}
	form.Set("appname", "myapp")
	form.Set("path", "/myapp")
	form.Set("appsource", "directory")
	form.Set("appdir", appDir)
	form.Set("workers", "0")
	form.Set("restrictaccess", "0")
	w = doFormRequest(router, http.MethodPost, "/admin/apps/new", form, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 creating app, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "successfuly") {
		t.Errorf("expected success message after creating app, got body=%s", w.Body.String())
	}

	// It shows up in the apps list.
	w = doRequest(router, http.MethodGet, "/admin/apps", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for apps list, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "myapp") {
		t.Errorf("expected apps list to contain the new app, got body=%s", w.Body.String())
	}

	// Its detail page renders (exercises the Status/StdErr section too).
	w = doRequest(router, http.MethodGet, "/admin/apps/myapp", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for app detail page, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `value="myapp"`) {
		t.Errorf("expected app detail form to be pre-filled with the app name, got body=%s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Console output") {
		t.Errorf("expected app detail page to render the console output section for an existing app, got body=%s", w.Body.String())
	}

	// Delete it.
	w = doRequest(router, http.MethodGet, "/admin/apps/myapp/delete", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 deleting app, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "myapp") {
		t.Errorf("expected delete confirmation message to name the app, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/apps", admin)
	if strings.Contains(w.Body.String(), `id="card-myapp"`) {
		t.Errorf("expected deleted app to be gone from the apps list, got body=%s", w.Body.String())
	}
}

// TestAdminUICanCreateViewAndDeleteUser drives the real admin user pages.
func TestAdminUICanCreateViewAndDeleteUser(t *testing.T) {
	router := newTestRouter(t)
	admin := adminCookie()

	w := doRequest(router, http.MethodGet, "/admin/users/new", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for new-user form, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "New user") {
		t.Errorf("expected new-user form to render its title, got body=%s", w.Body.String())
	}

	form := url.Values{}
	form.Set("username", "bob")
	form.Set("displayedname", "Bob Bobson")
	form.Set("password", "secret")
	w = doFormRequest(router, http.MethodPost, "/admin/users/new", form, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 creating user, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "has been updated") {
		t.Errorf("expected success message after creating user, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/users", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for users list, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Bob Bobson") {
		t.Errorf("expected users list to contain the new user, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/users/bob", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for user detail page, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `value="Bob Bobson"`) {
		t.Errorf("expected user detail form to be pre-filled with the displayed name, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/users/bob/delete", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 deleting user, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "has been deleted") {
		t.Errorf("expected delete confirmation message, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/users", admin)
	if strings.Contains(w.Body.String(), "Bob Bobson") {
		t.Errorf("expected deleted user to be gone from the users list, got body=%s", w.Body.String())
	}
}

// TestAdminUICanCreateViewAndDeleteGroup drives the real admin group pages.
func TestAdminUICanCreateViewAndDeleteGroup(t *testing.T) {
	router := newTestRouter(t)
	admin := adminCookie()

	w := doRequest(router, http.MethodGet, "/admin/groups/new", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for new-group form, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "New group") {
		t.Errorf("expected new-group form to render its title, got body=%s", w.Body.String())
	}

	form := url.Values{}
	form.Set("groupname", "editors")
	w = doFormRequest(router, http.MethodPost, "/admin/groups/new", form, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 creating group, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "has been created") {
		t.Errorf("expected success message after creating group, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/groups", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for groups list, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "editors") {
		t.Errorf("expected groups list to contain the new group, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/groups/editors", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for group detail page, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Group: editors") {
		t.Errorf("expected group detail page to show the group name, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/groups/editors/delete", admin)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 deleting group, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "has been deleted") {
		t.Errorf("expected delete confirmation message, got body=%s", w.Body.String())
	}

	w = doRequest(router, http.MethodGet, "/admin/groups", admin)
	if strings.Contains(w.Body.String(), `>editors<`) {
		t.Errorf("expected deleted group to be gone from the groups list, got body=%s", w.Body.String())
	}
}
