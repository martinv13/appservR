package models

import (
	"errors"
	"testing"

	"github.com/appservR/appservR/modules/config"
	"gorm.io/gorm"
)

// setUpNamed opens an isolated in-memory sqlite DB keyed by a unique name,
// so each test function gets its own database rather than sharing the
// process-wide "file::memory:?cache=shared" instance that setUp() (in
// models_test.go) uses for the single TestDataModelDB.
func setUpNamed(name string) (*gorm.DB, error) {
	conf := &MockConfig{
		keys: map[string]string{
			"database.type": "sqlite",
			"database.path": "file:" + name + "?mode=memory&cache=shared",
		},
		logger: config.NewLogger(3),
	}
	db, err := NewDB(conf)
	if err != nil {
		return nil, errors.New("unable to initialize named in-memory database")
	}
	return db, nil
}

func TestAppModelDeleteRemovesApp(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	appModel, err := NewAppModelDB(db, groupModel)
	if err != nil {
		t.Fatal(err)
	}

	// NewAppModelDB seeds a default "sample-app" when the DB is empty.
	if err := appModel.Delete("sample-app"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := appModel.Find("sample-app"); err == nil {
		t.Error("expected sample-app to be gone after Delete")
	}
	apps, err := appModel.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps after delete, got %d", len(apps))
	}
}

func TestAppModelAsMapReflectsAllowedGroupsAfterUpdate(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	appModel, err := NewAppModelDB(db, groupModel)
	if err != nil {
		t.Fatal(err)
	}
	if err := groupModel.Save(Group{Name: "editors"}, "new"); err != nil {
		t.Fatal(err)
	}

	// Create the app first with no groups (see
	// TestAppModelSaveOnCreateSilentlyDropsPreexistingAllowedGroups for why
	// specifying AllowedGroups directly on creation doesn't work), then set
	// AllowedGroups through an update, which goes through the
	// Association("AllowedGroups").Replace(groups) path.
	app := App{
		Name:           "myapp",
		Path:           "/myapp",
		AppDir:         "apps/sample-app/",
		Workers:        2,
		RestrictAccess: config.AccessLevels.SPECIFIC_GROUPS,
	}
	if err := appModel.Save(app, "new"); err != nil {
		t.Fatalf("Save (create) failed: %v", err)
	}
	app.AllowedGroups = []Group{{Name: "editors"}}
	if err := appModel.Save(app, "myapp"); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	saved, err := appModel.Find("myapp")
	if err != nil {
		t.Fatal(err)
	}
	m, err := appModel.AsMap(saved)
	if err != nil {
		t.Fatal(err)
	}
	groups, ok := m["AllowedGroups"].(map[string]bool)
	if !ok {
		t.Fatalf("expected AllowedGroups to be map[string]bool, got %T", m["AllowedGroups"])
	}
	if !groups["editors"] {
		t.Errorf("expected editors=true in AllowedGroups map, got %v", groups)
	}

	apps, err := appModel.All()
	if err != nil {
		t.Fatal(err)
	}
	slice, err := appModel.AsMapSlice(apps)
	if err != nil {
		t.Fatal(err)
	}
	if len(slice) != len(apps) {
		t.Errorf("expected AsMapSlice to return %d entries, got %d", len(apps), len(slice))
	}
}

// TestAppModelSaveOnCreateSilentlyDropsPreexistingAllowedGroups documents a
// real bug found while testing: Save's create path (oldName == "new") does
// `m.DB.Create(&app)` with the caller-supplied app.AllowedGroups verbatim,
// instead of the `groups` slice it already queried from the DB by name a few
// lines above (which is only used by the *update* path's
// Association(...).Replace(groups) call). Each Group in AllowedGroups has a
// zero ID, so gorm's auto-save-association logic tries to INSERT it; since
// `groups.name` is UNIQUE, that insert is a no-op on conflict for a group
// that already exists - but gorm then never links the join table row either,
// so the association is just dropped, silently, with no error. Net effect:
// creating a new app and assigning it to an *existing* group produces an app
// with AllowedGroups permanently empty, even though creation "succeeded".
func TestAppModelSaveOnCreateSilentlyDropsPreexistingAllowedGroups(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	appModel, err := NewAppModelDB(db, groupModel)
	if err != nil {
		t.Fatal(err)
	}
	if err := groupModel.Save(Group{Name: "editors"}, "new"); err != nil {
		t.Fatal(err)
	}

	app := App{Name: "myapp", Path: "/myapp", AppDir: "apps/sample-app/",
		AllowedGroups: []Group{{Name: "editors"}}}
	if err := appModel.Save(app, "new"); err != nil {
		t.Fatalf("Save unexpectedly failed: %v", err)
	}

	saved, err := appModel.Find("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.AllowedGroups) != 0 {
		t.Errorf("known-buggy behavior changed: expected AllowedGroups to end up empty on create, got %v", saved.AllowedGroups)
	}
}

// TestAppModelSaveOnCreateCreatesUnknownGroupsInstead documents the other
// half of the same bug: because the create path never validates
// AllowedGroups against the `groups` it queried by name, a name that doesn't
// exist yet gets created as a brand new, real Group row (and properly
// associated) - despite the code's own error message elsewhere implying
// unknown group names should be rejected.
func TestAppModelSaveOnCreateCreatesUnknownGroupsInstead(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	appModel, err := NewAppModelDB(db, groupModel)
	if err != nil {
		t.Fatal(err)
	}

	app := App{Name: "myapp", Path: "/myapp", AppDir: "apps/sample-app/",
		AllowedGroups: []Group{{Name: "typo-d-group"}}}
	if err := appModel.Save(app, "new"); err != nil {
		t.Fatalf("Save unexpectedly failed: %v", err)
	}

	saved, err := appModel.Find("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.AllowedGroups) != 1 || saved.AllowedGroups[0].Name != "typo-d-group" {
		t.Errorf("known-buggy behavior changed: expected a never-before-seen group name to be created and associated, got %v", saved.AllowedGroups)
	}
	names, err := groupModel.AllNames()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range names {
		if n == "typo-d-group" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'typo-d-group' to now exist as a real group, got %v", names)
	}
}

func TestAppModelSaveNameCannotBeNew(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	appModel, err := NewAppModelDB(db, groupModel)
	if err != nil {
		t.Fatal(err)
	}
	if err := appModel.Save(App{Name: "new"}, "new"); err == nil {
		t.Error("expected error creating an app literally named 'new'")
	}
}

func TestAppModelSaveUpdateUnknownOldNameFails(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	appModel, err := NewAppModelDB(db, groupModel)
	if err != nil {
		t.Fatal(err)
	}
	if err := appModel.Save(App{Name: "myapp", Path: "/myapp"}, "ghost"); err == nil {
		t.Error("expected error updating a nonexistent app")
	}
}

func TestGroupModelCRUDLifecycle(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)

	if err := groupModel.Save(Group{Name: "editors"}, "new"); err != nil {
		t.Fatalf("Save (create) failed: %v", err)
	}
	g, err := groupModel.Find("editors")
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if g.Name != "editors" {
		t.Errorf("expected group name 'editors', got %q", g.Name)
	}

	if err := groupModel.Save(Group{Name: "writers"}, "editors"); err != nil {
		t.Fatalf("Save (rename) failed: %v", err)
	}
	if _, err := groupModel.Find("editors"); err == nil {
		t.Error("expected old group name to be gone after rename")
	}
	if _, err := groupModel.Find("writers"); err != nil {
		t.Error("expected renamed group to be found")
	}

	names, err := groupModel.AllNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "writers" {
		t.Errorf("expected AllNames to return [writers], got %v", names)
	}

	slice, err := groupModel.AsMapSlice()
	if err != nil {
		t.Fatal(err)
	}
	if len(slice) != 1 {
		t.Errorf("expected AsMapSlice to return 1 entry, got %d", len(slice))
	}

	if err := groupModel.Delete("writers"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := groupModel.Find("writers"); err == nil {
		t.Error("expected group to be gone after Delete")
	}
}

func TestGroupModelAdminsGroupIsProtected(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)

	if err := groupModel.Save(Group{Name: "admins"}, "new"); err == nil {
		t.Error("expected error creating a group literally named 'admins'")
	}

	if err := db.Create(&Group{Name: "admins"}).Error; err != nil {
		t.Fatalf("failed to seed admins group directly: %v", err)
	}
	if err := groupModel.Save(Group{Name: "renamed"}, "admins"); err == nil {
		t.Error("expected error renaming the admins group")
	}
	if err := groupModel.Delete("admins"); err == nil {
		t.Error("expected error deleting the admins group")
	}
}

func TestGroupModelAddRemoveMemberAreNoOps(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	if err := groupModel.Save(Group{Name: "editors"}, "new"); err != nil {
		t.Fatal(err)
	}

	// AddMember/RemoveMember are currently unimplemented no-op stubs in
	// models/group.go: they always return nil without touching the
	// database. This test documents that current (likely incomplete)
	// behavior rather than assume it works.
	if err := groupModel.AddMember("editors", "alice"); err != nil {
		t.Fatalf("AddMember returned an unexpected error: %v", err)
	}
	g, err := groupModel.Find("editors")
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Users) != 0 {
		t.Errorf("expected AddMember to have no effect (known stub), but group has %d users", len(g.Users))
	}

	if err := groupModel.RemoveMember("editors", "alice"); err != nil {
		t.Fatalf("RemoveMember returned an unexpected error: %v", err)
	}
}

func TestUserModelAllAndAsMapSlice(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	userModel := NewUserModelDB(db, groupModel)

	if err := userModel.Save(User{Username: "admin", Password: "x"}, "new"); err != nil {
		t.Fatal(err)
	}
	if err := userModel.Save(User{Username: "bob", Password: "y"}, "new"); err != nil {
		t.Fatal(err)
	}

	users, err := userModel.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}

	slice, err := userModel.AsMapSlice(users)
	if err != nil {
		t.Fatal(err)
	}
	if len(slice) != 2 {
		t.Errorf("expected AsMapSlice to return 2 entries, got %d", len(slice))
	}
}

func TestUserModelAdminSaveAndDelete(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	userModel := NewUserModelDB(db, groupModel)

	if err := groupModel.Save(Group{Name: "editors"}, "new"); err != nil {
		t.Fatal(err)
	}

	newUser := User{Username: "carol", DisplayedName: "Carol", Password: "x", Groups: []Group{{Name: "editors"}}}
	if err := userModel.AdminSave(newUser, "new"); err != nil {
		t.Fatalf("AdminSave (create) failed: %v", err)
	}

	saved, err := userModel.Find("carol")
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Groups) != 1 || saved.Groups[0].Name != "editors" {
		t.Errorf("expected carol to be in the editors group, got %v", saved.Groups)
	}

	updated := User{Username: "carol2", DisplayedName: "Carol II", Groups: []Group{}}
	if err := userModel.AdminSave(updated, "carol"); err != nil {
		t.Fatalf("AdminSave (update) failed: %v", err)
	}
	if _, err := userModel.Find("carol"); err == nil {
		t.Error("expected old username to be gone after rename")
	}
	saved2, err := userModel.Find("carol2")
	if err != nil {
		t.Fatal(err)
	}
	if len(saved2.Groups) != 0 {
		t.Errorf("expected groups to be cleared, got %v", saved2.Groups)
	}

	if err := userModel.Delete("carol2"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := userModel.Find("carol2"); err == nil {
		t.Error("expected user to be gone after Delete")
	}
}

func TestUserModelAdminSaveNameCannotBeNew(t *testing.T) {
	db, err := setUpNamed(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	groupModel := NewGroupModelDB(db)
	userModel := NewUserModelDB(db, groupModel)

	if err := userModel.AdminSave(User{Username: "new"}, "new"); err == nil {
		t.Error("expected error creating a user literally named 'new'")
	}
}
