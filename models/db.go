package models

import (
	"errors"

	"github.com/appservR/appservR/modules/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			return nil, errors.New("failed to connect to the database")
		}
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Group{})
	db.AutoMigrate(&ShinyApp{})

	return db, nil
}
