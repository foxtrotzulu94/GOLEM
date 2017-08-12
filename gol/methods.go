package gol

import (
	"fmt"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//ManagerMethod Defines a type signature for all the manager methods
type ManagerMethod func([]string) int

//Null function
func Null([]string) int {
	return 1
}

func gatherInfo(mainChannel chan ListElement, url string, source InfoSource) {
	listElement := source(url)
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
	if len(fileContents) < 1 {
		fmt.Println("No new records were detected")
		return 0
	}

	mainChannel := make(chan ListElement)
	activeRoutines := 0
	animeList := make([]ListElement, 0)
	animeSet := make(map[string]bool)

	//TODO: CHANGE DYNAMICALLY
	infoSource := SourceMyAnimeList

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
		go gatherInfo(mainChannel, url, infoSource)

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
	if len(animeList) < 1 {
		fmt.Println("No new records were detected after filtering")
		return 0
	}

	//Now sort that list
	safeAnimeList := OrderedList(animeList)
	sort.Sort(sort.Reverse(safeAnimeList))

	fmt.Println("")
	db := getDatabase()
	defer db.Close()

	fmt.Printf("Storing %d new records in Database\n", activeRoutines)
	for _, anime := range safeAnimeList {
		//TODO: ABSTRACT AWAY into database.go or something
		animePtr := anime.(AnimeListElement)
		db.Create(&animePtr)
	}

	fmt.Println("Cleaning up text file...")
	rewriteFile(fileName)

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
