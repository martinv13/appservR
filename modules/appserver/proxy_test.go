package appserver

import (
	"bufio"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/appservR/appservR/middlewares"
	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/auth"
	"github.com/appservR/appservR/modules/config"
	"github.com/gin-gonic/gin"
)

func activeApp(t testing.TB, name, path string, restrict int, allowedGroups ...models.Group) models.App {
	t.Helper()
	return models.App{
		Name:           name,
		Path:           path,
		AppSource:      "directory",
		AppDir:         newTestAppDir(t),
		Workers:        1,
		IsActive:       true,
		RestrictAccess: restrict,
		AllowedGroups:  allowedGroups,
	}
}

func newProxyTestServer(t testing.TB, apps ...models.App) (*httptest.Server, *AppServer) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	s := newTestAppServer(t, apps...)
	for _, app := range apps {
		waitForAppRunningCount(t, s, app.Name, app.Workers, 5*time.Second)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	tmpl := template.Must(template.New("appnotfound.html").Parse("appnotfound: {{if .err}}{{.err}}{{end}}"))
	router.SetHTMLTemplate(tmpl)
	router.Use(middlewares.Auth())
	router.Use(s.CreateProxy())

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server, s
}

func waitForAppRunningCount(t testing.TB, s *AppServer, appName string, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := s.GetStatus(appName)
		if err == nil && status["RunningInst"] == want {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for app %q to have %d running instances", appName, want)
}

func noRedirectClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func TestProxyRedirectsAppRootToTrailingSlash(t *testing.T) {
	app := activeApp(t, "myapp", "/myapp", config.AccessLevels.PUBLIC)
	server, _ := newProxyTestServer(t, app)

	resp, err := noRedirectClient().Get(server.URL + "/myapp")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/myapp/" {
		t.Errorf("expected redirect to /myapp/, got %q", loc)
	}
}

func TestProxyServesAppRootAndSetsCookies(t *testing.T) {
	app := activeApp(t, "myapp", "/myapp", config.AccessLevels.PUBLIC)
	server, _ := newProxyTestServer(t, app)

	resp, err := http.Get(server.URL + "/myapp/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "mock-shiny-port=") {
		t.Errorf("expected proxied response from mock backend, got %q", body)
	}

	var gotAppID, gotSession bool
	for _, c := range resp.Cookies() {
		if c.Name == "appservr_appid" && c.Value == "myapp" {
			gotAppID = true
		}
		if c.Name == "appservr_session" && c.Value != "" {
			gotSession = true
		}
	}
	if !gotAppID {
		t.Error("expected appservr_appid cookie set to 'myapp'")
	}
	if !gotSession {
		t.Error("expected non-empty appservr_session cookie")
	}
}

func TestProxyStripsPathPrefixAndIsSessionSticky(t *testing.T) {
	app := activeApp(t, "myapp", "/myapp", config.AccessLevels.PUBLIC)
	app.Workers = 2 // multiple instances, so stickiness is meaningfully tested
	server, s := newProxyTestServer(t, app)
	waitForAppRunningCount(t, s, "myapp", 2, 5*time.Second)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}

	resp1, err := client.Get(server.URL + "/myapp/")
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()
	firstPort := extractMockPort(t, string(body1))

	resp2, err := client.Get(server.URL + "/myapp/some/sub/page")
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	defer resp2.Body.Close()
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body2)

	if !strings.Contains(text, "path=/some/sub/page") {
		t.Errorf("expected app path prefix to be stripped, got %q", text)
	}
	secondPort := extractMockPort(t, text)
	if secondPort != firstPort {
		t.Errorf("expected session to stick to the same instance (port %s), got %s", firstPort, secondPort)
	}
}

func extractMockPort(t testing.TB, body string) string {
	t.Helper()
	const marker = "mock-shiny-port="
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("could not find %q in response body %q", marker, body)
	}
	rest := body[i+len(marker):]
	end := strings.IndexAny(rest, " \n")
	if end < 0 {
		end = len(rest)
	}
	return rest[:end]
}

func TestProxyForwardsAuthHeadersToBackend(t *testing.T) {
	app := activeApp(t, "myapp", "/myapp", config.AccessLevels.PUBLIC)
	server, _ := newProxyTestServer(t, app)

	user := models.User{Username: "alice", DisplayedName: "Alice A."}
	token := auth.GenerateToken(user)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/myapp/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	text := string(body)
	if !strings.Contains(text, "username=alice") {
		t.Errorf("expected forwarded username=alice, got %q", text)
	}
	if !strings.Contains(text, "displayedname=Alice A.") {
		t.Errorf("expected forwarded displayedname, got %q", text)
	}
	if !strings.Contains(text, "appname=myapp") {
		t.Errorf("expected forwarded appname=myapp, got %q", text)
	}
}

func TestProxyRejectsUnauthorizedApp(t *testing.T) {
	app := activeApp(t, "myapp", "/myapp", config.AccessLevels.ALL_USERS)
	server, _ := newProxyTestServer(t, app)

	resp, err := http.Get(server.URL + "/myapp/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unauthorized access, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "appnotfound") {
		t.Errorf("expected appnotfound.html body, got %q", body)
	}
}

func TestProxyNoMatchingAppReturns404(t *testing.T) {
	app := activeApp(t, "myapp", "/myapp", config.AccessLevels.PUBLIC)
	server, _ := newProxyTestServer(t, app)

	resp, err := http.Get(server.URL + "/does-not-exist")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unmatched path, got %d", resp.StatusCode)
	}
}

func TestProxyWebsocketDisconnectClosesSessionAndDecrementsUserCount(t *testing.T) {
	app := activeApp(t, "myapp", "/myapp", config.AccessLevels.PUBLIC)
	server, s := newProxyTestServer(t, app)

	// Establish a normal session first, to get sticky cookies.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	resp, err := client.Get(server.URL + "/myapp/")
	if err != nil {
		t.Fatalf("initial request failed: %v", err)
	}
	resp.Body.Close()

	serverURL, _ := url.Parse(server.URL)
	var sessionID, appID string
	for _, c := range jar.Cookies(serverURL) {
		if c.Name == "appservr_session" {
			sessionID = c.Value
		}
		if c.Name == "appservr_appid" {
			appID = c.Value
		}
	}
	if sessionID == "" || appID == "" {
		t.Fatal("expected session/appid cookies to be set after initial request")
	}

	appProxy := s.appsByName["myapp"]
	appProxy.RLock()
	_, sessionPresent := appProxy.Sessions[sessionID]
	appProxy.RUnlock()
	if !sessionPresent {
		t.Fatal("expected session to be tracked by the app proxy before opening the websocket")
	}

	conn, err := net.Dial("tcp", serverURL.Host)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	request := "GET /myapp/websocket/ HTTP/1.1\r\n" +
		"Host: " + serverURL.Host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"Cookie: appservr_appid=" + appID + "; appservr_session=" + sessionID + "\r\n" +
		"\r\n"
	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatalf("failed to write websocket handshake: %v", err)
	}

	httpResp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodGet})
	if err != nil {
		t.Fatalf("failed to read websocket handshake response: %v", err)
	}
	if httpResp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101 Switching Protocols, got %d", httpResp.StatusCode)
	}

	// Simulate the client disconnecting.
	conn.Close()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		appProxy.RLock()
		_, stillPresent := appProxy.Sessions[sessionID]
		appProxy.RUnlock()
		if !stillPresent {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("expected session to be removed after websocket disconnect")
}
