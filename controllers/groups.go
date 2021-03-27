package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
	"gorm.io/gorm"
)

func GetGroups() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		group := models.Group{}
		groups, _ := group.GetAll(db)
		c.HTML(http.StatusOK, "groups.html", gin.H{
			"loggedUserName": GetLoggedName(c),
			"selTab":         "groups",
			"groups":         groups,
		})
	}
}

func GetGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		groupName := c.Param("groupname")

		if groupName == "new" {
			c.HTML(http.StatusOK, "group.html", gin.H{
				"selTab":         "groups",
				"loggedUserName": GetLoggedName(c),
			})
			return
		}

		group := models.Group{Name: groupName}
		err := group.Get(db)
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

func UpdateGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)

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
		err = group.Update(db, oldGroupName)

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

func AddGroupMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		group := models.Group{Name: c.Param("groupname")}
		group.AddMember(db, c.Param("username"))
	}
}

func RemoveGroupMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		group := models.Group{Name: c.Param("groupname")}
		group.RemoveMember(db, c.Param("username"))
	}
}

func DeleteGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		group := models.Group{Name: c.Param("groupname")}
		resData := gin.H{"loggedUserName": GetLoggedName(c), "selTab": "groups"}
		err := group.Delete(db)
		if err != nil {
			resData["errorMessage"] = fmt.Sprintf("Could not delete group '%s'", group.Name)
			c.HTML(http.StatusBadRequest, "group.html", resData)
			c.Abort()
			return
		}
		resData["successMessage"] = fmt.Sprintf("Group '%s' has been deleted.", group.Name)
		group = models.Group{}
		groups, err := group.GetAll(db)
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
