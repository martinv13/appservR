package appserver

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/appservR/appservR/modules/config"
)

var (
	mockShinyOnce sync.Once
	mockShinyPath string
	mockShinyErr  error
)

// buildMockShiny compiles the mockshiny test helper (testdata/mockshiny) once
// per test binary run, and returns the path to the resulting executable. It
// stands in for a real Rscript so Instance/AppProxy logic can be exercised
// without an R installation.
func buildMockShiny(t testing.TB) string {
	t.Helper()
	mockShinyOnce.Do(func() {
		dir, err := os.MkdirTemp("", "appservr-mockshiny")
		if err != nil {
			mockShinyErr = err
			return
		}
		out := filepath.Join(dir, "mockshiny")
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", out, "github.com/appservR/appservR/modules/appserver/testdata/mockshiny")
		cmd.Dir = mustRepoRoot()
		if output, err := cmd.CombinedOutput(); err != nil {
			mockShinyErr = err
			mockShinyPath = string(output)
			return
		}
		mockShinyPath = out
	})
	if mockShinyErr != nil {
		t.Fatalf("failed to build mockshiny helper: %v\n%s", mockShinyErr, mockShinyPath)
	}
	return mockShinyPath
}

func mustRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(wd, "..", "..")
}

// mockConfig implements config.Config for tests, pointing "Rscript" at the
// compiled mockshiny helper.
type mockConfig struct {
	rscript string
	logger  config.Logger
}

func newMockConfig(t testing.TB) *mockConfig {
	return &mockConfig{
		rscript: buildMockShiny(t),
		// level 3 = errors only, to keep test output free of Instance's Info logs.
		logger: config.NewLogger(3),
	}
}

func (c *mockConfig) ExecutableFolder() string   { return "." }
func (c *mockConfig) Logger() *config.Logger     { return &c.logger }
func (c *mockConfig) GetString(key string) string {
	if key == "Rscript" {
		return c.rscript
	}
	return ""
}

// newTestAppDir creates a temp directory containing a dummy app.R so
// appsource validation succeeds, and returns its path.
func newTestAppDir(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.R"), []byte("# mock app\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// waitForStatus polls an Instance's Status() until it matches want or the
// timeout elapses.
func waitForStatus(t testing.TB, inst *Instance, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if inst.Status() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out after %s waiting for status %q, got %q (stderr: %s)", timeout, want, inst.Status(), inst.StdErr())
}
