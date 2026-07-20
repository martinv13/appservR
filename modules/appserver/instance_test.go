package appserver

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInstanceStartBecomesRunningAndServesHTTP(t *testing.T) {
	conf := newMockConfig(t)
	appDir := newTestAppDir(t)

	inst := NewInstance("testapp", appDir, conf)
	if err := inst.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer inst.Stop()

	waitForStatus(t, inst, instStatus.RUNNING, 5*time.Second)

	resp, err := http.Get("http://127.0.0.1:" + inst.Port() + "/")
	if err != nil {
		t.Fatalf("failed to reach mock instance: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "mock-shiny-port="+inst.Port()) {
		t.Errorf("expected response to identify port %s, got %q", inst.Port(), body)
	}
}

func TestInstanceStartFailsWhenAppDirMissing(t *testing.T) {
	conf := newMockConfig(t)

	inst := NewInstance("testapp", filepath.Join(t.TempDir(), "does-not-exist"), conf)
	err := inst.Start()
	if err == nil {
		t.Fatal("expected error starting instance with missing app directory")
	}
	if inst.Status() != instStatus.ERROR {
		t.Errorf("expected status ERROR, got %s", inst.Status())
	}
}

func TestInstanceStopPreventsAutoRestart(t *testing.T) {
	conf := newMockConfig(t)
	appDir := newTestAppDir(t)

	inst := NewInstance("testapp", appDir, conf)
	if err := inst.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	waitForStatus(t, inst, instStatus.RUNNING, 5*time.Second)

	if err := inst.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	waitForStatus(t, inst, instStatus.STOPPED, 3*time.Second)

	// Auto-restart backoff starts at 1s; if Stop() didn't suppress it, the
	// instance would flip back to STARTING/RUNNING on its own.
	time.Sleep(1500 * time.Millisecond)
	if got := inst.Status(); got != instStatus.STOPPED {
		t.Errorf("expected instance to remain STOPPED after explicit Stop, got %s", got)
	}
}

func TestInstanceAutoRestartsAfterCrash(t *testing.T) {
	conf := newMockConfig(t)
	appDir := newTestAppDir(t)
	if err := os.WriteFile(filepath.Join(appDir, "crash_after_ms"), []byte("150"), 0644); err != nil {
		t.Fatal(err)
	}

	inst := NewInstance("testapp", appDir, conf)
	if err := inst.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer inst.Stop()

	waitForStatus(t, inst, instStatus.RUNNING, 5*time.Second)
	firstPort := inst.Port()

	// After crashing, Start() is retried (with a new port, since crash exits
	// don't release the old one back to the pool) after a backoff delay.
	deadline := time.Now().Add(6 * time.Second)
	restarted := false
	for time.Now().Before(deadline) {
		if inst.Status() == instStatus.RUNNING && inst.Port() != firstPort {
			restarted = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !restarted {
		t.Fatalf("expected instance to auto-restart on a new port after crashing; last status=%s port=%s", inst.Status(), inst.Port())
	}
}

func TestInstanceUserCount(t *testing.T) {
	conf := newMockConfig(t)
	inst := NewInstance("testapp", newTestAppDir(t), conf)

	if inst.UserCount() != 0 {
		t.Fatalf("expected initial user count 0, got %d", inst.UserCount())
	}
	inst.SetUserCount(1, true)
	inst.SetUserCount(1, true)
	if got := inst.UserCount(); got != 2 {
		t.Errorf("expected user count 2, got %d", got)
	}
	inst.SetUserCount(-1, true)
	if got := inst.UserCount(); got != 1 {
		t.Errorf("expected user count 1, got %d", got)
	}
	inst.SetUserCount(-5, true)
	if got := inst.UserCount(); got != 0 {
		t.Errorf("expected user count to floor at 0, got %d", got)
	}
	inst.SetUserCount(3, false)
	if got := inst.UserCount(); got != 3 {
		t.Errorf("expected absolute user count 3, got %d", got)
	}
}
