package gol

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/jinzhu/gorm"
)

type GameListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	Platform    string
	ReleaseDate time.Time
}

func (item GameListElement) rateElement() float32 {
	dateFactor := 0.25 / 10000000
	//But can you run Crisis?
	timeSinceCrisis := item.ReleaseDate.Unix() - time.Date(2007, 11, 13, 0, 0, 0, 0, time.UTC).Unix()
	return item.Base.SourceRating + (float32(dateFactor) * float32(timeSinceCrisis))
}

func (item GameListElement) getListName() string {
	return "games"
}

func (item GameListElement) getStoredName() string {
	return gorm.ToDBName(reflect.TypeOf(item).Name()) + "s"
}

func (item GameListElement) getDerivedID() int {
	return item.ID
}

func (item GameListElement) getListElementFields() ListElementFields {
	return item.Base
}

func (item GameListElement) printInfo() {
	fmt.Printf("[ID-%03d] (%.2f) \"%s\" - %s - %s\n", item.ID, item.Base.HeuristicRating, item.Base.Name, item.ReleaseDate.Format("2006-01-02"), item.Base.URL)
}

func (item GameListElement) printDetailedInfo() {
	fmt.Printf("[ID-%03d] \"%s\" (%s)\n", item.ID, item.Base.Name, item.Base.URL)
	fmt.Printf("\tHeuristic Rating: %.2f\n", item.Base.HeuristicRating)
	fmt.Printf("\tPlatform: %s | Release Date: %s \n", item.Platform, item.ReleaseDate.Format("2006-01-02"))

	PrintSetWidth(item.Base.Description, "\t ", "\n", 80)
	fmt.Println("")
}

func (item GameListElement) wasFinished() bool {
	return item.Base.WasViewed
}

func (item GameListElement) wasRemoved() bool {
	return item.Base.WasRemoved
}

func (item GameListElement) saveElement() ListElement {
	db := getDatabase()
	defer db.Close()

	//NOTE: this is prone to breaking
	if db.NewRecord(item) {
		db.Create(&item)
	} else {
		db.Update(&item)
	}

	return item
}

func (item GameListElement) saveOrderedList(list OrderedList) {
	db := getDatabase()
	for _, element := range list {
		listEntry := element.(GameListElement)
		db.Create(&listEntry)
	}
	db.Close()
}

func (item GameListElement) load(derivedID int) ListElement {
	db := getDatabase()
	defer db.Close()
	if derivedID == 0 {
		return nil
	}
	db.First(&item, derivedID)

	var BaseElement ListElementFields
	tableName := item.getStoredName()
	db.Where("owner_id = ? AND owner_type = ?", derivedID, tableName).Find(&BaseElement)
	item.Base = BaseElement

	return item
}

func (item GameListElement) loadOrderedList() OrderedList {
	db := getDatabase()
	defer db.Close()

	var MainList []GameListElement
	db.Find(&MainList)
	InterfaceList := make([]ListElement, len(MainList))

	for i, item := range MainList {
		var BaseElement ListElementFields
		tableName := item.getStoredName()
		db.Where("owner_id = ? AND owner_type = ?", item.getDerivedID(), tableName).First(&BaseElement)

		item.Base = BaseElement
		InterfaceList[i] = item
	}

	RetVal := OrderedList(InterfaceList)
	sort.Sort(sort.Reverse(RetVal))

	return RetVal
}
