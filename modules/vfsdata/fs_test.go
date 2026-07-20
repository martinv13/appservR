package vfsdata

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/appservR/appservR/modules/config"
)

type fakeConfig struct {
	execFolder string
}

func (c *fakeConfig) ExecutableFolder() string { return c.execFolder }
func (c *fakeConfig) GetString(string) string  { return "" }
func (c *fakeConfig) Logger() *config.Logger {
	l := config.NewLogger(3)
	return &l
}

func mustReadFile(t *testing.T, f http.File) string {
	t.Helper()
	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(buf[:n])
}

func TestHybridFileSystemPrefersLocalOverBundled(t *testing.T) {
	localDir := t.TempDir()
	bundledDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(localDir, "shared.txt"), []byte("local content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundledDir, "shared.txt"), []byte("bundled content"), 0644); err != nil {
		t.Fatal(err)
	}

	hfs := HybridFileSystem{
		LocalFS:   http.Dir(localDir),
		BundledFS: http.Dir(bundledDir),
	}

	f, err := hfs.Open("/shared.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()
	if got := mustReadFile(t, f); got != "local content" {
		t.Errorf("expected local content to take precedence, got %q", got)
	}
}

func TestHybridFileSystemFallsBackToBundled(t *testing.T) {
	localDir := t.TempDir()
	bundledDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(bundledDir, "only-bundled.txt"), []byte("bundled only"), 0644); err != nil {
		t.Fatal(err)
	}

	hfs := HybridFileSystem{
		LocalFS:   http.Dir(localDir),
		BundledFS: http.Dir(bundledDir),
	}

	f, err := hfs.Open("/only-bundled.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()
	if got := mustReadFile(t, f); got != "bundled only" {
		t.Errorf("expected fallback to bundled content, got %q", got)
	}
}

func TestHybridFileSystemErrorsWhenMissingFromBoth(t *testing.T) {
	hfs := HybridFileSystem{
		LocalFS:   http.Dir(t.TempDir()),
		BundledFS: http.Dir(t.TempDir()),
	}

	if _, err := hfs.Open("/does-not-exist.txt"); err == nil {
		t.Error("expected an error when the file exists in neither filesystem")
	}
}

func TestNewStaticPathsWiresExecutableFolder(t *testing.T) {
	execFolder := t.TempDir()
	if err := os.MkdirAll(filepath.Join(execFolder, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(execFolder, "assets", "custom.js"), []byte("custom asset"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(execFolder, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(execFolder, "templates", "custom.html"), []byte("custom template"), 0644); err != nil {
		t.Fatal(err)
	}

	sp := NewStaticPaths(&fakeConfig{execFolder: execFolder})

	f, err := sp.Assets.Open("/custom.js")
	if err != nil {
		t.Fatalf("expected to open the local asset override, got error: %v", err)
	}
	defer f.Close()
	if got := mustReadFile(t, f); got != "custom asset" {
		t.Errorf("expected local asset content, got %q", got)
	}

	f2, err := sp.Templates.Open("/custom.html")
	if err != nil {
		t.Fatalf("expected to open the local template override, got error: %v", err)
	}
	defer f2.Close()
	if got := mustReadFile(t, f2); got != "custom template" {
		t.Errorf("expected local template content, got %q", got)
	}
}
