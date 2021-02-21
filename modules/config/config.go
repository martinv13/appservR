package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/kardianos/osext"
	"github.com/kardianos/service"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var ExecutableFolder = "."

func LoadConfig() {

	if !service.Interactive() {
		exePath, _ := osext.ExecutableFolder()
		ExecutableFolder = exePath
	}

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.name", "localhost")

	flag.String("mode", "prod", "run mode")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	flag.Usage = func() {
		fmt.Println("Usage: server -mode {mode}")
		os.Exit(1)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/appname/")
	viper.AddConfigPath("$HOME/.appname")
	viper.AddConfigPath(ExecutableFolder)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.WriteConfigAs(ExecutableFolder + "/config.yaml")
		} else {
			fmt.Println(err)
			panic("config file cannot be read")
		}
	}

}
