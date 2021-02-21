package controllers

import (
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
			c.HTML(http.StatusNotFound, "groups.html", gin.H{
				"selTab":         "groups",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   "Group not found",
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
		oldGroupName := c.Param("groupname")
		var groupInfo GroupPayload
		err := c.ShouldBind(&groupInfo)
		if err != nil {
			c.HTML(http.StatusBadRequest, "group.html", gin.H{"errorMessage": "Update failed. Please check provided information."})
			c.Abort()
			return
		}

		group := models.Group{Name: groupInfo.GroupName}
		err = group.Update(db, oldGroupName)

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
		group.Delete(db)
	}
}
