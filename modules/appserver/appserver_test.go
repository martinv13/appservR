package appserver

import (
	"testing"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/ssehandler"
)

// fakeAppModel is a minimal in-memory models.AppModel, since AppServer only
// calls All() at construction time; the CRUD methods below are unused by
// AppServer itself (they exist for controllers) and are left as no-ops.
type fakeAppModel struct {
	apps []models.App
}

func (f *fakeAppModel) All() ([]models.App, error) { return f.apps, nil }
func (f *fakeAppModel) Find(name string) (models.App, error) {
	for _, a := range f.apps {
		if a.Name == name {
			return a, nil
		}
	}
	return models.App{}, nil
}
func (f *fakeAppModel) Save(app models.App, oldName string) error { return nil }
func (f *fakeAppModel) Delete(name string) error                 { return nil }
func (f *fakeAppModel) AsMap(app models.App) (map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeAppModel) AsMapSlice(apps []models.App) ([]map[string]interface{}, error) {
	return nil, nil
}

func inactiveApp(name, path string) models.App {
	return models.App{
		Name:      name,
		Path:      path,
		AppSource: "directory",
		AppDir:    ".",
		IsActive:  false,
	}
}

func newTestAppServer(t testing.TB, apps ...models.App) *AppServer {
	t.Helper()
	model := &fakeAppModel{apps: apps}
	broker := ssehandler.NewMessageBroker()
	s, err := NewAppServer(model, broker, newMockConfig(t))
	if err != nil {
		t.Fatalf("NewAppServer failed: %v", err)
	}
	return s
}

func TestNewAppServerIndexesAppsByNameAndPath(t *testing.T) {
	s := newTestAppServer(t, inactiveApp("root", "/"), inactiveApp("myapp", "/myapp"))

	if _, ok := s.appsByName["root"]; !ok {
		t.Error("expected 'root' app to be indexed by name")
	}
	if _, ok := s.appsByName["myapp"]; !ok {
		t.Error("expected 'myapp' app to be indexed by name")
	}
	if len(s.byPath) != 2 {
		t.Fatalf("expected 2 apps in byPath, got %d", len(s.byPath))
	}
}

func TestAppServerPrefixSortOrdersPrefixPathsBeforeExtensions(t *testing.T) {
	s := newTestAppServer(t,
		inactiveApp("private", "/myapp/private"),
		inactiveApp("myapp", "/myapp"),
		inactiveApp("root", "/"),
	)

	// prefixSort orders each pair so that a path which is a prefix of another
	// sorts first: "/" before "/myapp" before "/myapp/private". Note this
	// ordering doesn't currently affect GetApp's routing correctness (it
	// matches on exact trimmed-path equality, not first-prefix-wins scan;
	// sub-paths are resolved via the appservr_appid cookie instead) — this
	// test just pins down prefixSort's own documented behavior.
	positions := map[string]int{}
	for i, app := range s.byPath {
		positions[app.App.Name] = i
	}
	if positions["root"] >= positions["myapp"] {
		t.Errorf("expected 'root' (/) to sort before 'myapp' (/myapp), got positions %v", positions)
	}
	if positions["myapp"] >= positions["private"] {
		t.Errorf("expected 'myapp' (/myapp) to sort before 'private' (/myapp/private), got positions %v", positions)
	}
}

func TestAppServerUpdateAddsNewApp(t *testing.T) {
	s := newTestAppServer(t)

	newApp := inactiveApp("newapp", "/newapp")
	if err := s.Update("newapp", newApp); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if _, ok := s.appsByName["newapp"]; !ok {
		t.Error("expected new app to be added to appsByName")
	}
	if len(s.byPath) != 1 {
		t.Errorf("expected 1 app in byPath after adding, got %d", len(s.byPath))
	}
}

func TestAppServerUpdateRenamesExistingApp(t *testing.T) {
	s := newTestAppServer(t, inactiveApp("oldname", "/oldname"))

	renamed := inactiveApp("newname", "/oldname")
	if err := s.Update("oldname", renamed); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if _, ok := s.appsByName["oldname"]; ok {
		t.Error("expected old app name to be removed from appsByName")
	}
	if _, ok := s.appsByName["newname"]; !ok {
		t.Error("expected new app name to be present in appsByName")
	}
}

func TestAppServerDeleteRemovesApp(t *testing.T) {
	s := newTestAppServer(t, inactiveApp("app1", "/app1"), inactiveApp("app2", "/app2"))

	if err := s.Delete("app1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, ok := s.appsByName["app1"]; ok {
		t.Error("expected app1 to be removed from appsByName")
	}
	if len(s.byPath) != 1 {
		t.Errorf("expected 1 remaining app in byPath, got %d", len(s.byPath))
	}
}

func TestAppServerDeleteUnknownAppErrors(t *testing.T) {
	s := newTestAppServer(t)
	if err := s.Delete("does-not-exist"); err == nil {
		t.Error("expected error deleting an unknown app")
	}
}

func TestAppServerGetStatusUnknownAppErrors(t *testing.T) {
	s := newTestAppServer(t)
	if _, err := s.GetStatus("does-not-exist"); err == nil {
		t.Error("expected error getting status of an unknown app")
	}
}

func TestAppServerGetAllStatusIncludesEveryApp(t *testing.T) {
	s := newTestAppServer(t, inactiveApp("app1", "/app1"), inactiveApp("app2", "/app2"))

	all := s.GetAllStatus()
	if _, ok := all["app1"]; !ok {
		t.Error("expected app1 in GetAllStatus result")
	}
	if _, ok := all["app2"]; !ok {
		t.Error("expected app2 in GetAllStatus result")
	}
}
