package controllers

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/appservR/appservR/models"
	"github.com/gin-gonic/gin"
)

func TestUserControllerGetUsers(t *testing.T) {
	ctl := NewUserController(newFakeUserModel(models.User{Username: "alice"}))
	c, w := newTestContext(http.MethodGet, "/admin/users", nil)
	invoke(c, ctl.GetUsers())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUserControllerGetUserNew(t *testing.T) {
	ctl := NewUserController(newFakeUserModel())
	c, w := newTestContext(http.MethodGet, "/admin/users/new", nil)
	c.Params = gin.Params{{Key: "username", Value: "new"}}
	invoke(c, ctl.GetUser())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUserControllerGetUserNotFound(t *testing.T) {
	ctl := NewUserController(newFakeUserModel())
	c, w := newTestContext(http.MethodGet, "/admin/users/ghost", nil)
	c.Params = gin.Params{{Key: "username", Value: "ghost"}}
	invoke(c, ctl.GetUser())
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown user, got %d", w.Code)
	}
}

func TestUserControllerGetUserExisting(t *testing.T) {
	ctl := NewUserController(newFakeUserModel(models.User{Username: "alice"}))
	c, w := newTestContext(http.MethodGet, "/admin/users/alice", nil)
	c.Params = gin.Params{{Key: "username", Value: "alice"}}
	invoke(c, ctl.GetUser())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUserControllerAdminUpdateUserSuccess(t *testing.T) {
	userModel := newFakeUserModel()
	ctl := NewUserController(userModel)

	form := url.Values{}
	form.Set("username", "bob")
	form.Set("displayedname", "Bob")
	form.Set("groups", "editors")

	c, w := newFormRequest(http.MethodPost, "/admin/users/new", form, gin.Params{{Key: "username", Value: "new"}})
	invoke(c, ctl.AdminUpdateUser())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, ok := userModel.users["bob"]; !ok {
		t.Error("expected user to be created")
	}
}

func TestUserControllerAdminUpdateUserKeepsSelfAdmin(t *testing.T) {
	userModel := newFakeUserModel(models.User{Username: "alice", Groups: []models.Group{{Name: "admins"}}})
	ctl := NewUserController(userModel)

	// alice, logged in as herself, submits a form that removes her own
	// admin group membership.
	form := url.Values{}
	form.Set("username", "alice")
	form.Set("displayedname", "Alice")

	c, w := newFormRequest(http.MethodPost, "/admin/users/alice", form, gin.Params{{Key: "username", Value: "alice"}})
	c.Set("username", "alice")
	invoke(c, ctl.AdminUpdateUser())

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	updated := userModel.users["alice"]
	var stillAdmin bool
	for _, g := range updated.Groups {
		if g.Name == "admins" {
			stillAdmin = true
		}
	}
	if !stillAdmin {
		t.Error("expected the logged-in admin to be unable to remove their own admin group")
	}
}

func TestUserControllerAdminUpdateUserValidationFailure(t *testing.T) {
	userModel := newFakeUserModel()
	userModel.adminErr = errors.New("username already exists")
	ctl := NewUserController(userModel)

	form := url.Values{}
	form.Set("username", "bob")

	c, w := newFormRequest(http.MethodPost, "/admin/users/new", form, gin.Params{{Key: "username", Value: "new"}})
	invoke(c, ctl.AdminUpdateUser())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on model save error, got %d", w.Code)
	}
}

func TestUserControllerDeleteUserSuccess(t *testing.T) {
	userModel := newFakeUserModel(models.User{Username: "bob"})
	ctl := NewUserController(userModel)

	c, w := newTestContext(http.MethodGet, "/admin/users/bob/delete", nil)
	c.Params = gin.Params{{Key: "username", Value: "bob"}}
	c.Set("username", "alice") // a different logged-in user is deleting bob
	invoke(c, ctl.DeleteUser())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, ok := userModel.users["bob"]; ok {
		t.Error("expected bob to be deleted")
	}
}

func TestUserControllerDeleteUserCannotDeleteSelf(t *testing.T) {
	userModel := newFakeUserModel(models.User{Username: "alice"})
	ctl := NewUserController(userModel)

	c, w := newTestContext(http.MethodGet, "/admin/users/alice/delete", nil)
	c.Params = gin.Params{{Key: "username", Value: "alice"}}
	c.Set("username", "alice") // alice trying to delete herself
	invoke(c, ctl.DeleteUser())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when a user tries to delete themselves, got %d", w.Code)
	}
	if _, ok := userModel.users["alice"]; !ok {
		t.Error("expected alice to remain, self-deletion should be blocked")
	}
}

func TestUserControllerDeleteUserRequiresLoggedInUser(t *testing.T) {
	userModel := newFakeUserModel(models.User{Username: "bob"})
	ctl := NewUserController(userModel)

	c, w := newTestContext(http.MethodGet, "/admin/users/bob/delete", nil)
	c.Params = gin.Params{{Key: "username", Value: "bob"}}
	// no "username" set in context, simulating no logged-in user
	invoke(c, ctl.DeleteUser())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 without a logged-in user, got %d", w.Code)
	}
}
