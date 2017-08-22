package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"./gol"
)

func printUsage() {
	//TODO: Print the available actions!
}

func oneShotMode(cmdArgs []string) {
	if len(cmdArgs) < 1 {
		printUsage()
		return
	}
	fmt.Println(cmdArgs)
	fmt.Println("")

	//Check the list name
	if len(cmdArgs) >= 2 {
		if _, ok := gol.RegisteredTypes[strings.ToLower(cmdArgs[1])]; !ok {
			fmt.Printf("List \"%s\" was not recognized\n", cmdArgs[1])
			return
		}
	}

	//Check the action
	var action gol.ManagerAction
	if actPtr, ok := gol.Actions[strings.ToLower(cmdArgs[0])]; !ok {
		fmt.Printf("Action \"%s\" was not recognized\n", cmdArgs[0])
		fmt.Println("\tAdditional Parameters: ", cmdArgs[1:])
		return
	} else {
		action = actPtr
	}

	action(cmdArgs[1:])
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
	interactiveMode := flag.Bool("interactive", false, "Turn on interactive mode for REPL style functionality")
	flag.Parse()

	if interactiveMode != nil && *interactiveMode {
		interactiveModeLoop()
	} else {
		//cmdArgs := []string{"list", "anime", ""}
		cmdArgs := os.Args[1:]
		oneShotMode(cmdArgs)
	}

	fmt.Println("")
	fmt.Println("Done!")
}
