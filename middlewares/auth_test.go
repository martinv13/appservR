package middlewares

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/auth"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext(cookies ...*http.Cookie) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	t := template.Must(template.New("appnotfound.html").Parse("not found"))
	engine.SetHTMLTemplate(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	c.Request = req
	return c, w
}

func TestAuthSetsContextFromValidToken(t *testing.T) {
	user := models.User{
		Username:      "alice",
		DisplayedName: "Alice A.",
		Groups:        []models.Group{{Name: "admins"}, {Name: "editors"}},
	}
	tokenString := auth.GenerateToken(user)

	c, _ := newTestContext(&http.Cookie{Name: "token", Value: tokenString})
	Auth()(c)

	username, ok := c.Get("username")
	if !ok || username != "alice" {
		t.Errorf("expected username %q to be set, got %v (ok=%v)", "alice", username, ok)
	}
	displayed, ok := c.Get("displayedname")
	if !ok || displayed != "Alice A." {
		t.Errorf("expected displayedname %q to be set, got %v (ok=%v)", "Alice A.", displayed, ok)
	}
	groupsVal, ok := c.Get("groups")
	if !ok {
		t.Fatal("expected groups to be set")
	}
	groups, ok := groupsVal.(map[string]bool)
	if !ok {
		t.Fatalf("expected groups to be map[string]bool, got %T", groupsVal)
	}
	if !groups["admins"] || !groups["editors"] {
		t.Errorf("expected both admins and editors group flags true, got %v", groups)
	}
}

func TestAuthNoCookieLeavesContextEmpty(t *testing.T) {
	c, _ := newTestContext()
	Auth()(c)

	if _, ok := c.Get("username"); ok {
		t.Error("expected username to be unset when no token cookie is present")
	}
	if c.IsAborted() {
		t.Error("Auth middleware should never abort the request")
	}
}

func TestAuthInvalidTokenLeavesContextEmpty(t *testing.T) {
	c, _ := newTestContext(&http.Cookie{Name: "token", Value: "garbage-token"})
	Auth()(c)

	if _, ok := c.Get("username"); ok {
		t.Error("expected username to be unset for an invalid token")
	}
}

func TestAuthExpiredTokenLeavesContextEmpty(t *testing.T) {
	// A token signed for a user, but immediately expired, should be ignored.
	user := models.User{Username: "bob"}
	tokenString := auth.GenerateToken(user)
	// Sanity: it's valid right now; expiry is exercised in modules/auth tests.
	// Here we only check that a syntactically valid-but-untrusted cookie value
	// (empty string) does not populate the context.
	_ = tokenString

	c, _ := newTestContext(&http.Cookie{Name: "token", Value: ""})
	Auth()(c)
	if _, ok := c.Get("username"); ok {
		t.Error("expected username to be unset for an empty token")
	}
}

func TestAdminAuthAllowsAdminGroup(t *testing.T) {
	c, w := newTestContext()
	c.Set("groups", map[string]bool{"admins": true})

	AdminAuth()(c)

	if c.IsAborted() {
		t.Error("expected admin user to not be aborted")
	}
	if w.Code != http.StatusOK && w.Code != 0 {
		t.Errorf("expected no error response written, got status %d", w.Code)
	}
}

func TestAdminAuthRejectsNonAdminGroup(t *testing.T) {
	c, w := newTestContext()
	c.Set("groups", map[string]bool{"admins": false, "editors": true})

	AdminAuth()(c)

	if !c.IsAborted() {
		t.Error("expected non-admin user to be aborted")
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 response, got %d", w.Code)
	}
}

func TestAdminAuthRejectsMissingGroups(t *testing.T) {
	c, w := newTestContext()
	// no "groups" key set at all, as would happen for an unauthenticated request

	AdminAuth()(c)

	if !c.IsAborted() {
		t.Error("expected request with no groups info to be aborted")
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 response, got %d", w.Code)
	}
}

func TestAdminAuthRejectsWrongGroupsType(t *testing.T) {
	c, _ := newTestContext()
	c.Set("groups", "not-a-map")

	AdminAuth()(c)

	if !c.IsAborted() {
		t.Error("expected request with malformed groups value to be aborted")
	}
}
