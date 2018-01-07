package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"./gol"
)

func printUsage() {
	fmt.Println("\tUsage:", os.Args[0], "[action] [list name] [options] OR [-i | -interactive]")
	fmt.Println("\tArguments depend on Options being executed")
	fmt.Println("")
	gol.PrintKnownActions()
	fmt.Println("")
	gol.PrintKnownLists()
}

func oneShotMode(cmdArgs []string) {
	if len(cmdArgs) < 1 {
		printUsage()
		return
	}

	fmt.Println(cmdArgs)
	fmt.Println("")

	if strings.EqualFold("help", cmdArgs[0]) {
		gol.PrintKnownActions()
		gol.PrintKnownLists()
		return
	}
	if strings.EqualFold("actions", cmdArgs[0]) {
		gol.PrintKnownActions()
		return
	}

	//Check the list name
	if len(cmdArgs) >= 2 {
		if _, ok := gol.RegisteredTypes[strings.ToLower(cmdArgs[1])]; !ok {
			fmt.Printf("\tList \"%s\" was not recognized\n", cmdArgs[1])
			return
		}
	}

	//Check the action
	var action gol.ManagerAction
	if actPtr, ok := gol.Actions[strings.ToLower(cmdArgs[0])]; !ok {
		fmt.Printf("\tAction \"%s\" was not recognized\n", cmdArgs[0])
		fmt.Println("\tAdditional Parameters: ", cmdArgs[1:])
	} else {
		action = actPtr
		action(cmdArgs[1:])
	}
}

func interactiveModeLoop() {
	for {
		userInput := strings.TrimSpace(gol.RequestInput("> "))

		userWantsToQuit := strings.EqualFold("exit", userInput) || strings.EqualFold("quit", userInput)
		if userWantsToQuit {
			return
		}

		//Otherwise, split the input, check the action and try to run it
		args := strings.Split(strings.Replace(userInput, "\"", "", -1), " ")
		oneShotMode(args)
	}
}

func main() {
	fmt.Println("Golang Ordered List Executive Manager (GOLEM)")
	fmt.Println("Perfect for when you're bored and want some more stuff to do!")

	var interactiveMode = false
	flag.BoolVar(&interactiveMode, "interactive", false, "Turn on interactive mode for REPL style functionality")
	flag.BoolVar(&interactiveMode, "i", false, "Turn on interactive mode for REPL style functionality")
	flag.Parse()

	if interactiveMode {
		interactiveModeLoop()
	} else {
		//cmdArgs := []string{"list", "anime", ""}
		cmdArgs := os.Args[1:]
		oneShotMode(cmdArgs)
	}

	fmt.Println("")
	fmt.Println("Done!")
}
