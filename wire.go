//go:build wireinject
// +build wireinject

package main

import (
	"github.com/appservR/appservR/controllers"
	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/appserver"
	"github.com/appservR/appservR/modules/config"
	"github.com/appservR/appservR/modules/vfsdata"
	"github.com/appservR/appservR/server"
	"github.com/google/wire"
)

func InitializeServer(flags config.RunFlags) (*server.AppRouter, error) {
	wire.Build(server.NewAppRouter, models.NewDB, vfsdata.NewStaticPaths, appserver.NewAppServer,
		config.NewConfigViper, wire.Bind(new(config.Config), new(*config.ConfigViper)),
		models.NewAppModelDB, wire.Bind(new(models.AppModel), new(*models.AppModelDB)),
		models.NewUserModelDB, wire.Bind(new(models.UserModel), new(*models.UserModelDB)),
		models.NewGroupModelDB, wire.Bind(new(models.GroupModel), new(*models.GroupModelDB)),
		controllers.NewAppController, controllers.NewStatusController, controllers.NewUserController,
		controllers.NewGroupController, controllers.NewAuthController)
	return &server.AppRouter{}, nil
}
