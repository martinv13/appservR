package server

import (
	"fmt"

	"github.com/spf13/viper"
)

func Init() {
	r := NewRouter()
	r.Run(fmt.Sprintf("%s:%s", viper.GetString("server.host"), viper.GetString("server.port")))
}
