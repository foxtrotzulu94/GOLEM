package gol

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/jinzhu/gorm"
	//Needed for the gorm package on top
	_ "github.com/mattn/go-sqlite3"
)

var databaseName = "GOL.sqlite3"
var databaseHandle *gorm.DB

//TODO: Add an in memory cache for queries

func getDatabase() *gorm.DB {
	if databaseHandle != nil {
		return databaseHandle
	}

	databasePath, _ := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), databaseName))
	db, err := gorm.Open("sqlite3", databasePath)
	check(err)

	//Handle initialization/creation/migration automatically if possible
	db.AutoMigrate(&ListElementFields{})
	for _, val := range RegisteredTypes {
		db.AutoMigrate(val)
	}

	// Make the database file a passive mutex
	// We can handle concurrent reads (which almost never happen)
	// But the database breaks on concurrent writes (which almost always happen)
	db.DB().SetMaxOpenConns(1)

	databaseHandle = db
	return databaseHandle
}

func closeDatabase() {
	if databaseHandle != nil {
		databaseHandle.Close()
	}
	databaseHandle = nil
}

//Assumes a pointer is being passed
func dbCreateListElement(entry ListElement) {
	db := getDatabase()
	fmt.Println(db.HasTable(entry))
	fmt.Println(reflect.TypeOf(entry))

	if db.NewRecord(entry) {
		db.Create(&entry)
	} else {
		db.Save(&entry)
	}
}

func loadListElements(elementType string, filterActive, filterRemoved, filterViewed bool) OrderedList {
	if filterActive && filterViewed && filterRemoved {
		panic("Programmer error: Result of loadListElements will be empty!")
	}

	MainList := RegisteredTypes[elementType].loadOrderedList()
	InterfaceList := make([]ListElement, len(MainList))

	validEntries := 0
	for _, item := range MainList {
		BaseElement := item.getListElementFields()
		isActive := !(BaseElement.WasRemoved || BaseElement.WasViewed)
		skipItem := (BaseElement.WasRemoved && filterRemoved) || (BaseElement.WasViewed && filterViewed) || (isActive && filterActive)

		if !skipItem {
			InterfaceList[validEntries] = item
			validEntries++
		}
	}

	return OrderedList(InterfaceList[0:validEntries])
}

func getElementByID(elementType string, elementID int) ListElement {
	getDatabase()

	entry := RegisteredTypes[elementType]
	return entry.load(elementID)
}

func modifyListElementFields(entry ListElement, elementType, fieldName string, newValue interface{}) {
	db := getDatabase()

	var element ListElementFields
	tableName := entry.getStoredName()

	db.Where("owner_id = ? AND owner_type = ?", entry.getDerivedID(), tableName).Find(&element)

	reflectedObject := reflect.ValueOf(&element).Elem()
	objectField := reflectedObject.FieldByName(fieldName)
	if objectField.CanSet() {
		objectField.Set(reflect.ValueOf(newValue))
	} else {
		fmt.Println("Object cannot be modified")
	}

	db.Save(&element)
}
