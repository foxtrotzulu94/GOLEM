package gol

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
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
	root := retrieveHTML(URL)

	// Define the matcher so it collects all the tags of interest in one pass, we'll sort them out later
	generalMatcher := func(n *html.Node) bool {
		if n == nil {
			return false
		}

		isName := scrape.Attr(n, "itemprop") == "name" && n.DataAtom == atom.Span && n.Parent != nil && n.Parent.Parent != nil && scrape.Attr(n.Parent.Parent, "class") == "blockbg"
		isReviewSpan := strings.Contains(scrape.Attr(n, "class"), "user_reviews_summary_row") || scrape.Attr(n, "id") == "game_area_metascore"
		isDescription := scrape.Attr(n, "class") == "game_description_snippet"
		isReleaseDate := scrape.Attr(n, "class") == "date" && n.DataAtom == atom.Span && n.Parent != nil && scrape.Attr(n.Parent, "class") == "release_date"

		return (isName || isReviewSpan || isDescription || isReleaseDate)
	}

	//Find and iterate
	matches := scrape.FindAll(root, generalMatcher)

	steamRatingMap := map[string]float64{
		"overwhelmingly positive": 99.0,
		"very positive":           95.0,
		"positive":                90.0,
		"mostly positive":         80.0,
		"mixed":                   65.0,
		"mostly negative":         40.0,
		"negative":                25.0,
		"very negative":           15.0,
		"overwhelmingly negative": 5.0,
	}
	var name, description string
	var steamRating, metacriticRating float32
	var releaseDate time.Time

	for _, tagMatch := range matches {
		//Place it accordingly
		stringMatch := scrape.Text(tagMatch)

		classVal := scrape.Attr(tagMatch, "class")
		switch classVal {
		case "date":
			originalRelease, e := time.Parse("2 Jan, 2006", stringMatch)
			check(e)
			releaseDate = originalRelease
		case "game_description_snippet":
			description = stringMatch
		case "user_reviews_summary_row":
			//It's the Steam Rating
			beginTarget := strings.Index(stringMatch, ": ")
			endTarget := strings.Index(stringMatch, "(")
			if beginTarget == -1 || endTarget == -1 {
				panic(URL)
			}

			descriptor := strings.TrimSpace(strings.ToLower(stringMatch[beginTarget+2 : endTarget]))
			//Steam reviews tend to be very high. So we need to shift to sligthly lower
			steamRating = float32(steamRatingMap[descriptor] - 65)

		default:
			if scrape.Attr(tagMatch, "id") == "game_area_metascore" {
				//It's the Metacritic block
				//Get the first child and parse it
				ratingInt, _ := strconv.Atoi(stringMatch[0:2])
				metacriticRating = float32(ratingInt)
			} else if scrape.Attr(tagMatch, "itemprop") == "name" && len(name) < 3 {
				name = stringMatch
			}
		}
	}

	//CHECK that everything was scraped alright
	if len(name) < 1 {
		startIdx := strings.Index(URL, "app/") + len("app/")
		endIdx := strings.Index(URL[startIdx:], "/")
		if endIdx < 0 {
			endIdx = len(URL)
		} else {
			endIdx += startIdx
		}

		appID := URL[startIdx:endIdx]
		infoURL := "http://store.steampowered.com/api/appdetails?appids=" + appID
		resp, err := http.Get(infoURL)
		check(err)
		data, err := ioutil.ReadAll(resp.Body)
		check(err)

		fmt.Println(appID)
		canFallback, err := jsonparser.GetBoolean(data, appID, "success")
		check(err)

		if canFallback {
			newURL, err := jsonparser.GetString(data, appID, "data", "metacritic", "url")

			if !strings.Contains(newURL, "http") {
				fmt.Println("Steam Parse Error on ", URL, err, "- Skipping")
				return nil
			}

			//Fallback to Metacritic
			fmt.Println("Steam Parse Error on ", URL, "- Falling back to Metacritic")
			return SourceMetacritic(newURL)
		}

		fmt.Println("Steam Parse Error on ", URL, "- Skipping")
		return nil
	}

	//Steam reviews tend to be very high. So we need to shift to sligthly lower
	steamWeight, metacriticWeight := float32(0.4), float32(0.8)
	sourceRating := (steamWeight * steamRating) + (metacriticWeight * metacriticRating)

	//Create the object and populate fields
	var retVal GameListElement
	retVal.Base = CreateListElementFields(URL, name, description, sourceRating)
	retVal.Platform = "PC" //All Steam games are for PC
	retVal.ReleaseDate = releaseDate
	retVal.Base.HeuristicRating = retVal.rateElement()
	retVal.Base.IsRated = true

	return retVal
}

func SourceMetacritic(URL string) ListElement {
	root := retrieveHTML(URL)

	// Define the matcher so it collects all the tags of interest in one pass, we'll sort them out later
	generalMatcher := func(n *html.Node) bool {
		if n == nil {
			return false
		}

		itempropVal := scrape.Attr(n, "itemprop")

		isName := itempropVal == "name" && n.DataAtom == atom.Span && n.Parent != nil && n.Parent.Parent != nil && scrape.Attr(n.Parent.Parent, "class") == "product_title"
		isPlatform := itempropVal == "device" && n.Parent != nil && n.Parent.Parent != nil && scrape.Attr(n.Parent.Parent, "class") == "platform"
		isReview := (strings.Contains(scrape.Attr(n, "class"), "metascore_w") && n.Parent != nil && scrape.Attr(n.Parent, "class") == "metascore_anchor") || itempropVal == "ratingValue"
		isDescription := itempropVal == "description"
		isReleaseDate := itempropVal == "datePublished"

		return (isName || isReview || isDescription || isReleaseDate || isPlatform)
	}

	//Find and iterate
	matches := scrape.FindAll(root, generalMatcher)

	var name, description, platform string
	var criticRating, userRating float32
	var releaseDate time.Time

	for _, tagMatch := range matches {
		//Place it accordingly
		stringMatch := scrape.Text(tagMatch)

		itempropVal := scrape.Attr(tagMatch, "itemprop")
		switch itempropVal {
		case "name":
			name = stringMatch
		case "datePublished":
			originalRelease, e := time.Parse("Jan 2, 2006", stringMatch)
			check(e)
			releaseDate = originalRelease
		case "description":
			description = stringMatch
		case "ratingValue":
			rating, _ := strconv.Atoi(stringMatch)
			criticRating = float32(rating)
			fmt.Println("Rating: ", criticRating)
		case "device":
			platform = stringMatch
		default:
			isLikelyUserReview := strings.Contains(scrape.Attr(tagMatch, "class"), "metascore_w") && tagMatch.Parent != nil && tagMatch.Parent.Parent != nil && tagMatch.Parent.Parent.Parent != nil
			if isLikelyUserReview {
				scoreSummaryNode := tagMatch.Parent.Parent.Parent
				if scrape.Attr(scoreSummaryNode, "class") == "score_summary" {
					floatVal, e := strconv.ParseFloat(stringMatch, 32)
					if e != nil {
						fmt.Println(URL)
						panic(e)
					}
					userRating = float32(floatVal * 10.0)
				}
			}
		}
	}

	if len(name) < 1 {
		fmt.Println("Skipping ", URL, "- Parse Error")
		return nil
	}

	sourceRating := criticRating + userRating
	for sourceRating > 100 {
		sourceRating = sourceRating / 2
	}

	//Create the object and populate fields
	var retVal GameListElement
	retVal.Base = CreateListElementFields(URL, name, description, sourceRating)
	retVal.Platform = platform
	retVal.ReleaseDate = releaseDate
	retVal.Base.HeuristicRating = retVal.rateElement()
	retVal.Base.IsRated = true

	return retVal
}

func SourceAmazonUS(URL string) ListElement {
	//TODO:
	return nil
}

func SourceAmazonCanada(URL string) ListElement {
	//TODO:
	return nil
}

func SourceIMDB(URL string) ListElement {
	root := retrieveHTML(URL)

	// Define the matcher so it collects all the tags of interest in one pass, we'll sort them out later
	generalMatcher := func(n *html.Node) bool {
		if n == nil {
			return false
		}

		itempropVal := scrape.Attr(n, "itemprop")

		isName := itempropVal == "name" && n.DataAtom == atom.H1
		isDescription := itempropVal == "description"
		isScore := itempropVal == "ratingValue"
		isViews := itempropVal == "ratingCount"

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
