package config

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/kardianos/osext"
	"github.com/kardianos/service"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var ExecutableFolder = "."

// Set default config and load config from config.yml
func LoadConfig() {

	if !service.Interactive() {
		exePath, _ := osext.ExecutableFolder()
		ExecutableFolder = exePath
	}

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.name", "localhost")

	// find R executable
	RScript := "Rscript"
	if runtime.GOOS == "windows" {
		RPath := "C:/Program Files/R"
		file, err := os.Open(RPath)
		if err == nil {
			defer file.Close()
			names, err := file.Readdirnames(0)
			if err == nil {
				RScript = RPath + "/" + names[len(names)-1] + "/bin/Rscript.exe"
			}
		}
	}

	viper.SetDefault("RScript", RScript)

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
