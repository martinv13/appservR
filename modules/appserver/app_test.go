package appserver

import (
	"testing"
	"time"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/config"
	"github.com/appservR/appservR/modules/ssehandler"
	"github.com/gin-gonic/gin"
)

func newTestAppProxy(t testing.TB, app models.App) *AppProxy {
	t.Helper()
	broker := ssehandler.NewMessageBroker()
	p, err := NewAppProxy(app, broker, newMockConfig(t))
	if err != nil {
		t.Fatalf("NewAppProxy failed: %v", err)
	}
	t.Cleanup(p.Cleanup)
	return p
}

func waitForRunningCount(t testing.TB, p *AppProxy, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if status := p.GetStatus(false); status["RunningInst"] == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d running instances, last status: %v", want, p.GetStatus(true))
}

func TestNewAppProxyStartsConfiguredWorkers(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   2,
		IsActive:  true,
	}
	p := newTestAppProxy(t, app)
	waitForRunningCount(t, p, 2, 5*time.Second)
}

func TestAppProxyInactiveAppStartsNoWorkers(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   2,
		IsActive:  false,
	}
	p := newTestAppProxy(t, app)

	// Give it a moment to (not) start workers, then confirm it stays at zero.
	time.Sleep(200 * time.Millisecond)
	if status := p.GetStatus(false); status["RunningInst"] != 0 {
		t.Errorf("expected 0 running instances for inactive app, got %v", status["RunningInst"])
	}
}

func TestAppProxyUpdateRescalesUpAndDown(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   1,
		IsActive:  true,
	}
	p := newTestAppProxy(t, app)
	waitForRunningCount(t, p, 1, 5*time.Second)

	app.Workers = 3
	p.Update(app)
	waitForRunningCount(t, p, 3, 5*time.Second)

	app.Workers = 1
	p.Update(app)
	waitForRunningCount(t, p, 1, 5*time.Second)
}

func TestAppProxyRescaleDownStopsIdleInstanceFirst(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   2,
		IsActive:  true,
	}
	p := newTestAppProxy(t, app)
	waitForRunningCount(t, p, 2, 5*time.Second)

	// Mark one instance busy, leave the other idle.
	var busyID string
	p.Lock()
	for id, inst := range p.Instances {
		if busyID == "" {
			inst.SetUserCount(5, true)
			busyID = id
		}
	}
	p.Unlock()

	app.Workers = 1
	p.Update(app)
	waitForRunningCount(t, p, 1, 5*time.Second)

	p.RLock()
	_, busyStillRunning := p.Instances[busyID]
	p.RUnlock()
	if !busyStillRunning {
		t.Error("expected the busy instance to remain running after scale-down")
	}
}

func TestAppProxyRescaleDownWaitsForBusyInstanceToEmpty(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   2,
		IsActive:  true,
	}
	p := newTestAppProxy(t, app)
	waitForRunningCount(t, p, 2, 5*time.Second)

	// Both instances have connected users.
	var ids []string
	p.Lock()
	for id, inst := range p.Instances {
		inst.SetUserCount(1, true)
		ids = append(ids, id)
	}
	p.Unlock()

	app.Workers = 1
	p.Update(app)

	// One instance should be phasing out but not yet stopped, since it still
	// has a connected user.
	deadline := time.Now().Add(3 * time.Second)
	sawPhasingOut := false
	for time.Now().Before(deadline) {
		status := p.GetStatus(false)
		if status["PhasingOutInst"] == 1 {
			sawPhasingOut = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !sawPhasingOut {
		t.Fatal("expected one instance to enter PHASING_OUT while it still has a connected user")
	}

	p.RLock()
	stillTwo := len(p.Instances) == 2
	p.RUnlock()
	if !stillTwo {
		t.Error("expected the phasing-out instance to remain until its users leave")
	}

	// Now free up all users and rescale again; the phased-out instance
	// should actually stop.
	p.Lock()
	for _, inst := range p.Instances {
		inst.SetUserCount(0, false)
	}
	p.Unlock()
	go p.Rescale()

	waitForRunningCount(t, p, 1, 3*time.Second)
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		p.RLock()
		n := len(p.Instances)
		p.RUnlock()
		if n == 1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("expected the phased-out idle instance to eventually be removed")
	_ = ids
}

func TestAppProxyAuthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		app     models.App
		setup   func(c *gin.Context)
		allowed bool
	}{
		{
			name:    "public app is always allowed",
			app:     models.App{RestrictAccess: config.AccessLevels.PUBLIC},
			setup:   func(c *gin.Context) {},
			allowed: true,
		},
		{
			name:    "all-users app rejects anonymous",
			app:     models.App{RestrictAccess: config.AccessLevels.ALL_USERS},
			setup:   func(c *gin.Context) {},
			allowed: false,
		},
		{
			name: "all-users app allows logged in user",
			app:  models.App{RestrictAccess: config.AccessLevels.ALL_USERS},
			setup: func(c *gin.Context) {
				c.Set("username", "alice")
			},
			allowed: true,
		},
		{
			name: "specific-groups app rejects user without matching group",
			app: models.App{
				RestrictAccess: config.AccessLevels.SPECIFIC_GROUPS,
				AllowedGroups:  []models.Group{{Name: "editors"}},
			},
			setup: func(c *gin.Context) {
				c.Set("groups", map[string]bool{"viewers": true})
			},
			allowed: false,
		},
		{
			name: "specific-groups app allows user with matching group",
			app: models.App{
				RestrictAccess: config.AccessLevels.SPECIFIC_GROUPS,
				AllowedGroups:  []models.Group{{Name: "editors"}},
			},
			setup: func(c *gin.Context) {
				c.Set("groups", map[string]bool{"editors": true})
			},
			allowed: true,
		},
		{
			name:    "unknown access level rejects",
			app:     models.App{RestrictAccess: 99},
			setup:   func(c *gin.Context) {},
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &AppProxy{App: tt.app}
			c, _ := gin.CreateTestContext(nil)
			tt.setup(c)
			if got := p.Authorized(c); got != tt.allowed {
				t.Errorf("expected Authorized()=%v, got %v", tt.allowed, got)
			}
		})
	}
}

func TestAppProxyGetSessionStickiness(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   2,
		IsActive:  true,
	}
	p := newTestAppProxy(t, app)
	waitForRunningCount(t, p, 2, 5*time.Second)

	sess1, err := p.GetSession("", false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if sess1.Instance == nil {
		t.Fatal("expected a session bound to a running instance")
	}

	sess2, err := p.GetSession(sess1.ID, false)
	if err != nil {
		t.Fatalf("GetSession (reuse) failed: %v", err)
	}
	if sess2.ID != sess1.ID {
		t.Errorf("expected same session ID on reuse, got %s vs %s", sess1.ID, sess2.ID)
	}
	if sess2.Instance != sess1.Instance {
		t.Error("expected session to stick to the same instance across calls")
	}
}

func TestAppProxyGetSessionErrorsWithNoRunningInstance(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   0,
		IsActive:  true,
	}
	p := newTestAppProxy(t, app)

	_, err := p.GetSession("", false)
	if err == nil {
		t.Fatal("expected error when no running instance is available")
	}
}

func TestAppProxyCloseSessionRemovesIt(t *testing.T) {
	appDir := newTestAppDir(t)
	app := models.App{
		Name:      "testapp",
		AppSource: "directory",
		AppDir:    appDir,
		Workers:   1,
		IsActive:  true,
	}
	p := newTestAppProxy(t, app)
	waitForRunningCount(t, p, 1, 5*time.Second)

	sess, err := p.GetSession("", false)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if err := p.CloseSession(sess.ID); err != nil {
		t.Fatalf("CloseSession failed: %v", err)
	}

	p.RLock()
	_, exists := p.Sessions[sess.ID]
	p.RUnlock()
	if exists {
		t.Error("expected session to be removed after CloseSession")
	}

	if err := p.CloseSession(sess.ID); err == nil {
		t.Error("expected error closing an already-closed session")
	}
}
