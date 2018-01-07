package gol

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/jinzhu/gorm"
)

type BookListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	Pages       int
	Price       float64
	ReleaseDate time.Time
}

func (item BookListElement) rateElement() float32 {
	//Book ratings are a function of their price, time and total length
	//Thus we favor old, cheap, short highly regarded books
	sourceRatingFactor := float32(0.7)
	timeFactor := 0.1 / 10000000
	lengthFactor := 0.1 * 200.0 //Books shorter than 200 pages are preferrable
	priceFactor := 0.4          //Price is an important factor! (for now)

	timeSinceToday := time.Now().Unix() - item.ReleaseDate.Unix()
	averagePrice := 50.0 //Paying more than $60.0 for a book is discouraged

	sourceContrib := sourceRatingFactor * item.Base.SourceRating
	timeContrib := float32(timeFactor) * float32(timeSinceToday)
	var lengthContrib float32
	if item.Pages != 0 {
		lengthContrib = float32(lengthFactor / float64(item.Pages))
	} else {
		lengthContrib = 1
	}

	priceContrib := float32(priceFactor * (averagePrice - item.Price))
	if sourceContrib < 0.01 {
		sourceContrib = 50*(timeContrib+lengthContrib) + 10*priceContrib
	}
	return sourceContrib + timeContrib + lengthContrib + priceContrib
}

func (item BookListElement) getListName() string {
	return "books"
}

func (item BookListElement) getStoredName() string {
	return gorm.ToDBName(reflect.TypeOf(item).Name()) + "s"
}

func (item BookListElement) getDerivedID() int {
	return item.ID
}

func (item BookListElement) getListElementFields() ListElementFields {
	return item.Base
}

func (item BookListElement) printInfo() {
	//URL is too long sometimes here :/
	fmt.Printf("[ID-%03d] (%.2f) \"%s\" - %s\n", item.ID, item.Base.HeuristicRating, item.Base.Name, item.ReleaseDate.Format("2006-01-02"))
}

func (item BookListElement) printDetailedInfo() {
	fmt.Printf("[ID-%03d] \"%s\"\n(%s)\n", item.ID, item.Base.Name, item.Base.URL)
	fmt.Printf("\tHeuristic Rating: %.2f - Source Rating: %.2f - Price: %.2f\n", item.Base.HeuristicRating, item.Base.SourceRating, item.Price)
	fmt.Printf("\tLength: %d pages | Release Date: %s \n", item.Pages, item.ReleaseDate.Format("2006-01-02"))

	PrintSetWidth(item.Base.Description, "\t ", "\n", 80)
	fmt.Println("")
}

func (item BookListElement) wasFinished() bool {
	return item.Base.WasViewed
}

func (item BookListElement) wasRemoved() bool {
	return item.Base.WasRemoved
}

func (item BookListElement) updateRating(newRating float32) ListElement {
	item.Base.HeuristicRating = newRating
	return item.saveElement()
}

func (item BookListElement) saveElement() ListElement {
	db := getDatabase()

	//NOTE: this is prone to breaking
	if db.NewRecord(item) {
		db.Create(&item)
	} else {
		db.Save(&item)
	}

	return item
}

func (item BookListElement) saveOrderedList(list OrderedList) {
	db := getDatabase()
	for _, element := range list {
		listEntry := element.(BookListElement)
		db.Create(&listEntry)
	}
}

func (item BookListElement) load(derivedID int) ListElement {
	db := getDatabase()
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

func (item BookListElement) loadOrderedList() OrderedList {
	db := getDatabase()

	var MainList []BookListElement
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
