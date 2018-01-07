package gol

import (
	"math"
	"reflect"

	"github.com/jinzhu/gorm"
)

type ListElement interface {
	rateElement() float32

	getListName() string
	getStoredName() string
	getDerivedID() int
	getRating() float32
	getListElementFields() ListElementFields

	wasFinished() bool
	wasRemoved() bool

	printInfo()
	printDetailedInfo()

	// Updates the rating and returns a new list element
	updateRating() ListElement

	//Returns Primary Key
	saveElement() ListElement
	//Intended for bulk operations ONLY
	saveOrderedList(list OrderedList)
	//get a single item
	load(derivedID int) ListElement
	//Load all known elements of this type in sorted order
	loadOrderedList() OrderedList
}

type OrderedList []ListElement

func (slice OrderedList) Len() int {
	return len(slice)
}

func (slice OrderedList) Less(i, j int) bool {
	return slice[i].getRating() < slice[j].getRating()
}

func (slice OrderedList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice OrderedList) save() {
	slice[0].saveOrderedList(slice)
}

type ListElementFields struct {
	gorm.Model

	URL         string `sql:"unique;not null"`
	Name        string `sql:"unique;not null"`
	Description string
	IsRated     bool
	WasViewed   bool
	WasRemoved  bool

	SourceRating    float32
	HeuristicRating float32

	OwnerId   int
	OwnerType string
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

func CreateListElement(elementType, url, name, description string, sourceRating float32) ListElement {
	baseElement := CreateListElementFields(url, name, description, sourceRating)
	retVal := RegisteredTypes[elementType]

	//NOTE: Black magic through reflection due to the inability to modify a struct generically
	reflect.ValueOf(retVal).Elem().FieldByName("Base").Set(reflect.ValueOf(baseElement))

	return retVal
}

//RegisteredTypes Map of all usable types. Returns a pointer to the type
var RegisteredTypes = map[string]ListElement{
	"anime":  &AnimeListElement{},
	"movies": &MovieListElement{},
	"games":  &GameListElement{},
	"books":  &BookListElement{},
}
