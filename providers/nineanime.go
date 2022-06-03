// 9Anime.me provider

package providers

import (
	"blackbeard/blb"
	"net/url"
	"github.com/PuerkitoBio/goquery"
)

type NineAnime struct{}

func (a NineAnime) SearchShows(query string) []blackbeard.Show {
	rootUrl := "https://9anime.vc"
	url := rootUrl + "/search?keyword=" + url.QueryEscape(query)

	// Find shows
	var shows []blackbeard.Show
	request := blackbeard.Request{Url: url}
	blackbeard.ScrapePage(request, ".flw-item", func(i int, s *goquery.Selection) {
		aTag := s.Find("a")
		title := aTag.Text()
		href := aTag.AttrOr("href", "")
		shows = append(shows, blackbeard.Show{Url: rootUrl + href, Title: title})
	})
	return shows
}

func (a NineAnime) SearchEpisodes(shows *blackbeard.Show, query string) []blackbeard.Episode {
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

func (a NineAnime) GetVideo(episode *blackbeard.Episode) blackbeard.Video {
	return blackbeard.Video{Url: "TODO", Format: "mp4"}
}
