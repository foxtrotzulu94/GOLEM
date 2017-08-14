package gol

import (
	"fmt"
	"time"
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

func (item AnimeListElement) printInfo() {
	fmt.Printf("[ID-%03d] (%.2f) \"%s\" - %d Episode(s) - %s\n", item.ID, item.Base.HeuristicRating, item.Base.Name, item.NumEpisodes, item.Base.URL)
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
