package hltv

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//Player represents a player that has played in a match.
type Player struct {
	URL         string
	ID          int
	Name        string
	Nationality string
	Kills       int
	Deaths      int
	ADR         float64
	KAST        float64
	Rating      float64
}

//Team represents a counter-strike team
type Team struct {
	ID          int
	Name        string
	URL         string
	Nationality string
	Players     []Player
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
	println(url)

	teams := make([]Team, 2)
	r := regexp.MustCompile("/([0-9]+)/")
	matchID, _ := strconv.Atoi(r.FindStringSubmatch(url)[1])
	scores := make([]int, 5)

	divTeam := document.Find("div.team")
	divTeam.Each(func(index int, element *goquery.Selection) {
		if index > 1 {
			return
		}

		nationality, _ := element.Find("img").First().Attr("title")
		name := element.Find(".teamName").Text()
		teamURL, _ := element.Find("div a").Attr("href")
		teamID, _ := strconv.Atoi(r.FindStringSubmatch(teamURL)[1])
		score, _ := strconv.Atoi(element.Find("div").Last().Text())

		teams[index] = Team{
			Name:        name,
			URL:         "https://hltv.org" + teamURL,
			ID:          teamID,
			Nationality: nationality}

		scores[index] = score
	})

	nameRegex := regexp.MustCompile("/([^/]*)$")
	divScoreboard := document.Find("#all-content")
	divScoreboard.Find("table").Each(func(index int, element *goquery.Selection) {
		players := make([]Player, 5)

		element.Find("tr").Each(func(i int, e *goquery.Selection) {
			if i == 0 || i >= 6 {
				return
			}

			i--

			url := "https://hltv.org" + e.Find(".players a").AttrOr("href", "")
			name := nameRegex.FindStringSubmatch(url)[1]
			id, _ := strconv.Atoi(r.FindStringSubmatch(url)[1])
			nation := e.Find(".flag").AttrOr("title", "")
			kd := strings.Split(e.Find(".kd").Text(), "-")
			kills, _ := strconv.Atoi(kd[0])
			deaths, _ := strconv.Atoi(kd[1])
			adr, _ := strconv.ParseFloat(e.Find(".adr").Text(), 64)
			kast, _ := strconv.ParseFloat(strings.Trim(e.Find(".kast").Text(), "%"), 64)
			rating, _ := strconv.ParseFloat(e.Find(".rating").Text(), 64)

			player := Player{
				URL:         url,
				Name:        name,
				ID:          id,
				Nationality: nation,
				Kills:       kills,
				Deaths:      deaths,
				ADR:         adr,
				KAST:        kast,
				Rating:      rating}

			players[i] = player
		})

		teams[index].Players = players
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
		EventURL:  "https://hltv.com" + eventHref,
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
