package gol

import (
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
