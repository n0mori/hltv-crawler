package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

//Team represents a counter-strike team
type Team struct {
	Id          int
	Name        string
	Url         string
	Nationality string
}

//Match represents a match between the home team and away team
type Match struct {
	Id        int
	Url       string
	Home      Team
	HomeScore int
	Away      Team
	AwayScore int
}

func getDocument(url string) *goquery.Document {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Error in goquery.", err)
	}

	return document
}

func matchData(url string) *Match {
	document := getDocument(url)

	teams := make([]Team, 2)
	r := regexp.MustCompile("/([0-9]+)/")
	matchID, _ := strconv.Atoi(r.FindStringSubmatch(url)[1])
	scores := make([]int, 2)

	divTeam := document.Find("div.team")
	divTeam.Each(func(index int, element *goquery.Selection) {

		nationality, _ := element.Find("img").First().Attr("title")
		name := element.Find(".teamName").Text()
		teamURL, _ := element.Find("div a").Attr("href")
		teamID, _ := strconv.Atoi(r.FindStringSubmatch(url)[1])
		score, _ := strconv.Atoi(element.Find("div").Last().Text())

		teams[index] = Team{
			Name:        name,
			Url:         "https://hltv.org" + teamURL,
			Id:          teamID,
			Nationality: nationality}

		scores[index] = score
	})

	return &Match{
		Id:        matchID,
		Url:       url,
		Home:      teams[0],
		HomeScore: scores[0],
		Away:      teams[1],
		AwayScore: scores[1]}

}

func matchLinks() []string {

	document := getDocument("https://hltv.org/results?content=stats&stars=1")

	links := make([]string, 0, 100)

	document.Find(".result-con a").Each(func(index int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if exists {
			links = append(links, "https://hltv.org"+href)
		}
	})

	return links
}

func main() {
	links := matchLinks()

	done := make([]chan bool, 0, 100)

	matches := make([]*Match, 0, 100)
	matchMutex := &sync.Mutex{}

	file, err := os.OpenFile("matches", os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, link := range links {
		c := make(chan bool, 1)
		done = append(done, c)

		go func(ch chan bool, url string) {
			m := matchData(url)
			json, _ := json.Marshal(m)
			fmt.Fprintln(file, string(json))

			matchMutex.Lock()
			matches = append(matches, m)
			matchMutex.Unlock()

			ch <- true

		}(c, link)

		//println(link)
	}

	for _, c := range done {
		<-c
	}
}
