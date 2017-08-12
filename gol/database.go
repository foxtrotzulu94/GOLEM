package gol

import (
	"sort"

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

func LoadListElements(elementType string, filterRemoved, filterViewed bool) OrderedList {
	db := getDatabase()
	defer db.Close()

	var RetVal OrderedList
	var InterfaceList []ListElement
	validEntries := 0

	switch elementType {
	case "anime":
		var MainList []AnimeListElement
		db.Find(&MainList)
		InterfaceList = make([]ListElement, len(MainList))

		for _, item := range MainList {
			var BaseElement ListElementFields
			db.Where("owner_id = ?", item.ID).First(&BaseElement)

			item.Base = BaseElement
			skipItem := (BaseElement.WasRemoved && filterRemoved) || (BaseElement.WasViewed && filterViewed)
			if !skipItem {
				InterfaceList[validEntries] = item
				validEntries++
			}
		}

	case "books":
		panic("Not Implemented Yet")
	case "games":
		panic("Not Implemented Yet")
	}

	RetVal = OrderedList(InterfaceList[0:validEntries])
	sort.Sort(sort.Reverse(RetVal))

	return RetVal
}
