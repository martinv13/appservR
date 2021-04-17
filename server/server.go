package server

import (
	"fmt"

	"github.com/martinv13/go-shiny/modules/ssehandler"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func Init(db *gorm.DB, stream *ssehandler.Event) {
	r := NewRouter(db, stream)
	r.Run(fmt.Sprintf("%s:%s", viper.GetString("server.host"), viper.GetString("server.port")))
}
