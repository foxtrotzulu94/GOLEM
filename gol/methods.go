package gol

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

//ManagerMethod Defines a type signature for all the manager methods
type ManagerMethod func([]string) int

var DB *xorm.Engine = nil

//Null function
func Null([]string) int {
	return 1
}

func gatherInfo(mainChannel chan ListElement, url string) {
	listElement := SourceMyAnimeList(url)
	mainChannel <- listElement
}

func scan(args []string) int {
	// Read just one argument: the name of the list
	listName := strings.ToLower(args[0])
	if !isValidListName(listName) {
		fmt.Println(listName)
		panic("Given List Name was invalid")
	}

	fileName := getListFilename(listName)
	fileContents := readFile(fileName)
	mainChannel := make(chan ListElement)
	activeRoutines := 0
	animeList := make([]ListElement, 0)
	animeSet := make(map[string]bool)

	//Spawn all the go routines
	for _, url := range fileContents {
		//Avoid requesting something that's NOT a URL
		if !strings.Contains(url, "http") {
			continue
		}

		// Avoid requesting a previously seen URL
		if _, ok := animeSet[url]; ok {
			fmt.Println("Duplicate " + url)
			continue
		}

		//Send off the request concurrently
		go gatherInfo(mainChannel, url)

		//go gatherInfo(mainChannel, url)
		animeSet[url] = true
		activeRoutines++
	}
	//Wait for them to come back in order
	for i := 0; i < activeRoutines; i++ {
		animePtr := <-mainChannel
		// animeObj := *animePtr
		if animePtr != nil {
			animeList = append(animeList, animePtr)
		}
	}

	//Now sort that list
	safeAnimeList := OrderedList(animeList)
	sort.Sort(sort.Reverse(safeAnimeList))

	fmt.Println("")

	var animeEpisodeCount = 0
	for i, anime := range safeAnimeList {
		fmt.Printf("%d. ", i+1)
		animePtr := anime.(AnimeListElement)
		DB.Insert(animePtr)
		PrintAnime(animePtr)
		animeEpisodeCount += anime.(AnimeListElement).NumEpisodes
	}

	return 0
}

func next(args []string) int {
	return 1
}

func pop(args []string) int {
	return 1
}

func push(args []string) int {
	return 1
}

func list(args []string) int {
	return 1
}

func detail(args []string) int {
	return 1
}

func remove(args []string) int {
	return 1
}

func review(args []string) int {
	return 1
}

var Methods = map[string]ManagerMethod{
	"scan":   scan,
	"next":   next,
	"push":   push,
	"pop":    pop,
	"list":   list,
	"detail": detail,
	"remove": remove,
	"review": review,
}
