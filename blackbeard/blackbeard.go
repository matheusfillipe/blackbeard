// Main interface for blackbeard providers

package blackbeard

import (
	"log"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

type Episode struct {
	Title  string
	Number int
	Url    string
}

type Shows struct {
	Title    string
	Url      string
	Episodes []Episode
}

type VideoProvider interface {
	SearchShows(string) []Shows
	SearchEpisodes(*Shows, string) []Episode
}

func ScrapePage(url string, selector string, handler func(int, *goquery.Selection)) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(selector).Each(handler)
}

