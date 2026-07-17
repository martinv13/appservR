package controllers

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/appservR/appservR/models"
	"github.com/gin-gonic/gin"
)

func TestGroupControllerGetGroups(t *testing.T) {
	ctl := NewGroupController(newFakeGroupModel(models.Group{Name: "editors"}))
	c, w := newTestContext(http.MethodGet, "/admin/groups", nil)
	invoke(c, ctl.GetGroups())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGroupControllerGetGroupNew(t *testing.T) {
	ctl := NewGroupController(newFakeGroupModel())
	c, w := newTestContext(http.MethodGet, "/admin/groups/new", nil)
	c.Params = gin.Params{{Key: "groupname", Value: "new"}}
	invoke(c, ctl.GetGroup())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGroupControllerGetGroupNotFound(t *testing.T) {
	ctl := NewGroupController(newFakeGroupModel())
	c, w := newTestContext(http.MethodGet, "/admin/groups/ghost", nil)
	c.Params = gin.Params{{Key: "groupname", Value: "ghost"}}
	invoke(c, ctl.GetGroup())
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown group, got %d", w.Code)
	}
}

func TestGroupControllerGetGroupExisting(t *testing.T) {
	ctl := NewGroupController(newFakeGroupModel(models.Group{Name: "editors"}))
	c, w := newTestContext(http.MethodGet, "/admin/groups/editors", nil)
	c.Params = gin.Params{{Key: "groupname", Value: "editors"}}
	invoke(c, ctl.GetGroup())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGroupControllerUpdateGroupCreatesNew(t *testing.T) {
	groupModel := newFakeGroupModel()
	ctl := NewGroupController(groupModel)

	form := url.Values{}
	form.Set("groupname", "editors")

	c, w := newFormRequest(http.MethodPost, "/admin/groups/new", form, gin.Params{{Key: "groupname", Value: "new"}})
	invoke(c, ctl.UpdateGroup())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, ok := groupModel.groups["editors"]; !ok {
		t.Error("expected new group to be created")
	}
}

func TestGroupControllerUpdateGroupMissingNameFails(t *testing.T) {
	ctl := NewGroupController(newFakeGroupModel())

	c, w := newFormRequest(http.MethodPost, "/admin/groups/new", url.Values{}, gin.Params{{Key: "groupname", Value: "new"}})
	invoke(c, ctl.UpdateGroup())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing required groupname, got %d", w.Code)
	}
}

func TestGroupControllerUpdateGroupModelErrorFails(t *testing.T) {
	groupModel := newFakeGroupModel()
	groupModel.saveErr = errors.New("Admins group cannot be modified.")
	ctl := NewGroupController(groupModel)

	form := url.Values{}
	form.Set("groupname", "admins")

	c, w := newFormRequest(http.MethodPost, "/admin/groups/admins", form, gin.Params{{Key: "groupname", Value: "admins"}})
	invoke(c, ctl.UpdateGroup())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when the model rejects the update, got %d", w.Code)
	}
}

func TestGroupControllerDeleteGroupSuccess(t *testing.T) {
	groupModel := newFakeGroupModel(models.Group{Name: "editors"})
	ctl := NewGroupController(groupModel)

	c, w := newTestContext(http.MethodGet, "/admin/groups/editors/delete", nil)
	c.Params = gin.Params{{Key: "groupname", Value: "editors"}}
	invoke(c, ctl.DeleteGroup())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, ok := groupModel.groups["editors"]; ok {
		t.Error("expected group to be deleted")
	}
}

func TestGroupControllerDeleteGroupModelErrorFails(t *testing.T) {
	groupModel := newFakeGroupModel()
	groupModel.delErr = errors.New("Group 'admins' cannot be deleted")
	ctl := NewGroupController(groupModel)

	c, w := newTestContext(http.MethodGet, "/admin/groups/admins/delete", nil)
	c.Params = gin.Params{{Key: "groupname", Value: "admins"}}
	invoke(c, ctl.DeleteGroup())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when delete is rejected, got %d", w.Code)
	}
}

// AddGroupMember/RemoveGroupMember currently delegate straight to the model
// and do nothing with the result - not even rendering a response. These
// tests pin down that (likely incomplete) behavior rather than assume it: at
// the GroupModel layer, models.GroupModelDB.AddMember/RemoveMember are
// themselves no-op stubs that always return nil (see models/group.go), so
// group membership changes triggered through the admin UI currently have no
// effect at all in the real application.
func TestGroupControllerAddGroupMemberDelegatesToModel(t *testing.T) {
	groupModel := newFakeGroupModel()
	ctl := NewGroupController(groupModel)

	c, _ := newTestContext(http.MethodGet, "/admin/groups/editors/add/bob", nil)
	c.Params = gin.Params{{Key: "groupname", Value: "editors"}, {Key: "username", Value: "bob"}}
	invoke(c, ctl.AddGroupMember())

	if len(groupModel.addCalls) != 1 || groupModel.addCalls[0] != "editors/bob" {
		t.Errorf("expected AddMember to be called with editors/bob, got %v", groupModel.addCalls)
	}
}

func TestGroupControllerRemoveGroupMemberDelegatesToModel(t *testing.T) {
	groupModel := newFakeGroupModel()
	ctl := NewGroupController(groupModel)

	c, _ := newTestContext(http.MethodGet, "/admin/groups/editors/remove/bob", nil)
	c.Params = gin.Params{{Key: "groupname", Value: "editors"}, {Key: "username", Value: "bob"}}
	invoke(c, ctl.RemoveGroupMember())

	if len(groupModel.remCalls) != 1 || groupModel.remCalls[0] != "editors/bob" {
		t.Errorf("expected RemoveMember to be called with editors/bob, got %v", groupModel.remCalls)
	}
}
