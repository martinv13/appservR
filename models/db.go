package models

import (
	"errors"

	"github.com/appservR/appservR/modules/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(conf config.Config) (*gorm.DB, error) {

	dbType := conf.GetString("database.type")

	var db *gorm.DB
	var err error

	if dbType == "sqlite" {
		dbPath := conf.GetString("database.path")
		if dbPath == "memory" {
			dbPath = "file::memory:?cache=shared"
		}
		mode := conf.GetString("mode")
		gormConf := &gorm.Config{}
		if mode == "prod" {
			gormConf.Logger = logger.Default.LogMode(logger.Silent)
		}
		db, err = gorm.Open(sqlite.Open(dbPath), gormConf)
		if err != nil {
			return nil, errors.New("failed to connect to the database")
		}
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Group{})
	db.AutoMigrate(&App{})

	return db, nil
}
