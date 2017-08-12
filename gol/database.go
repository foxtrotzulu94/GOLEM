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

func saveAnimeEntries(db *gorm.DB, sortedElements OrderedList) {
	for _, element := range sortedElements {
		listEntry := element.(AnimeListElement)
		db.Create(&listEntry)
	}
}

func saveBookEntries(db *gorm.DB, sortedElements OrderedList) {
	panic("Not Implemented Yet")
}

func saveGameEntries(db *gorm.DB, sortedElements OrderedList) {
	panic("Not Implemented Yet")
}

func saveListElements(elementType string, sortedElements OrderedList) {
	db := getDatabase()
	switch elementType {
	case "anime":
		saveAnimeEntries(db, sortedElements)
	case "books":
		saveBookEntries(db, sortedElements)
	case "games":
		saveGameEntries(db, sortedElements)
	}
	db.Close()
}
