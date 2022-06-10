// 1337x.wtf provider

package providers

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/matheusfillipe/blackbeard/blb"
	"net/url"
)

type Leetx struct{}

const leetUrl = "https://1337x.wtf"


func (a Leetx) Info() blackbeard.ProviderInfo {
	return blackbeard.ProviderInfo{
		Name: "1337x",
		Url: "https://1337x.wtf/",
		Description: "Search and find torrents for movies, tv shows and animes",
	}
}

func (a Leetx) SearchShows(query string) []blackbeard.Show {
	url := leetUrl + "/search/" + url.PathEscape(query) + "/1/"

	// Find shows
	var shows []blackbeard.Show
	request := blackbeard.Request{
		Url: url,
		Curl: true,
	}
	hasAny := false
	blackbeard.ScrapePage(request, ".table-list > tbody:nth-child(2) > tr", func(i int, s *goquery.Selection) {
		hasAny = true
	})
	if hasAny {
		shows = append(shows, blackbeard.Show{Url: url, Title: query})
	}
	return shows
}

func (a Leetx) GetEpisodes(shows *blackbeard.Show) []blackbeard.Episode {
	url := shows.Url
	request := blackbeard.Request{
		Url: url,
		Curl: true,
	}
	blackbeard.ScrapePage(request, ".table-list > tbody:nth-child(2) > tr", func(i int, s *goquery.Selection) {
		aTag := s.Find("a:nth-child(2)")
		href := aTag.AttrOr("href", "")
		title := aTag.Text()
		req := blackbeard.Request{
			Url: leetUrl + href,
			Curl: true,
		}
		description := ""
		magnetic := ""
		blackbeard.ScrapePage(req, "body", func(i int, ss *goquery.Selection) {
			found := false
			ss.Find("a").Each(func(i int, sss *goquery.Selection) {
				if found {
					return
				}
				if sss.Text() == "Magnet Download" {
					magnetic = sss.AttrOr("href", "")
					found = true
				}
			})
			ss.Find("div.clearfix > ul.list").Each(func(i int, ssss *goquery.Selection){
				description += ssss.Text()
			})
		})

		shows.Episodes = append(shows.Episodes, blackbeard.Episode{
			Title: title,
			Url: magnetic,
			Number: i,
			Metadata: blackbeard.Metadata{
				Description: description,
			},
		})
	})
	return shows.Episodes
}

func (a Leetx) GetVideo(episode *blackbeard.Episode) blackbeard.Video {
	return blackbeard.Video{
		Request: blackbeard.Request{
			Url: episode.Url,
		},
		Format: "magnetic",
	}
}
