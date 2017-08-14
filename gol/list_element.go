package gol

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
)

type ListElement interface {
	rateElement() float32

	getListName() string
	wasFinished() bool
	wasRemoved() bool

	printInfo()

	//Returns Primary Key
	saveElement() int

	//Intended for bulk operations ONLY
	saveOrderedList(list OrderedList)

	//TODO: Add "printDetailedInfo" and "getListElementFields"
}

type OrderedList []ListElement

func (slice OrderedList) Len() int {
	return len(slice)
}

func (slice OrderedList) Less(i, j int) bool {
	return slice[i].rateElement() < slice[j].rateElement()
}

func (slice OrderedList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type ListElementFields struct {
	gorm.Model

	URL         string `sql:"unique"`
	Name        string `sql:"unique"`
	Description string
	IsRated     bool
	WasViewed   bool
	WasRemoved  bool

	SourceRating    float32
	HeuristicRating float32

	OwnerId   int
	OwnerType string
}

//TODO: Refactor to split the struct that implement the interface

type AnimeListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	NumEpisodes int
	AirTime     time.Time
}

type GameListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	Platform string
	Release  time.Time
	GameType string
}

type BookListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	Category string
}

//FIXME: Can't seem to pass by reference in here without breaking everything
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

func (item AnimeListElement) printInfo() {
	fmt.Printf("[%04d] (%.2f) \"%s\" - %d Episode(s) - %s\n", item.ID, item.Base.HeuristicRating, item.Base.Name, item.NumEpisodes, item.Base.URL)
}

func (item AnimeListElement) wasFinished() bool {
	return item.Base.WasViewed
}

func (item AnimeListElement) wasRemoved() bool {
	return item.Base.WasRemoved
}

func (item AnimeListElement) saveElement() int {
	db := getDatabase()
	defer db.Close()

	//NOTE: this is prone to breaking
	if db.NewRecord(item) {
		db.Create(&item)
	} else {
		db.Update(&item)
	}

	return item.ID
}

func (item AnimeListElement) saveOrderedList(list OrderedList) {
	db := getDatabase()
	for _, element := range list {
		listEntry := element.(AnimeListElement)
		db.Create(&listEntry)
	}
	db.Close()
}

// TODO: Implement the interface methods for each struct

func CreateListElementFields(url, name, description string, sourceRating float32) ListElementFields {
	var common ListElementFields
	common.URL = url
	common.Name = name
	common.SourceRating = sourceRating
	common.HeuristicRating = float32(math.NaN())
	common.Description = description
	common.IsRated = false
	common.WasRemoved = false
	common.WasViewed = false

	return common
}

//RegisteredTypes Map of all usable types. Returns a pointer to the type
var RegisteredTypes = map[string]ListElement{
	"anime": &AnimeListElement{},
}

func CreateListElement(elementType, url, name, description string, sourceRating float32) ListElement {
	baseElement := CreateListElementFields(url, name, description, sourceRating)
	retVal := RegisteredTypes[elementType]

	//NOTE: Black magic through reflection due to the inability to modify a struct generically
	reflect.ValueOf(retVal).Elem().FieldByName("Base").Set(reflect.ValueOf(baseElement))

	return retVal
}
