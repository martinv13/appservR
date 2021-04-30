package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
)

type GroupController struct {
	groupModel models.GroupModel
}

func NewGroupController(groupModel models.GroupModel) *GroupController {
	return &GroupController{
		groupModel: groupModel,
	}
}

func (ctl *GroupController) GetGroups() gin.HandlerFunc {
	return func(c *gin.Context) {
		groups, _ := ctl.groupModel.All()
		c.HTML(http.StatusOK, "groups.html", gin.H{
			"loggedUserName": GetLoggedName(c),
			"selTab":         "groups",
			"groups":         groups,
		})
	}
}

func (ctl *GroupController) GetGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupName := c.Param("groupname")

		if groupName == "new" {
			c.HTML(http.StatusOK, "group.html", gin.H{
				"selTab":         "groups",
				"loggedUserName": GetLoggedName(c),
			})
			return
		}

		group, err := ctl.groupModel.FindByName(groupName)

		if err != nil {
			c.HTML(http.StatusNotFound, "group.html", gin.H{
				"selTab":         "groups",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   fmt.Sprintf("Group '%s' not found.", groupName),
			})
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "group.html", gin.H{
			"loggedUserName": GetLoggedName(c),
			"selTab":         "groups",
			"GroupName":      group.Name,
		})
	}
}

type GroupPayload struct {
	GroupName string `form:"groupname" binding:"required"`
}

func (ctl *GroupController) UpdateGroup() gin.HandlerFunc {
	return func(c *gin.Context) {

		resMap := gin.H{
			"selTab":         "groups",
			"loggedUserName": GetLoggedName(c),
		}

		oldGroupName := c.Param("groupname")
		var groupInfo GroupPayload
		err := c.ShouldBind(&groupInfo)
		if err != nil {
			resMap["errorMessage"] = "Update failed. Please check provided information."
			c.HTML(http.StatusBadRequest, "group.html", resMap)
			c.Abort()
			return
		}

		group := models.Group{Name: groupInfo.GroupName}
		err = ctl.groupModel.Save(&group, oldGroupName)

		if err != nil {
			resMap["errorMessage"] = "Update failed. Please check provided information."
			c.HTML(http.StatusBadRequest, "group.html", resMap)
			c.Abort()
			return
		}

		if oldGroupName == "new" {
			resMap["successMessage"] = "Group has been created."
		} else {
			resMap["successMessage"] = "Group has been updated."
		}
		resMap["GroupName"] = group.Name
		c.HTML(http.StatusOK, "group.html", resMap)
	}
}

func (ctl *GroupController) AddGroupMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctl.groupModel.AddMember(c.Param("groupname"), c.Param("username"))
	}
}

func (ctl *GroupController) RemoveGroupMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctl.groupModel.RemoveMember(c.Param("groupname"), c.Param("username"))
	}
}

func (ctl *GroupController) DeleteGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		group := models.Group{Name: c.Param("groupname")}
		resData := gin.H{"loggedUserName": GetLoggedName(c), "selTab": "groups"}
		err := ctl.groupModel.Delete(&group)
		if err != nil {
			resData["errorMessage"] = fmt.Sprintf("Could not delete group '%s'", group.Name)
			c.HTML(http.StatusBadRequest, "group.html", resData)
			c.Abort()
			return
		}
		resData["successMessage"] = fmt.Sprintf("Group '%s' has been deleted.", group.Name)
		groups, err := ctl.groupModel.All()
		if err != nil {
			resData["errorMessage"] = "Could not retrieve groups."
			c.HTML(http.StatusBadRequest, "groups.html", resData)
			c.Abort()
			return
		}
		resData["groups"] = groups
		c.HTML(http.StatusOK, "groups.html", resData)
	}
}
