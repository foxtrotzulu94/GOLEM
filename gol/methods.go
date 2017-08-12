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
	// This is used throughout in the most generic manner in this function
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
	entries := make([]ListElement, 0)
	entrySet := make(map[string]bool)

	//Spawn all the go routines
	for _, url := range fileContents {
		//Avoid requesting something that's NOT a URL
		if !strings.Contains(url, "http") {
			continue
		}

		// Avoid requesting a previously seen URL
		if _, ok := entrySet[url]; ok {
			fmt.Println("Duplicate " + url)
			continue
		}

		//Send off the request concurrently
		infoSource := determineAppropriateSource(url)
		go gatherInfo(mainChannel, url, infoSource)

		entrySet[url] = true
		activeRoutines++
	}
	//Wait for them to come back in order
	for i := 0; i < activeRoutines; i++ {
		listElement := <-mainChannel
		if listElement != nil {
			entries = append(entries, listElement)
		}
	}
	if len(entries) < 1 {
		fmt.Println("No new records were detected after filtering")
		return 0
	}

	//Now sort that list
	sortedElements := OrderedList(entries)
	sort.Sort(sort.Reverse(sortedElements))

	fmt.Println("")
	fmt.Printf("Storing %d new records in Database\n", activeRoutines)
	saveListElements(listName, sortedElements)

	fmt.Println("Cleaning up text file...")
	//rewriteFile(fileName)

	return 0
}

func next(args []string) int {
	//Load the required elements and then just pick the first one
	listName := strings.ToLower(args[0])
	if !isValidListName(listName) {
		fmt.Println(listName)
		panic("Given List Name was invalid")
	}

	orderedList := LoadListElements(listName, true, true)
	orderedList[0].printInfo()
	return 0
}

func pop(args []string) int {
	//Same as "next" but confirm deletion

	return 1
}

func push(args []string) int {
	return 1
}

func list(args []string) int {
	//Load all active items
	listName := strings.ToLower(args[0])
	if !isValidListName(listName) {
		fmt.Println(listName)
		panic("Given List Name was invalid")
	}

	orderedList := LoadListElements(listName, true, true)
	for _, entry := range orderedList {
		entry.printInfo()
	}
	return 0
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
