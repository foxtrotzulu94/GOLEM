package gol

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

//ManagerAction Defines a type signature for all the manager methods
type ManagerAction func([]string) int

//Null function
func Null([]string) int {
	return 1
}

func validateListName(str string) string {
	listName := strings.ToLower(str)
	if !isValidListName(listName) {
		fmt.Println(listName)
		panic("Given List Name was invalid")
	}

	return listName
}

func requestInput(message string) string {
	fmt.Printf("\n%s", message)
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	return choice
}

func gatherInfo(mainChannel chan ListElement, url string, source InfoSource) {
	listElement := source(url)
	if mainChannel != nil {
		mainChannel <- listElement
	}
}

func scan(args []string) int {
	// Read just one argument: the name of the list
	// This is used throughout in the most generic manner in this function
	listName := validateListName(args[0])

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
	fmt.Printf("Storing %d new records in Database\n", len(sortedElements))
	saveListElements(listName, sortedElements)

	fmt.Println("Cleaning up text file...")
	rewriteFile(fileName)

	return 0
}

func next(args []string) int {
	//Load the required elements and then just pick the first one
	listName := validateListName(args[0])

	orderedList := loadListElements(listName, false, true, true)
	orderedList[0].printInfo()
	return 0
}

func pop(args []string) int {
	//Same as "next" but confirm deletion
	//Load the required elements and then just pick the first one
	listName := validateListName(args[0])

	orderedList := loadListElements(listName, false, true, true)
	orderedList[0].printInfo()

	choice := strings.ToLower(requestInput("Are you sure you want to proceed? (Y/n): "))

	if strings.Contains(choice, "y") {
		modifyListElement(orderedList[0], listName, "WasViewed", true)
		fmt.Println("Marked as finished!")
	} else {
		os.Exit(0)
	}

	return 0
}

func push(args []string) int {
	// This action is slightly simpler than "scan"
	// Just take the named element, rate it and then insert it in the database

	listName := validateListName(args[0])
	newEntry := args[1]

	sourceFunction := determineAppropriateSource(newEntry)
	var listElement ListElement
	if sourceFunction == nil {
		fmt.Println("\"", newEntry, "\"")
		choice := strings.ToLower(requestInput("Cannot determine appropriate source. Add anyway? (Y/n): "))
		if strings.Contains(choice, "n") {
			os.Exit(0)
		}
		//Make a generic list element
		listElement = CreateListElement(listName, newEntry, newEntry, "N/A", 50.0)
	} else {
		fmt.Println("Processing Info Online")
		//Make a small gather info routine
		mainChannel := make(chan ListElement)
		go gatherInfo(mainChannel, newEntry, sourceFunction)
		listElement = <-mainChannel
	}

	listElement.printInfo()
	fmt.Printf("Adding to %s list", listName)
	listElement.saveElement()
	listElement = nil

	return 0
}

func list(args []string) int {
	//Load all active items
	listName := validateListName(args[0])

	orderedList := loadListElements(listName, false, true, true)
	for _, entry := range orderedList {
		entry.printInfo()
	}
	return 0
}

func detail(args []string) int {
	// TODO: finish implementing
	return 1
}

//This is used mostly by the "finished" and "remove" actions
func changeListElementField(args []string, fieldName string, newValue interface{}) {
	//First arg is listName, second is ID
	listName := strings.ToLower(args[0])
	if !isValidListName(listName) {
		fmt.Println(listName)
		panic("Given List Name was invalid")
	}
	listID, _ := strconv.Atoi(args[1])
	entry := getElementByID(listName, listID)
	entry.printInfo()

	if entry.wasFinished() {
		fmt.Println("This entry was previously marked as viewed")
		return
	}
	if entry.wasRemoved() {
		fmt.Println("This entry was removed from the lists entirely")
		return
	}

	fmt.Print("\nAre you sure you want to proceed? (Y/n): ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.ToLower(choice)

	if strings.Contains(choice, "y") {
		modifyListElementFields(listName, "WasViewed", true, listID)
		fmt.Println("Marked as finished!")
	} else {
		os.Exit(0)
	}
}

func finished(args []string) int {
	changeListElementField(args, "WasViewed", true)
	return 0
}

func remove(args []string) int {
	changeListElementField(args, "WasRemoved", true)
	return 0
}

func review(args []string) int {
	listName := args[0]
	filters := strings.ToLower(strings.Join(args[1:], " "))
	if len(filters) < 1 {
		//just list args
		return list(args)
	}

	reviewFinished := strings.Contains(filters, "viewed") || strings.Contains(filters, "finished")
	reviewRemoved := strings.Contains(filters, "removed")
	if !reviewFinished && !reviewRemoved {
		fmt.Println("A valid option was not selected")
	}

	if reviewFinished {
		namedList := loadListElements(listName, true, true, false)
		fmt.Printf("Finished entries in %s: %d\n", listName, len(namedList))
		for _, item := range namedList {
			fmt.Print("\t")
			item.printInfo()
		}
		fmt.Println("")
	}

	if reviewRemoved {
		namedList := loadListElements(listName, true, false, true)
		fmt.Printf("Removed entries in %s: %d\n", listName, len(namedList))
		for _, item := range namedList {
			fmt.Print("\t")
			item.printInfo()
		}
		fmt.Println("")
	}

	return 0
}

//Actions The functions that this program can do.
var Actions = map[string]ManagerAction{
	"scan":     scan,
	"next":     next,
	"push":     push,
	"pop":      pop,
	"list":     list,
	"detail":   detail,
	"finished": finished,
	"remove":   remove,
	"review":   review,
}
