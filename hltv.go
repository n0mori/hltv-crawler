package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hltv/lib"
	"log"
	"os"
	"sync"
)

func full(fileName string) {
	links := hltv.MatchLinks()

	done := make([]chan bool, 0, 100)

	matches := make([]*hltv.Match, 0, 100)
	matchMutex := &sync.Mutex{}

	file, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, link := range links {
		c := make(chan bool, 1)
		done = append(done, c)

		go func(ch chan bool, url string) {
			m := hltv.MatchData(url)

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

	for _, m := range matches {
		json, _ := json.Marshal(m)
		fmt.Fprintln(file, string(json))
	}

}

func single(fileName, url string) {
	file, err := os.Create(fileName)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	json, _ := json.MarshalIndent(hltv.MatchData(url), "", "\t")
	fmt.Fprint(file, string(json))

}

func main() {
	var fileName, url string
	var opt, help bool

	flag.StringVar(&fileName, "o", "matches.json", "specifies the name of the output file")
	flag.BoolVar(&opt, "single", false, "Extracts data from a single results page")
	flag.StringVar(&url, "url", "https://www.hltv.org/matches/2317926/faze-vs-sk-esl-pro-league-season-6-finals", "Specifies the url for the single option, if not present, a default page will be used")
	flag.BoolVar(&help, "help", false, "shows this dialog")

	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if opt {
		single(fileName, url)
	} else {
		full(fileName)
	}

	/*
		println("Escreva a opção desejada")
		println("1 - Full")
		println("2 - Unico link")

		fmt.Scanln(&opt)

		switch opt {
		case 1:
			full(fileName)
		case 2:
			single(fileName)
		default:
			println("bye!")
		}
	*/
}
