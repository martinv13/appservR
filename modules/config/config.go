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

type Config struct {
	ExecutableFolder string
	v                *viper.Viper
}

func NewConfig() (*Config, error) {

	c := &Config{
		ExecutableFolder: ".",
	}

	c.v = viper.New()

	if !service.Interactive() {
		exePath, _ := osext.ExecutableFolder()
		c.ExecutableFolder = exePath
	}

	c.v.SetDefault("server.port", 8080)
	c.v.SetDefault("server.host", "localhost")
	c.v.SetDefault("server.name", "localhost")

	// find R executable
	RScript := "Rscript"
	if runtime.GOOS == "windows" {
		RPath := "C:/Program Files/R"
		file, err := os.Open(RPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		names, err := file.Readdirnames(0)
		if err != nil {
			return nil, err
		}
		RScript = RPath + "/" + names[len(names)-1] + "/bin/Rscript.exe"
	}

	c.v.SetDefault("RScript", RScript)

	c.v.SetDefault("database.type", "sqlite")
	c.v.SetDefault("database.path", c.ExecutableFolder+"/data.db")

	flag.String("mode", "prod", "run mode")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	c.v.BindPFlags(pflag.CommandLine)

	flag.Usage = func() {
		fmt.Println("Usage: server -mode {mode}")
		os.Exit(1)
	}

	c.v.SetConfigName("config")
	c.v.AddConfigPath("/etc/appname/")
	c.v.AddConfigPath("$HOME/.appname")
	c.v.AddConfigPath(c.ExecutableFolder)

	if err := c.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			c.v.WriteConfigAs(c.ExecutableFolder + "/config.yaml")
		} else {
			return nil, err
		}
	}
	return c, nil
}

func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

func (c *Config) Set(key string, val interface{}) {
	c.v.Set(key, val)
}
