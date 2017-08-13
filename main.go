package main

import (
	"fmt"
	"os"

	"./gol"
)

func main() {
	//cmdArgs := []string{"finished", "anime", ""}
	cmdArgs := os.Args[1:]
	fmt.Println(cmdArgs)
	action := gol.Methods[cmdArgs[0]]

	action(cmdArgs[1:])
	fmt.Println("")
	fmt.Println("Done!")
}
