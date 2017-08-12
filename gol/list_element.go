package gol

import (
	"fmt"
	"math"
	"time"
)

type ListElement interface {
	rateElement() float32
	getListName() string
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

//TODO: Can't seem to pass by reference in here :/
func (item AnimeListElement) rateElement() float32 {
	fmt.Println("Rating!")
	if item.base.isRated {
		return item.base.heuristicRating
	}

	lengthFactor := float32(1.5)
	// dateFactor := float32(1.0)
	item.base.heuristicRating = (item.base.sourceRating * 10.0) - (float32(item.numEpisodes) * lengthFactor)
	item.base.isRated = true

	return item.base.heuristicRating
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
