package gol

import "math"

func ListElementCreator(typeName, url, elementName string, sourceRating float32) *ListElement {
	var common ListElementFields
	common.url = url
	common.name = elementName
	common.sourceRating = sourceRating
	common.heuristicRating = float32(math.NaN())
	common.description = ""
	common.isRated = false

	var retVal *ListElement
	//TODO!
	// switch typeName {
	// case "AnimeListElement":
	// 	retVal := new(AnimeListElement)
	// 	retVal.base = common
	// case "GameListElement":

	// case "BookListElement"
	// }

	return retVal
}
