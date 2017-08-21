package gol

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//InfoSource Defines the type signature of an information source that can be used to create and rate new ListElements
type InfoSource func(string) ListElement

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func retrieveHTML(URL string) *html.Node {
	// Request the page
	resp, err := http.Get(URL)
	check(err)
	root, err := html.Parse(resp.Body)
	check(err)

	return root
}

func SourceNull(URL string) ListElement {
	return nil
}

func SourceMyAnimeList(URL string) ListElement {
	root := retrieveHTML(URL)

	// Define the matcher so it collects all the tags of interest in one pass, we'll sort them out later
	generalMatcher := func(n *html.Node) bool {
		if n == nil {
			return false
		}

		isName := scrape.Attr(n, "itemprop") == "name" && n.Parent != nil && n.Parent.DataAtom == atom.H1
		isDescription := scrape.Attr(n, "itemprop") == "description"
		isNumEpisodes := (scrape.Attr(n, "class") == "spaceit" && strings.Contains(scrape.Text(n), "Episodes:"))
		isScore := scrape.Attr(n, "data-title") == "score"

		return (isName || isDescription || isNumEpisodes || isScore)
	}

	//Find and iterate
	matches := scrape.FindAll(root, generalMatcher)

	var name, description string
	var numEpisodes int
	var sourceRating float32
	for _, tagMatch := range matches {
		//Place it accordingly
		stringMatch := scrape.Text(tagMatch)
		if strings.Contains(stringMatch, "Episodes: ") {
			numEpisodes, _ = strconv.Atoi(stringMatch[len("Episodes: "):])
		} else {
			if scrape.Attr(tagMatch, "data-title") == "score" {
				floatVal, e := strconv.ParseFloat(stringMatch, 32)
				if e != nil {
					fmt.Println(URL)
					panic(e)
				}
				sourceRating = float32(floatVal)
				if floatVal < 7.00 {
					//Fuck it, I'm not watching bad anime.
					fmt.Println("Discarded " + URL)
					return nil
				}
			} else {
				itempropVal := scrape.Attr(tagMatch, "itemprop")
				if itempropVal == "name" {
					name = stringMatch
				} else if itempropVal == "description" {
					description = stringMatch
				}
			}
		}
	}

	//Create the object and populate fields
	var retVal AnimeListElement
	retVal.NumEpisodes = numEpisodes
	retVal.Base = CreateListElementFields(URL, name, description, sourceRating)
	retVal.Base.HeuristicRating = retVal.rateElement()
	retVal.Base.IsRated = true
	return retVal
}

func SourceSteamOnline(URL string) ListElement {
	//TODO:
	panic("Not Implemented Yet")
}

func SourceMetacritic(URL string) ListElement {
	//TODO:
	panic("Not Implemented Yet")
}

func SourceAmazonUS(URL string) ListElement {
	//TODO:
	panic("Not Implemented Yet")
}

func SourceAmazonCanada(URL string) ListElement {
	//TODO:
	panic("Not Implemented Yet")
}

func SourceIMDB(URL string) ListElement {
	root := retrieveHTML(URL)

	// Define the matcher so it collects all the tags of interest in one pass, we'll sort them out later
	generalMatcher := func(n *html.Node) bool {
		if n == nil {
			return false
		}

		isName := scrape.Attr(n, "itemprop") == "name" && n.DataAtom == atom.H1
		isDescription := scrape.Attr(n, "itemprop") == "description"
		isScore := scrape.Attr(n, "itemprop") == "ratingValue"
		isViews := scrape.Attr(n, "itemprop") == "ratingCount"

		return (isName || isDescription || isScore || isViews)
	}

	//Find and iterate
	matches := scrape.FindAll(root, generalMatcher)

	var name, description string
	var sourceRating float32
	var count int
	for _, tagMatch := range matches {
		//Place it accordingly
		stringMatch := scrape.Text(tagMatch)

		itempropVal := scrape.Attr(tagMatch, "itemprop")
		switch itempropVal {
		case "name":
			name = stringMatch
		case "description":
			description = stringMatch
		case "ratingValue":
			floatVal, _ := strconv.ParseFloat(stringMatch, 32)
			if floatVal == 0 {
				continue
			}
			sourceRating = float32(floatVal)
		case "ratingCount":
			reviews, _ := strconv.Atoi(stringMatch)
			if reviews > 0 && count == 0 {
				count = reviews
			}
		}
	}

	//Create the object and populate fields
	var retVal MovieListElement
	retVal.Base = CreateListElementFields(URL, name, description, sourceRating)
	retVal.ReviewCount = count
	retVal.Base.HeuristicRating = retVal.rateElement()
	retVal.Base.IsRated = true

	return retVal
}

var Sources = map[string]InfoSource{
	"myanimelist.net":        SourceMyAnimeList,
	"store.steampowered.com": SourceSteamOnline,
	"www.amazon.ca":          SourceAmazonCanada,
	"amazon.com":             SourceAmazonUS,
	"www.metacritic.com":     SourceMetacritic,
	"www.imdb.com":           SourceIMDB,
}

func determineAppropriateSource(URL string) InfoSource {
	domainName := ExtractDomainName(URL)

	//1st try is straight up matching
	if val, ok := Sources[domainName]; ok {
		return val
	}

	//2nd try is substring matching
	for key := range Sources {
		if strings.Contains(URL, key) {
			return Sources[key]
		}
	}

	//no 3rd try, just return nil
	return nil
}
