package appsource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/config"
)

type mockConfig struct {
	executableFolder string
	logger           config.Logger
}

func (c *mockConfig) ExecutableFolder() string { return c.executableFolder }
func (c *mockConfig) GetString(string) string  { return "" }
func (c *mockConfig) Logger() *config.Logger {
	return &c.logger
}

func newMockConfig(dir string) *mockConfig {
	return &mockConfig{executableFolder: dir, logger: config.NewLogger(0)}
}

func TestNewAppSourceDirWithAppR(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.R"), []byte("# app"), 0644); err != nil {
		t.Fatal(err)
	}

	app := models.App{AppDir: dir}
	src := NewAppSourceDir(app, newMockConfig(dir))
	if src.Error() != nil {
		t.Fatalf("expected no error, got %v", src.Error())
	}
	if src.Path() != dir {
		t.Errorf("expected path %q, got %q", dir, src.Path())
	}
}

func TestNewAppSourceDirWithServerAndUiR(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "server.R"), []byte("# server"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ui.R"), []byte("# ui"), 0644); err != nil {
		t.Fatal(err)
	}

	app := models.App{AppDir: dir}
	src := NewAppSourceDir(app, newMockConfig(dir))
	if src.Error() != nil {
		t.Fatalf("expected no error, got %v", src.Error())
	}
}

func TestNewAppSourceDirMissingDirectory(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist")

	app := models.App{AppDir: missing}
	src := NewAppSourceDir(app, newMockConfig(dir))
	if src.Error() == nil {
		t.Fatal("expected error for missing app directory")
	}
	if src.Path() != "" {
		t.Errorf("expected empty path on error, got %q", src.Path())
	}
}

func TestNewAppSourceDirMissingSourceFiles(t *testing.T) {
	dir := t.TempDir()
	// directory exists but contains none of app.R / server.R+ui.R

	app := models.App{AppDir: dir}
	src := NewAppSourceDir(app, newMockConfig(dir))
	if src.Error() == nil {
		t.Fatal("expected error when app.R and server.R/ui.R are all missing")
	}
}

func TestNewAppSourceDirOnlyServerRIsNotEnough(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "server.R"), []byte("# server"), 0644); err != nil {
		t.Fatal(err)
	}

	app := models.App{AppDir: dir}
	src := NewAppSourceDir(app, newMockConfig(dir))
	if src.Error() == nil {
		t.Fatal("expected error when ui.R is missing alongside server.R")
	}
}

func TestNewAppSourceDirRelativePathJoinedWithExecutableFolder(t *testing.T) {
	base := t.TempDir()
	rel := "myapp"
	if err := os.Mkdir(filepath.Join(base, rel), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, rel, "app.R"), []byte("# app"), 0644); err != nil {
		t.Fatal(err)
	}

	app := models.App{AppDir: rel}
	src := NewAppSourceDir(app, newMockConfig(base))
	if src.Error() != nil {
		t.Fatalf("expected no error, got %v", src.Error())
	}
	if src.Path() != filepath.Join(base, rel) {
		t.Errorf("expected relative path resolved against executable folder, got %q", src.Path())
	}
}

func TestNewAppSourceSampleAppCreatesAppR(t *testing.T) {
	dir := t.TempDir()
	app := models.App{AppSource: "sample-app"}

	src := NewAppSourceSampleApp(app, newMockConfig(dir))
	if src.Error() != nil {
		t.Fatalf("expected no error, got %v", src.Error())
	}

	content, err := os.ReadFile(filepath.Join(dir, "apps", "sample-app", "app.R"))
	if err != nil {
		t.Fatalf("expected app.R to be created: %v", err)
	}
	if string(content) != sampleApp {
		t.Error("expected generated app.R to match bundled sample app source")
	}
}

func TestNewAppSourceSampleAppDoesNotOverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "apps", "sample-app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}
	custom := "# custom user edits, should be preserved"
	if err := os.WriteFile(filepath.Join(appDir, "app.R"), []byte(custom), 0644); err != nil {
		t.Fatal(err)
	}

	app := models.App{AppSource: "sample-app"}
	src := NewAppSourceSampleApp(app, newMockConfig(dir))
	if src.Error() != nil {
		t.Fatalf("expected no error, got %v", src.Error())
	}

	content, err := os.ReadFile(filepath.Join(appDir, "app.R"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != custom {
		t.Error("expected existing app.R to remain untouched, but it was overwritten")
	}
}

func TestNewAppSourceDispatchesOnAppSourceField(t *testing.T) {
	dir := t.TempDir()
	conf := newMockConfig(dir)

	if err := os.WriteFile(filepath.Join(dir, "app.R"), []byte("# app"), 0644); err != nil {
		t.Fatal(err)
	}

	sample := NewAppSource(models.App{AppSource: "sample-app"}, conf, false)
	if sample == nil {
		t.Fatal("expected non-nil AppSource for 'sample-app'")
	}
	if sample.Error() != nil {
		t.Fatalf("expected sample-app source to succeed, got %v", sample.Error())
	}

	directory := NewAppSource(models.App{AppSource: "directory", AppDir: dir}, conf, false)
	if directory == nil {
		t.Fatal("expected non-nil AppSource for 'directory'")
	}
	if directory.Error() != nil {
		t.Fatalf("expected directory source to succeed, got %v", directory.Error())
	}

	other := NewAppSource(models.App{AppSource: "something-else"}, conf, false)
	if other != nil {
		t.Errorf("expected nil AppSource for unrecognized AppSource value, got %v", other)
	}
}
