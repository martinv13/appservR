package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestAccessLevelsAreDistinct(t *testing.T) {
	levels := []int{AccessLevels.PUBLIC, AccessLevels.ALL_USERS, AccessLevels.SPECIFIC_GROUPS}
	seen := map[int]bool{}
	for _, l := range levels {
		if seen[l] {
			t.Errorf("expected distinct AccessLevels values, got duplicate %d", l)
		}
		seen[l] = true
	}
}

func TestConfigViperAccessors(t *testing.T) {
	v := viper.New()
	v.Set("server.port", "9090")
	c := &ConfigViper{
		executableFolder: "/some/folder",
		v:                v,
		logger:           NewLogger(logLevels.WARNING),
	}

	if got := c.ExecutableFolder(); got != "/some/folder" {
		t.Errorf("expected ExecutableFolder %q, got %q", "/some/folder", got)
	}
	if got := c.GetString("server.port"); got != "9090" {
		t.Errorf("expected GetString(server.port) %q, got %q", "9090", got)
	}
	if got := c.GetString("does.not.exist"); got != "" {
		t.Errorf("expected empty string for unknown key, got %q", got)
	}
	if c.Logger() == nil {
		t.Error("expected non-nil Logger")
	}
}

// withTempWorkingDir temporarily chdirs into a fresh temp directory for the
// duration of fn, restoring the original working directory afterward. Uses
// plain os.Chdir rather than testing.T.Chdir (Go 1.24+) since this repo's CI
// matrix pins an older Go toolchain.
func withTempWorkingDir(t *testing.T, fn func(dir string)) {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatal(err)
		}
	}()
	fn(dir)
}

func TestNewConfigViperWritesDefaultConfigOnFirstRun(t *testing.T) {
	withTempWorkingDir(t, func(dir string) {
		c, err := NewConfigViper(RunFlags{})
		if err != nil {
			t.Fatalf("NewConfigViper failed: %v", err)
		}

		if got := c.GetString("server.port"); got != "8080" {
			t.Errorf("expected default server.port 8080, got %q", got)
		}
		if got := c.GetString("server.host"); got != "localhost" {
			t.Errorf("expected default server.host localhost, got %q", got)
		}
		if got := c.GetString("mode"); got != "prod" {
			t.Errorf("expected default mode prod, got %q", got)
		}

		if _, err := os.Stat(filepath.Join(dir, "config.yml")); err != nil {
			t.Errorf("expected config.yml to be written on first run: %v", err)
		}
	})
}

func TestNewConfigViperRunFlagsOverrideDefaults(t *testing.T) {
	withTempWorkingDir(t, func(dir string) {
		c, err := NewConfigViper(RunFlags{Address: "0.0.0.0", Port: "1234", Mode: "debug"})
		if err != nil {
			t.Fatalf("NewConfigViper failed: %v", err)
		}

		if got := c.GetString("server.host"); got != "0.0.0.0" {
			t.Errorf("expected server.host overridden to 0.0.0.0, got %q", got)
		}
		if got := c.GetString("server.port"); got != "1234" {
			t.Errorf("expected server.port overridden to 1234, got %q", got)
		}
		if got := c.GetString("mode"); got != "debug" {
			t.Errorf("expected mode overridden to debug, got %q", got)
		}
	})
}

func TestNewConfigViperReadsExistingConfigFile(t *testing.T) {
	withTempWorkingDir(t, func(dir string) {
		content := "server:\n  port: 9999\n  host: myhost\n"
		if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		c, err := NewConfigViper(RunFlags{})
		if err != nil {
			t.Fatalf("NewConfigViper failed: %v", err)
		}
		if got := c.GetString("server.port"); got != "9999" {
			t.Errorf("expected server.port from existing config file (9999), got %q", got)
		}
		if got := c.GetString("server.host"); got != "myhost" {
			t.Errorf("expected server.host from existing config file (myhost), got %q", got)
		}
	})
}
