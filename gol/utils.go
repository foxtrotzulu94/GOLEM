package gol

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// What list are currently valid names to use in this program

func isValidListName(listName string) bool {
	for name := range RegisteredTypes {
		if name == listName {
			return true
		}
	}

	return false
}

func PrintAnime(object AnimeListElement) {
	fmt.Printf("\"%s\" [%.2f] (%s)\n", object.Base.Name, object.Base.HeuristicRating, object.Base.URL)
	fmt.Printf("Episodes: %d | Rating: %.2f \n", object.NumEpisodes, object.Base.SourceRating)

	maxChars := 80
	length := len(object.Base.Description)
	for i := 0; i < length; i += maxChars {
		var extent int = i + maxChars
		if extent > length { //Slicing beyond length causes an exception. Careful with this
			extent = length
		}
		fmt.Printf("\t%s\n", object.Base.Description[i:extent])
	}
	fmt.Println("")
}

func getListFilename(name string) string {
	return strings.ToUpper(name[0:1]) + name[1:] + ".txt"
}

//Reads the file contents and extracts useful information, line by line
func readFile(filename string) []string {
	//Read all data
	data, err := ioutil.ReadFile(filename)
	check(err)
	plaintext := string(data)
	lines := strings.Split(plaintext, "\r\n")

	// Filter out comments
	validLines := make([]string, 0)
	for _, line := range lines {
		if !strings.Contains(line, "#") {
			validLines = append(validLines, strings.TrimSpace(line))
		}
	}

	return validLines
}

//Makes sure the file is only left with comments
func rewriteFile(filename string) {
	//Read all data
	data, err := ioutil.ReadFile(filename)
	check(err)
	plaintext := string(data)
	lines := strings.Split(plaintext, "\r\n")

	// Filter out comments
	commentedLines := make([]string, 0)
	for _, line := range lines {
		if strings.Contains(line, "#") {
			commentedLines = append(commentedLines, strings.TrimSpace(line))
		}
	}

	//Now, writeback
	_ = os.Remove(filename)
	newFile, _ := os.Create(filename)
	defer newFile.Close()
	for _, line := range commentedLines {
		newFile.WriteString(line)
		//Unfortunately, Windows line endings (just in case)
		newFile.WriteString("\r\n")
	}
}

func ExtractDomainName(URL string) string {
	startIdx := strings.Index(URL, "//") + 2
	endIdx := strings.Index(URL[startIdx:], "/") + startIdx

	//Check if this was invalid
	if startIdx < 2 || endIdx < 2 {
		return ""
	}

	return URL[startIdx:endIdx]
}

func PrintSetWidth(text, linePrefix, newlineSeq string, columnWidth int) {
	length := len(text)
	for i := 0; i < length; i += columnWidth {
		var extent int = i + columnWidth
		if extent > length { //Slicing beyond length causes an exception. Careful with this
			extent = length
		}
		fmt.Print(linePrefix, text[i:extent], newlineSeq)
	}
}

func RequestInput(message string) string {
	fmt.Printf("\n%s", message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return input
}
