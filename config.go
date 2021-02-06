package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func loadConfig() {

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
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.WriteConfigAs("./config.yaml")
		} else {
			fmt.Println(err)
			panic("config file cannot be read")
		}
	}

}
