package main

import (
	"encoding/json"
	"fmt"
	"hltv/lib"
	"os"
	"sync"
)

func full() {
	links := hltv.MatchLinks()

	done := make([]chan bool, 0, 100)

	matches := make([]*hltv.Match, 0, 100)
	matchMutex := &sync.Mutex{}

	file, err := os.OpenFile("matches.json", os.O_CREATE|os.O_WRONLY, 0666)

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
		json, _ := json.MarshalIndent(m, "", "\t")
		fmt.Fprintln(file, string(json))
	}

}

func single() {
	var url string

	println("Escreva a URL:")
	fmt.Scanln(&url)

	json, _ := json.MarshalIndent(hltv.MatchData(url), "", "\t")
	println(string(json))

}

func main() {
	var opt int

	println("Escreva a opção desejada")
	println("1 - Full")
	println("2 - Unico link")

	fmt.Scanln(&opt)

	switch opt {
	case 1:
		full()
	case 2:
		single()
	default:
		println("bye!")
	}

}
