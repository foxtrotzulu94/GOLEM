package main

import (
	"fmt"
	"os"

	"./gol"
)

func main() {
	//cmdArgs := []string{"list", "anime", ""}
	cmdArgs := os.Args[1:]
	fmt.Println(cmdArgs)
	fmt.Println("")
	action := gol.Actions[cmdArgs[0]]

	action(cmdArgs[1:])
	fmt.Println("")
	fmt.Println("Done!")
}
