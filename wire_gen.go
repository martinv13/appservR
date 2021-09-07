// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package main

import (
	"github.com/appservR/appservR/controllers"
	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/appserver"
	"github.com/appservR/appservR/modules/config"
	"github.com/appservR/appservR/modules/ssehandler"
	"github.com/appservR/appservR/modules/vfsdata"
	"github.com/appservR/appservR/server"
)

// Injectors from wire.go:

func InitializeServer(flags config.RunFlags) (*server.AppRouter, error) {
	configViper, err := config.NewConfigViper(flags)
	if err != nil {
		return nil, err
	}
	staticPaths := vfsdata.NewStaticPaths(configViper)
	db, err := models.NewDB(configViper)
	if err != nil {
		return nil, err
	}
	groupModelDB := models.NewGroupModelDB(db)
	appModelDB, err := models.NewAppModelDB(db, groupModelDB)
	if err != nil {
		return nil, err
	}
	messageBroker := ssehandler.NewMessageBroker()
	appServer, err := appserver.NewAppServer(appModelDB, messageBroker, configViper)
	if err != nil {
		return nil, err
	}
	appController := controllers.NewAppController(appModelDB, appServer, configViper)
	userModelDB := models.NewUserModelDB(db, groupModelDB)
	userController := controllers.NewUserController(userModelDB)
	groupController := controllers.NewGroupController(groupModelDB)
	authController := controllers.NewAuthController(userModelDB)
	appRouter, err := server.NewAppRouter(configViper, staticPaths, appServer, messageBroker, appController, userController, groupController, authController)
	if err != nil {
		return nil, err
	}
	return appRouter, nil
}
