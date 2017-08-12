package main

import (
	"./gol"
)

func main() {
	//cmdArgs := os.Args[1:]
	cmdArgs := []string{"scan", "anime"}
	//fmt.Println(cmdArgs)
	action := gol.Methods[cmdArgs[0]]

	//gol.Null([]string{})

	//var testy gol.AnimeListElement = gol.SourceMyAnimeList("https://myanimelist.net/anime/71/Full_Metal_Panic").(gol.AnimeListElement)
	//gol.PrintAnime(testy)

	action(cmdArgs[1:])
}
