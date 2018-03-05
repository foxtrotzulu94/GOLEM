package gol

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/jinzhu/gorm"
)

type AnimeListElement struct {
	ID   int
	Base ListElementFields `gorm:"polymorphic:Owner;"`

	NumEpisodes int
	AirTime     time.Time
}

func (item AnimeListElement) rateElement() float32 {
	lengthModifier := float32(-0.5)
	dateModifier := float32(0.5)

	baseRating := (item.Base.SourceRating * 10.0)
	lengthFactor := (float32(item.NumEpisodes) * lengthModifier)

	// 2006 had some amazing Anime (source: https://www.reddit.com/r/anime/comments/7i41a0/discussion_what_is_the_best_year_of_anime/)
	dateFactor := float32(item.AirTime.Year()-time.Date(2006, 1, 1, 1, 0, 0, 0, time.UTC).Year()) * dateModifier
	fmt.Printf("\tdateFactor %.2f\n", dateFactor)

	return baseRating + lengthFactor + dateFactor
}

func (item AnimeListElement) getRating() float32 {
	return item.Base.HeuristicRating
}

func (item AnimeListElement) getListName() string {
	return "anime"
}

func (item AnimeListElement) getStoredName() string {
	return gorm.ToDBName(reflect.TypeOf(item).Name()) + "s"
}

func (item AnimeListElement) getDerivedID() int {
	return item.ID
}

func (item AnimeListElement) getListElementFields() ListElementFields {
	return item.Base
}

func (item AnimeListElement) printInfo() {
	fmt.Printf("[ID-%03d] (%.2f) \"%s\" - %d Episode(s) - %s\n", item.ID, item.Base.HeuristicRating, item.Base.Name, item.NumEpisodes, item.Base.URL)
}

func (item AnimeListElement) printDetailedInfo() {
	fmt.Printf("[ID-%03d] \"%s\" (%s)\n", item.ID, item.Base.Name, item.Base.URL)
	fmt.Printf("\tHeuristic Rating: %.2f\n", item.Base.HeuristicRating)
	fmt.Printf("\tEpisodes: %d | Base Rating: %.2f \n", item.NumEpisodes, item.Base.SourceRating)

	PrintSetWidth(item.Base.Description, "\t ", "\n", 80)
	fmt.Println("")
}

func (item AnimeListElement) wasFinished() bool {
	return item.Base.WasViewed
}

func (item AnimeListElement) wasRemoved() bool {
	return item.Base.WasRemoved
}

func (item AnimeListElement) updateRating() ListElement {
	fmt.Printf("Element: %s\n", item.Base.Name)
	newItem := determineAppropriateSource(item.Base.URL)(item.Base.URL).(AnimeListElement)

	item.Base.SourceRating = newItem.getListElementFields().SourceRating
	fmt.Printf("\tNew Source Rating: %f\n", item.Base.SourceRating)
	item.Base.HeuristicRating = newItem.rateElement()
	fmt.Printf("\tRecalculated Heuristic: %f\n", item.Base.HeuristicRating)

	item.NumEpisodes = newItem.NumEpisodes
	item.Base.IsRated = true

	return item.saveElement()
}

func (item AnimeListElement) saveElement() ListElement {
	db := getDatabase()
	db.Save(&item)

	return item
}

func (item AnimeListElement) saveOrderedList(list OrderedList) {
	db := getDatabase()
	for _, element := range list {
		listEntry := element.(AnimeListElement)
		db.Create(&listEntry)
	}
}

func (item AnimeListElement) load(derivedID int) ListElement {
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

func (item AnimeListElement) loadOrderedList() OrderedList {
	db := getDatabase()

	var MainList []AnimeListElement
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
