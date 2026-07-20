package controllers

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/appservR/appservR/models"
)

func TestAuthControllerDoLoginSuccessSetsCookieAndRedirects(t *testing.T) {
	userModel := newFakeUserModel()
	userModel.loginUser = models.User{Username: "alice", DisplayedName: "Alice"}
	ctl := NewAuthController(userModel)

	form := url.Values{}
	form.Set("username", "alice")
	form.Set("password", "secret")
	form.Set("refurl", "/somewhere")

	c, w := newFormRequest(http.MethodPost, "/auth/login", form, nil)
	invoke(c, ctl.DoLogin())

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect on successful login, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/somewhere" {
		t.Errorf("expected redirect to /somewhere, got %q", loc)
	}
	var tokenCookieSet bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "token" && c.Value != "" {
			tokenCookieSet = true
		}
	}
	if !tokenCookieSet {
		t.Error("expected token cookie to be set on successful login")
	}
}

func TestAuthControllerDoLoginRedirectsSignupRefererToRoot(t *testing.T) {
	userModel := newFakeUserModel()
	userModel.loginUser = models.User{Username: "alice"}
	ctl := NewAuthController(userModel)

	form := url.Values{}
	form.Set("username", "alice")
	form.Set("password", "secret")
	form.Set("refurl", "/auth/signup")

	c, w := newFormRequest(http.MethodPost, "/auth/login", form, nil)
	invoke(c, ctl.DoLogin())

	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("expected referer of /auth/signup to redirect to /, got %q", loc)
	}
}

func TestAuthControllerDoLoginFailureRendersError(t *testing.T) {
	userModel := newFakeUserModel()
	userModel.loginErr = errors.New("wrong password")
	ctl := NewAuthController(userModel)

	form := url.Values{}
	form.Set("username", "alice")
	form.Set("password", "wrong")

	c, w := newFormRequest(http.MethodPost, "/auth/login", form, nil)
	invoke(c, ctl.DoLogin())

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on failed login, got %d", w.Code)
	}
}

func TestAuthControllerDoLogoutClearsCookieAndRedirects(t *testing.T) {
	ctl := NewAuthController(newFakeUserModel())

	c, w := newTestContext(http.MethodGet, "/auth/logout", nil)
	invoke(c, ctl.DoLogout())

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
	var cleared bool
	for _, ck := range w.Result().Cookies() {
		if ck.Name == "token" && ck.Value == "" {
			cleared = true
		}
	}
	if !cleared {
		t.Error("expected token cookie to be cleared on logout")
	}
}

func TestAuthControllerDoSignupSuccess(t *testing.T) {
	userModel := newFakeUserModel()
	ctl := NewAuthController(userModel)

	form := url.Values{}
	form.Set("username", "bob")
	form.Set("displayedname", "Bob")
	form.Set("password", "secret")
	form.Set("password2", "secret")

	c, w := newFormRequest(http.MethodPost, "/auth/signup", form, nil)
	invoke(c, ctl.DoSignup())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on successful signup, got %d", w.Code)
	}
	if _, ok := userModel.users["bob"]; !ok {
		t.Error("expected new user to be saved")
	}
}

func TestAuthControllerDoSignupPasswordMismatch(t *testing.T) {
	ctl := NewAuthController(newFakeUserModel())

	form := url.Values{}
	form.Set("username", "bob")
	form.Set("password", "secret")
	form.Set("password2", "different")

	c, w := newFormRequest(http.MethodPost, "/auth/signup", form, nil)
	invoke(c, ctl.DoSignup())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for password mismatch, got %d", w.Code)
	}
}

func TestAuthControllerDoSignupDuplicateUsernameFails(t *testing.T) {
	userModel := newFakeUserModel()
	userModel.saveErr = errors.New("username already exists")
	ctl := NewAuthController(userModel)

	form := url.Values{}
	form.Set("username", "bob")
	form.Set("password", "secret")
	form.Set("password2", "secret")

	c, w := newFormRequest(http.MethodPost, "/auth/signup", form, nil)
	invoke(c, ctl.DoSignup())

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for duplicate username, got %d", w.Code)
	}
}
