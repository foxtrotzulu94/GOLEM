package gol

import "fmt"

func PrintAnime(object AnimeListElement) {
	fmt.Printf("\"%s\" [%.2f] (%s)\n", object.base.name, object.base.heuristicRating, object.base.url)
	fmt.Printf("Episodes: %d | Rating: %.2f \n", object.numEpisodes, object.base.sourceRating)

	maxChars := 80
	length := len(object.base.description)
	for i := 0; i < length; i += maxChars {
		var extent int = i + maxChars
		if extent > length { //Slicing beyond length causes an exception. Careful with this
			extent = length
		}
		fmt.Printf("\t%s\n", object.base.description[i:extent])
	}
	fmt.Println("")
}
