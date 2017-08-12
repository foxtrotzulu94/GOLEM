package gol

import (
	"math"
	"time"
)

type ListElement interface {
	rateElement() float32
	getListName() string
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

	URL         string `xorm:"unique"`
	Name        string
	Description string
	IsRated     bool

	SourceRating    float32
	HeuristicRating float32

	OwnerId   int
	OwnerType string
}

type AnimeListElement struct {
	ID   int               `xorm:"unique"`
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	NumEpisodes int
	AirTime     time.Time
}

type GameListElement struct {
	base ListElementFields

	platform string
	release  time.Time
	gameType string
}

type BookListElement struct {
	base ListElementFields

	platform string
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

// TODO: Implement the interface methods for each struct

func CreateListElementFields(url, name, description string, sourceRating float32) ListElementFields {
	var common ListElementFields
	common.URL = url
	common.Name = name
	common.SourceRating = sourceRating
	common.HeuristicRating = float32(math.NaN())
	common.Description = description
	common.IsRated = false

	return common
}
