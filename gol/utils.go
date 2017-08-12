package gol

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// What list are currently valid names to use in this program
var supportedListNames = []string{"anime" /*"games", "books"*/}

func isValidListName(listName string) bool {
	for _, name := range supportedListNames {
		if name == listName {
			return true
		}
	}

	return false
}

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

func getListFilename(name string) string {
	return strings.ToUpper(name[0:1]) + name[1:] + ".txt"
}

//Reads the file contents and extracts useful information, line by line
func readFile(filename string) []string {
	//Read all data
	data, err := ioutil.ReadFile(filename)
	check(err)
	plaintext := string(data)
	lines := strings.Split(plaintext, "\r\n")

	// Filter out comments
	validLines := make([]string, 0)
	for _, line := range lines {
		if !strings.Contains(line, "#") {
			validLines = append(validLines, strings.TrimSpace(line))
		}
	}

	return validLines
}
