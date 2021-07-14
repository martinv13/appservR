package config

import (
	"os"
	"runtime"

	"github.com/kardianos/osext"
	"github.com/kardianos/service"
	"github.com/spf13/viper"
)

type RunFlags struct {
	Address string
	Mode    string
	Port    string
}

type Config interface {
	ExecutableFolder() string
	GetString(string) string
	Logger() *Logger
}

type ConfigViper struct {
	executableFolder string
	v                *viper.Viper
	logger           Logger
}

func (c *ConfigViper) ExecutableFolder() string {
	return c.executableFolder
}

func (c *ConfigViper) GetString(key string) string {
	return c.v.GetString(key)
}

func (c *ConfigViper) Logger() *Logger {
	return &c.logger
}

func NewConfigViper(flags RunFlags) (*ConfigViper, error) {

	c := &ConfigViper{
		executableFolder: ".",
	}

	c.v = viper.New()

	if !service.Interactive() {
		exePath, _ := osext.ExecutableFolder()
		c.executableFolder = exePath
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
	c.v.SetDefault("database.path", c.executableFolder+"/data.db")

	c.v.SetConfigName("config")
	c.v.AddConfigPath("/etc/appname/")
	c.v.AddConfigPath("$HOME/.appname")
	c.v.AddConfigPath(c.executableFolder)

	if err := c.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			c.v.WriteConfigAs(c.executableFolder + "/config.yml")
		} else {
			return nil, err
		}
	}

	c.v.SetDefault("mode", "prod")

	if flags.Port != "" {
		c.v.Set("server.port", flags.Port)
	}
	if flags.Address != "" {
		c.v.Set("server.host", flags.Address)
	}
	if flags.Mode != "" {
		c.v.Set("mode", flags.Mode)
	}

	mode := c.v.GetString("mode")
	if mode == "prod" {
		c.logger = NewLogger(logLevels.WARNING)
	} else {
		c.logger = NewLogger(logLevels.DEBUG)
	}

	return c, nil
}
