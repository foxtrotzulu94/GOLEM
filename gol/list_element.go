package gol

import (
	"fmt"
	"math"
	"time"
)

type ListElement interface {
	rateElement() float32
	getListName() string
	printInfo()
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
	ID int

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
	fmt.Printf("(%04d) \"%s\" [%.2f] - %d Episode(s) - %s\n", item.ID, item.Base.Name, item.Base.HeuristicRating, item.NumEpisodes, item.Base.URL)
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
