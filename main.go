package main

import (
	"fmt"
	"os"
	"reflect"

	"./gol"
)

func main() {
	cmdArgs := os.Args[1:]
	fmt.Println(cmdArgs)
	gol.Null([]string{})

	//Currently doesn't work
	//functy := gol.SourceMyAnimeList
	fmt.Println(reflect.TypeOf(gol.SourceNull))

	var testy gol.AnimeListElement = gol.SourceMyAnimeList("https://myanimelist.net/anime/71/Full_Metal_Panic").(gol.AnimeListElement)
	gol.PrintAnime(testy)
}
