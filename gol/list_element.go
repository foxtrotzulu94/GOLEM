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
	url         string
	name        string
	description string
	isRated     bool

	sourceRating    float32
	heuristicRating float32
}

type AnimeListElement struct {
	base ListElementFields

	numEpisodes int
	airTime     time.Time
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
	if item.base.isRated {
		return item.base.heuristicRating
	}

	lengthFactor := float32(1.5)
	// dateFactor := float32(1.0)
	return (item.base.sourceRating * 10.0) - (float32(item.numEpisodes) * lengthFactor)
}

func (item AnimeListElement) getListName() string {
	return "anime"
}

// TODO: Implement the interface methods for each struct

func CreateListElementFields(url, name, description string, sourceRating float32) ListElementFields {
	var common ListElementFields
	common.url = url
	common.name = name
	common.sourceRating = sourceRating
	common.heuristicRating = float32(math.NaN())
	common.description = description
	common.isRated = false

	return common
}
