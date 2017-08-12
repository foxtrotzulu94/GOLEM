package main

import (
	"fmt"
	"os"

	"./mgr"
)

func main() {
	cmdArgs := os.Args[1:]
	fmt.Println(cmdArgs)
	mgr.Null([]string{})
}
