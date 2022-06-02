// wcofun.com provider

package providers

import (
	"blackbeard/blb"
	"github.com/PuerkitoBio/goquery"
)

var UserAgent = map[string]string{"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:96.0) Gecko/20100101 Firefox/96.0"}

type Wcofun struct{}

func (a Wcofun) SearchShows(query string) []blackbeard.Shows {
	rootUrl := "https://www.wcofun.com"
	_url := rootUrl + "/search"

	// Find shows
	var shows []blackbeard.Shows

	request := blackbeard.Request{
		Url: _url,
		Method: "POST",
		Headers: UserAgent,
		Curl: true,
		Body: map[string]string{
			"catara": query,
			"konuara": "series",
		},
	}

	blackbeard.ScrapePage(request, ".flw-item", func(i int, s *goquery.Selection) {
		aTag := s.Find("a")
		title := aTag.Text()
		href := aTag.AttrOr("href", "")
		shows = append(shows, blackbeard.Shows{Url: rootUrl + href, Title: title})
	})
	return shows
}

func (a Wcofun) SearchEpisodes(shows *blackbeard.Shows, query string) []blackbeard.Episode {
	url := shows.Url
	request := blackbeard.Request{Url: url}
	blackbeard.ScrapePage(request, ".episodes-ul", func(i int, s *goquery.Selection) {
		aTag := s.Find("a")
		title := aTag.Text()
		href := aTag.AttrOr("href", "")
		shows.Episodes = append(shows.Episodes, blackbeard.Episode{Title: title, Url: href, Number: i})
	})
	return shows.Episodes
}
