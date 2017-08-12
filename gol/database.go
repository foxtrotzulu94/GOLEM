package gol

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var DatabaseName = "GOL.sqlite3"

func getDatabase() *gorm.DB {
	db, err := gorm.Open("sqlite3", DatabaseName)
	check(err)

	//Handle initialization/creation/migration automatically if possible
	db.AutoMigrate(&ListElementFields{}, &AnimeListElement{})

	return db
}
