package models

import (
	"github.com/martinv13/go-shiny/modules/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {

	db, err := gorm.Open(sqlite.Open(config.ExecutableFolder+"/data.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})

	db.AutoMigrate(&ShinyApp{})

	return db
}
