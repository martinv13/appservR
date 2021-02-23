package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Debug(s string) {
	if gin.Mode() == gin.DebugMode {
		fmt.Println(s)
	}
}
