package gol

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/jinzhu/gorm"
)

type MovieListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	ReviewCount int
	Duration    int
}

func (item MovieListElement) rateElement() float32 {
	if item.Base.IsRated {
		return item.Base.HeuristicRating
	}

	viewersFactor := float32(0.25)
	return (item.Base.SourceRating * 10.0) + (float32(item.ReviewCount/100000) * viewersFactor)
}

func (item MovieListElement) getListName() string {
	return "movies"
}

func (item MovieListElement) getStoredName() string {
	return gorm.ToDBName(reflect.TypeOf(item).Name()) + "s"
}

func (item MovieListElement) getDerivedID() int {
	return item.ID
}

func (item MovieListElement) getListElementFields() ListElementFields {
	return item.Base
}

func (item MovieListElement) printInfo() {
	fmt.Printf("[ID-%03d] (%.2f) \"%s\" - %s\n", item.ID, item.Base.HeuristicRating, item.Base.Name, item.Base.URL)
}

func (item MovieListElement) printDetailedInfo() {
	fmt.Printf("[ID-%03d] \"%s\" (%s)\n", item.ID, item.Base.Name, item.Base.URL)
	fmt.Printf("\tHeuristic Rating: %.2f\n", item.Base.HeuristicRating)
	fmt.Printf("\tDuration: ? | Base Rating: %.2f \n", item.Base.SourceRating)

	PrintSetWidth(item.Base.Description, "\t ", "\n", 80)
	fmt.Println("")
}

func (item MovieListElement) wasFinished() bool {
	return item.Base.WasViewed
}

func (item MovieListElement) wasRemoved() bool {
	return item.Base.WasRemoved
}

func (item MovieListElement) saveElement() ListElement {
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

func (item MovieListElement) saveOrderedList(list OrderedList) {
	db := getDatabase()
	for _, element := range list {
		listEntry := element.(MovieListElement)
		db.Create(&listEntry)
	}
	db.Close()
}

func (item MovieListElement) load(derivedID int) ListElement {
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

func (item MovieListElement) loadOrderedList() OrderedList {
	db := getDatabase()
	defer db.Close()

	var MainList []MovieListElement
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
