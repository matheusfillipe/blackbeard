// Main interface for blackbeard providers

package blackbeard

import (
	"log"
	"net/http"
	"net/url"
	"strings"

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

type Request struct {
	Url string
	Method string
	Headers map[string]string
	Body map[string]string
}


func ScrapePage(request Request, selector string, handler func(int, *goquery.Selection)) {
	// Defaults to get request
	method := request.Method
	if method == "" {
		method = "GET"
	}

	_url := request.Url
	headers := request.Headers
	client := &http.Client{}

	data := url.Values{}
	for key, value := range request.Body {
		data.Set(key, value)
	}

	req, _ := http.NewRequest(method, _url, strings.NewReader(data.Encode()))

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res, err := client.Do(req)
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

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
    res := map[K]V{}
    for _, m := range maps {
        for k, v := range m {
            res[k] = v
        }
    }
    return res
}
