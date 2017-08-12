package main

import (
	"fmt"
	"log"

	"./gol"
	"github.com/go-xorm/xorm"
)

// ORM engine
var x *xorm.Engine

func init() {
	// Create ORM engine and database
	var err error
	x, err = xorm.NewEngine("sqlite3", "./bank.db")
	if err != nil {
		log.Fatalf("Fail to create engine: %v\n", err)
	}

	// Sync tables
	x.Sync(new(gol.ListElementFields))
	if err = x.Sync(new(gol.AnimeListElement)); err != nil {
		log.Fatalf("Fail to sync database: %v\n", err)
	}
	fmt.Println("Finished setting up DB")
}

func main() {
	//cmdArgs := os.Args[1:]
	cmdArgs := []string{"scan", "anime"}
	//fmt.Println(cmdArgs)
	action := gol.Methods[cmdArgs[0]]

	//gol.Null([]string{})

	//var testy gol.AnimeListElement = gol.SourceMyAnimeList("https://myanimelist.net/anime/71/Full_Metal_Panic").(gol.AnimeListElement)
	//gol.PrintAnime(testy)
	gol.DB = x
	action(cmdArgs[1:])
}
