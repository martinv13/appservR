// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/martinv13/go-shiny/controllers"
	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/modules/appproxy"
	"github.com/martinv13/go-shiny/modules/config"
	"github.com/martinv13/go-shiny/modules/ssehandler"
	"github.com/martinv13/go-shiny/modules/vfsdata"
	"github.com/martinv13/go-shiny/server"
)

func InitializeServer() (*server.AppRouter, error) {
	wire.Build(server.NewAppRouter, config.NewConfig, models.NewDB, vfsdata.NewStaticPaths,
		ssehandler.NewMessageBroker, appproxy.NewAppServer,
		models.NewAppModelDB, wire.Bind(new(models.AppModel), new(*models.AppModelDB)),
		models.NewUserModelDB, wire.Bind(new(models.UserModel), new(*models.UserModelDB)),
		models.NewGroupModelDB, wire.Bind(new(models.GroupModel), new(*models.GroupModelDB)),
		controllers.NewAppController, controllers.NewUserController, controllers.NewGroupController,
		controllers.NewAuthController)
	return &server.AppRouter{}, nil
}
