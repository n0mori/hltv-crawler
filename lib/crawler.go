package hltv

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

//Team represents a counter-strike team
type Team struct {
	ID          int
	Name        string
	URL         string
	Nationality string
}

//Match represents a match between the home team and away team
type Match struct {
	ID        int
	URL       string
	Home      Team
	HomeScore int
	Away      Team
	AwayScore int
	EventURL  string
	BestOf    int
	Date      int64
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

//MatchData return a Match pointer for the specified match url
func MatchData(url string) *Match {
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
			URL:         "https://hltv.org" + teamURL,
			ID:          teamID,
			Nationality: nationality}

		scores[index] = score
	})

	divEvent := document.Find(".timeAndEvent").First()

	dateText, _ := divEvent.Find(".time").Attr("data-unix")

	matchDate, err := strconv.ParseInt(dateText, 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	eventLink := divEvent.Find(".event a")

	eventHref, _ := eventLink.Attr("href")

	bestDiv := document.Find(".veto-box").First()
	bestOfText := bestDiv.Children().First().Text()

	br := regexp.MustCompile("^Best of ([0-9])")

	bestOf, err := strconv.Atoi(br.FindStringSubmatch(bestOfText)[1])

	if err != nil {
		log.Fatal(err)
	}

	return &Match{
		ID:        matchID,
		URL:       url,
		Home:      teams[0],
		HomeScore: scores[0],
		Away:      teams[1],
		AwayScore: scores[1],
		Date:      matchDate,
		EventURL:  eventHref,
		BestOf:    bestOf}

}

// MatchLinks returns all match links
func MatchLinks() []string {

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
