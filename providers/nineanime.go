// 9Anime.me provider

package providers

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/matheusfillipe/blackbeard/blb"
	"net/url"
)

type NineAnime struct{}


func (a NineAnime) Info() blackbeard.ProviderInfo {
	return blackbeard.ProviderInfo{
		Name: "9anime",
		Url: "https://9anime.vc/",
		Description: "9anime is a free anime website where millions visit to watch anime online.",
	}
}

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

func (a NineAnime) GetEpisodes(shows *blackbeard.Show) []blackbeard.Episode {
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
	return blackbeard.Video{Request: blackbeard.Request{Url: "TODO"}, Format: "mp4"}
}
