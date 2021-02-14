package server

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func Init(db *gorm.DB) {
	r := NewRouter(db)
	r.Run(fmt.Sprintf("%s:%s", viper.GetString("server.host"), viper.GetString("server.port")))
}
