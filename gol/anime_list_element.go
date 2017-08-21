package gol

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/jinzhu/gorm"
)

type AnimeListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	NumEpisodes int
	AirTime     time.Time
}

func (item AnimeListElement) rateElement() float32 {
	if item.Base.IsRated {
		return item.Base.HeuristicRating
	}

	lengthFactor := float32(1.5)
	// dateFactor := float32(1.0)
	return (item.Base.SourceRating * 10.0) - (float32(item.NumEpisodes) * lengthFactor)
}

func (item AnimeListElement) getListName() string {
	return "anime"
}

func (item AnimeListElement) getStoredName() string {
	return gorm.ToDBName(reflect.TypeOf(item).Name()) + "s"
}

func (item AnimeListElement) getDerivedID() int {
	return item.ID
}

func (item AnimeListElement) getListElementFields() ListElementFields {
	return item.Base
}

func (item AnimeListElement) printInfo() {
	fmt.Printf("[ID-%03d] (%.2f) \"%s\" - %d Episode(s) - %s\n", item.ID, item.Base.HeuristicRating, item.Base.Name, item.NumEpisodes, item.Base.URL)
}

func (item AnimeListElement) printDetailedInfo() {
	fmt.Printf("[ID-%03d] \"%s\" (%s)\n", item.ID, item.Base.Name, item.Base.URL)
	fmt.Printf("\tHeuristic Rating: %.2f\n", item.Base.HeuristicRating)
	fmt.Printf("\tEpisodes: %d | Base Rating: %.2f \n", item.NumEpisodes, item.Base.SourceRating)

	PrintSetWidth(item.Base.Description, "\t ", "\n", 80)
	fmt.Println("")
}

func (item AnimeListElement) wasFinished() bool {
	return item.Base.WasViewed
}

func (item AnimeListElement) wasRemoved() bool {
	return item.Base.WasRemoved
}

func (item AnimeListElement) saveElement() ListElement {
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

func (item AnimeListElement) saveOrderedList(list OrderedList) {
	db := getDatabase()
	for _, element := range list {
		listEntry := element.(AnimeListElement)
		db.Create(&listEntry)
	}
	db.Close()
}

func (item AnimeListElement) load(derivedID int) ListElement {
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

func (item AnimeListElement) loadOrderedList() OrderedList {
	db := getDatabase()
	defer db.Close()

	var MainList []AnimeListElement
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
