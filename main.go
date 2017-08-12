package main

import (
	"fmt"

	"./gol"
)

func main() {
	cmdArgs := []string{"list", "anime"}
	fmt.Println(cmdArgs)
	action := gol.Methods[cmdArgs[0]]

	action(cmdArgs[1:])
	fmt.Println("")
	fmt.Println("Done!")
}
