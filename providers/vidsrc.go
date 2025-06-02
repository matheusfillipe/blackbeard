// vidsrc.me provider

package providers

import (
	"fmt"
	"regexp"
	"strings"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/matheusfillipe/blackbeard/blb"

)

type vidsrc struct{}


func (a vidsrc) Info() blackbeard.ProviderInfo {
	return blackbeard.ProviderInfo{
		Name: "vidsrc",
		Url: "https://vidsrc.me/",
		Description: "Movie/TV embedding service",
	}
}

func (a vidsrc) SearchShows(query string) []blackbeard.Show {
	rootUrl := "https://www.themoviedb.org"
	url := rootUrl + "/search?query=" + url.QueryEscape(query)

	// Find shows
	var shows []blackbeard.Show
	request := blackbeard.Request{Url: url, Method: "GET", Debug: true}
	blackbeard.ScrapePage(request, ".title", func(i int, s *goquery.Selection) {
		aTag := s.Find("a")
		if aTag.Text() == "" {
			return
		}
		title := aTag.Text()
		href := aTag.AttrOr("href", "")
		shows = append(shows, blackbeard.Show{Url: rootUrl + href, Title: title})
	})
	return shows
}

func (a vidsrc) GetEpisodes(shows *blackbeard.Show) []blackbeard.Episode {
	rootUrl := "https://vidsrc.xyz"
	re := regexp.MustCompile("\\/(tv|movie)\\/[0-9]*")
	url := rootUrl + "/embed" + re.FindString(shows.Url)
	if strings.Contains(url, "movie") {
		shows.Episodes = append(shows.Episodes, blackbeard.Episode{Title: shows.Title, Url: url})
		return shows.Episodes
	}
	request := blackbeard.Request{Url: url, Debug: true}
	blackbeard.ScrapePage(request, ".ep", func(i int, s *goquery.Selection) {
		title := s.Text()
		href := rootUrl + s.AttrOr("data-iframe", "")
		shows.Episodes = append(shows.Episodes, blackbeard.Episode{Title: title, Url: href, Number: i})
	})
	return shows.Episodes
}

func GetVidUrl(embed string) string {
	var video string

	page := rod.New().MustConnect().MustPage()
	defer page.MustClose()

	wait := page.EachEvent(func(e *proto.NetworkResponseReceived) (stop bool) { 
 				if strings.Contains(e.Response.URL, "m3u8") {
					video = e.Response.URL
					return true
				}
				fmt.Println(e.Response.URL)
				return
			},
		) 
 		page.MustNavigate(embed)
		page.MustWaitStable().MustElement("i").MustClick()
 		wait()
		return video
}

func (a vidsrc) GetVideo(episode *blackbeard.Episode) blackbeard.Video {
	var embed string
	request := blackbeard.Request{Url: episode.Url}
	blackbeard.ScrapePage(request, "iframe", func(i int, s *goquery.Selection) {
		embed = "https:" + s.AttrOr("src", "")
		fmt.Println("%s", embed)
		return
	})

	url := GetVidUrl(embed)
		

	return blackbeard.Video{Name: episode.Title, Request: blackbeard.Request{Url: url}, Format: "m3u8"}
}
