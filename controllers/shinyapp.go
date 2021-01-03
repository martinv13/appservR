package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
)

type ShinyAppController struct{}

var shinyAppModel = new(models.ShinyApp)

func (u ShinyAppController) Update(c *gin.Context) {
	var app models.ShinyAppUpdate

	if err := c.ShouldBind(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if app.ID != "" {
		app, err := shinyAppModel.Update(app)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve user", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User founded!", "user": app})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
	return
}
