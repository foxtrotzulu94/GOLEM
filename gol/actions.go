package gol

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//ManagerAction Defines a type signature for all the manager methods
type ManagerAction func([]string) int

func validateListName(str string) string {
	listName := strings.ToLower(str)
	if !isValidListName(listName) {
		fmt.Println(listName)
		panic("Given List Name was invalid")
	}

	return listName
}

func gatherInfo(mainChannel chan ListElement, url string, source InfoSource) {
	listElement := source(url)
	if mainChannel != nil {
		mainChannel <- listElement
	}
}

func scan(args []string) int {
	//If no list is specified, just do all of them!
	if len(args) < 1 {
		var synch sync.WaitGroup
		synch.Add(len(RegisteredTypes))

		for listName := range RegisteredTypes {
			go func(activeList string) {
				defer synch.Done()
				scan([]string{activeList})
			}(listName)
		}

		synch.Wait()
		fmt.Println("All registered types were scanned")
		return 0
	}

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
	sortedElements.save()

	fmt.Println("Cleaning up text file...")
	//rewriteFile(fileName)

	return 0
}

func next(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager next <list name>")
		PrintKnownLists()
		return 1
	}

	//Load the required elements and then just pick the first one
	listName := validateListName(args[0])

	orderedList := loadListElements(listName, false, true, true)
	orderedList[0].printInfo()
	return 0
}

func pop(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager pop <list name>")
		PrintKnownLists()
		return 1
	}

	//Same as "next" but confirm deletion
	//Load the required elements and then just pick the first one
	listName := validateListName(args[0])

	orderedList := loadListElements(listName, false, true, true)
	orderedList[0].printInfo()

	choice := strings.ToLower(RequestInput("Are you sure you want to proceed? (Y/n): "))

	if strings.Contains(choice, "y") {
		modifyListElementFields(orderedList[0], listName, "WasViewed", true)
		fmt.Println("Marked as finished!")
	} else {
		os.Exit(0)
	}

	return 0
}

func push(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager push|add <list name> <URL>")
		PrintKnownLists()
		return 1
	}

	// This action is slightly simpler than "scan"
	// Just take the named element, rate it and then insert it in the database

	listName := validateListName(args[0])
	newEntry := args[1]

	sourceFunction := determineAppropriateSource(newEntry)
	var listElement ListElement
	if sourceFunction == nil {
		fmt.Println("\"", newEntry, "\"")
		choice := strings.ToLower(RequestInput("Cannot determine appropriate source. Add anyway? (Y/n): "))
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

	if listElement == nil {
		return 1
	}

	listElement = listElement.saveElement()
	listElement.printInfo()
	fmt.Printf("Added to %s list", listName)
	listElement = nil

	return 0
}

func list(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager list <list name>")
		PrintKnownLists()
		return 1
	}

	//Load all active items
	listName := validateListName(args[0])

	orderedList := loadListElements(listName, false, true, true)
	for i, entry := range orderedList {
		fmt.Printf("%4d.  ", i+1)
		entry.printInfo()
	}
	return 0
}

func detail(args []string) int {
	if len(args) < 2 {
		fmt.Println("\tUsage: manager detail|view|info <list name> <ID>")
		PrintKnownLists()
		return 1
	}

	listName := validateListName(args[0])
	listID, _ := strconv.Atoi(args[1])

	entry := getElementByID(listName, listID)
	if entry == nil {
		fmt.Printf("Entry with ID %d not found in %s list", listID, listName)
		return 1
	}

	entry.printDetailedInfo()

	return 0
}

//This is used mostly by the "finished" and "remove" actions
func changeListElementField(args []string, fieldName string, newValue interface{}) {
	//First arg is listName, second is ID
	listName := validateListName(args[0])
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

	choice := strings.ToLower(RequestInput("Are you sure you want to proceed? (Y/n): "))

	if strings.Contains(choice, "y") {
		modifyListElementFields(entry, listName, "WasViewed", true)
		fmt.Println("Marked as finished!")
	} else {
		os.Exit(0)
	}
}

func finished(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager finished|finish <list name> <ID>")
		PrintKnownLists()
		return 1
	}

	changeListElementField(args, "WasViewed", true)
	return 0
}

func remove(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager remove|delete <list name> <ID>")
		PrintKnownLists()
		return 1
	}

	changeListElementField(args, "WasRemoved", true)
	return 0
}

func review(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager review|check <list name> [viewed|finished|removed]")
		PrintKnownLists()
		return 1
	}

	listName := args[0]
	filters := strings.ToLower(strings.Join(args[1:], " "))
	if len(filters) < 1 {
		//just list args
		return list(args)
	}
	//FIXME: Maybe error out when removed/viewed/finished is miss spelled
	reviewFinished := strings.Contains(filters, "viewed") || strings.Contains(filters, "finished")
	reviewRemoved := strings.Contains(filters, "removed")
	if !reviewFinished && !reviewRemoved {
		return list(args)
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

func search(args []string) int {
	if len(args) != 2 {
		fmt.Println("\tUsage: manager search|find <list name> <string>")
		PrintKnownLists()
		return 1
	}

	listName := validateListName(args[0])
	keyword := strings.ToLower(strings.TrimSpace(args[1]))

	orderedList := loadListElements(listName, false, false, false)
	//Do a linear search
	var results int
	for _, entry := range orderedList {
		if strings.Contains(strings.ToLower(entry.getListElementFields().Name), keyword) {
			entry.printInfo()
			results++
		}
	}

	if results == 0 {
		fmt.Println("No results found")
	}

	return 0
}

func reconsider(args []string) int {
	if len(args) < 1 {
		fmt.Println("\tUsage: manager search|find <list name>")
		PrintKnownLists()
		return 1
	}

	//Reload and rerate whatever active elements were in the list
	listName := validateListName(args[0])
	orderedList := loadListElements(listName, false, true, true)

	fmt.Println("Reconsidering all active", listName)
	var synch sync.WaitGroup
	synch.Add(len(orderedList))
	for _, activeListEntry := range orderedList {
		go func(entry ListElement) {
			defer synch.Done()

			modifyListElementFields(entry, listName, "HeuristicRating", entry.rateElement())

		}(activeListEntry)
	}

	synch.Wait()
	fmt.Println("All registered types were rerated")

	return list(args)
}

//Actions The functions that this program can do.
var Actions = map[string]ManagerAction{
	"scan":       scan,
	"load":       scan,
	"next":       next,
	"push":       push,
	"add":        push,
	"pop":        pop,
	"list":       list,
	"detail":     detail,
	"view":       detail,
	"info":       detail,
	"finished":   finished,
	"finish":     finished,
	"remove":     remove,
	"delete":     remove,
	"review":     review,
	"check":      review,
	"find":       search,
	"search":     search,
	"lists":      enumerate,
	"enumerate":  enumerate,
	"reconsider": reconsider,
	"reorganize": reconsider,
	"rate":       reconsider,
}

func enumerate(args []string) int {
	PrintKnownLists()
	return 0
}

func functions(args []string) int {
	PrintKnownActions()
	return 0
}
