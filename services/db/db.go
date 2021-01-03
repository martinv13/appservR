package db

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

var db *xorm.Engine

func Init() {
	engine, err := xorm.NewEngine("sqlite3", "./test.db")
	if err != nil {
		fmt.Println(err)
	}
	db = engine
}

func GetDB() *xorm.Engine {
	return db
}
