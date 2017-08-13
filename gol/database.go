package gol

import (
	"fmt"
	"reflect"
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
			tableName := gorm.ToDBName(reflect.TypeOf(item).Name()) + "s"
			db.Where("owner_id = ? AND owner_type = ?", item.ID, tableName).First(&BaseElement)

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

func getElementByID(elementType string, elementID int) ListElement {
	db := getDatabase()
	defer db.Close()

	switch elementType {
	case "anime":
		var entry AnimeListElement
		db.First(&entry, elementID)

		var BaseElement ListElementFields
		tableName := gorm.ToDBName(reflect.TypeOf(entry).Name()) + "s" //It would seem the gorm people missed the 's'
		db.Where("owner_id = ? AND owner_type = ?", elementID, tableName).Find(&BaseElement)
		entry.Base = BaseElement
		return entry
	case "books":
		panic("Not Implemented Yet")
	case "games":
		panic("Not Implemented Yet")
	}

	return nil
}

func modifyListElementFields(elementType, fieldName string, newValue interface{}, elementID int) {
	db := getDatabase()
	defer db.Close()

	var element ListElementFields
	var tableName string
	switch elementType {
	case "anime":
		var entry AnimeListElement
		tableName = gorm.ToDBName(reflect.TypeOf(entry).Name()) + "s"
	case "books":
		panic("Not Implemented Yet")
	case "games":
		panic("Not Implemented Yet")
	}

	db.Where("owner_id = ? AND owner_type = ?", elementID, tableName).Find(&element)

	reflectedObject := reflect.ValueOf(&element).Elem()
	objectField := reflectedObject.FieldByName(fieldName)
	if objectField.CanSet() {
		objectField.Set(reflect.ValueOf(newValue))
	} else {
		fmt.Println("Object cannot be modified")
	}

	db.Save(&element)
}

func modifyListElement(entry ListElement, elementType, fieldName string, newValue interface{}) {
	var elementID int
	switch elementType {
	case "anime":
		elementID = entry.(AnimeListElement).ID
	case "books":
		panic("Not Implemented Yet")
	case "games":
		panic("Not Implemented Yet")
	}

	//Send it off to the larger method to avoid code duplication
	modifyListElementFields(elementType, fieldName, newValue, elementID)
}
