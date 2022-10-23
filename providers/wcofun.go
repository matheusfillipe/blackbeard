// wcofun.com provider

package providers

import (
	"encoding/base64"
	"fmt"
	"github.com/matheusfillipe/blackbeard/blb"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var wcofunUserAgent = map[string]string{"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:96.0) Gecko/20100101 Firefox/96.0"}

const wcofunRootUrl = "https://www.wcofun.net"

type Wcofun struct{}

type WcofunCdnResponse struct {
	Cdn    string `json:"cdn"`
	Enc    string `json:"enc"`
	Server string `json:"server"`
	Hd     string `json:"hd"`
}


func (a Wcofun) Info() blackbeard.ProviderInfo {
	return blackbeard.ProviderInfo{
		Name: "9anime",
		Url: "https://www.wcofun.com/",
		Description: "Watch Cartoons and Anime Online in HD for Free",
		Cloudflared: true,
	}
}

func (a Wcofun) SearchShows(query string) []blackbeard.Show {
	url := wcofunRootUrl + "/search"

	// Find shows
	var shows []blackbeard.Show

	request := blackbeard.Request{
		Url:     url,
		Method:  "POST",
		Headers: wcofunUserAgent,
		Curl:    true,
		Body: map[string]string{
			"catara":  query,
			"konuara": "series",
		},
	}

	blackbeard.ScrapePage(request, "div.img", func(i int, s *goquery.Selection) {
		title := s.Find("img").AttrOr("alt", "No Title")
		href := s.Find("a").AttrOr("href", "")
		shows = append(shows, blackbeard.Show{Url: href, Title: title})
	})
	for i := range shows {
		blackbeard.EpisodatePopulateShowMetadata(&shows[i])
	}
	return shows
}

func (a Wcofun) GetEpisodes(show *blackbeard.Show) []blackbeard.Episode {
	url := show.Url
	request := blackbeard.Request{
		Url:     url,
		Headers: wcofunUserAgent,
		Curl:    true,
	}

	blackbeard.ScrapePage(request, "#sidebar_right3 a", func(i int, s *goquery.Selection) {
		title := s.AttrOr("title", "No Title")
		href := s.AttrOr("href", "")
		show.Episodes = append(show.Episodes, blackbeard.Episode{Title: title, Url: href, Number: i})
	})

	// Invert episode numbers
	length := len(show.Episodes)
	for i := 0; i < length; i++ {
		show.Episodes[i].Number = length - i - 1
	}
	for i, j := 0, length-1; i < j; i, j = i+1, j-1 {
		show.Episodes[i], show.Episodes[j] = show.Episodes[j], show.Episodes[i]
	}
	blackbeard.EpisodatePopulateEpisodesMetadata(show)
	return show.Episodes
}

func (a Wcofun) GetVideo(episode *blackbeard.Episode) blackbeard.Video {
	url := episode.Url
	request := blackbeard.Request{
		Url:     url,
		Headers: wcofunUserAgent,
		Curl:    true,
	}

	next_path := ""
	blackbeard.ScrapePage(request, "div.fourteen.columns script:nth-child(2)", func(i int, s *goquery.Selection) {
		script := s.Text()
		offset_ := regexp.MustCompile(`^.*- (\d+)\).*$`).FindAllStringSubmatch(script, -1)
		if len(offset_) < 1 {
			return
		}
		offset, err := strconv.Atoi(offset_[0][1])
		if err != nil {
			fmt.Println(err)
			return
		}

		encoded_strings_ := regexp.MustCompile(`^.*\[(.*)\].*$`).FindAllStringSubmatch(script, -1)
		if len(encoded_strings_) < 1 {
			return
		}
		encoded_strings := encoded_strings_[0][1]

		decoded_script := ""
		for _, str := range strings.Split(encoded_strings, ",") {
			// Remove surrounding quotes
			str = strings.ReplaceAll(str, " ", "")
			str = strings.Trim(str, "\"")

			// Decode
			decoded, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Remove characters and leave numbers only
			decoded = regexp.MustCompile(`[^\d]`).ReplaceAll(decoded, []byte(""))
			number, err := strconv.Atoi(string(decoded))
			if err != nil {
				fmt.Println(err)
				return
			}

			decoded_script += string(rune(number - offset))
		}
		blackbeard.Soup(decoded_script, "iframe", func(i int, s *goquery.Selection) {
			next_path = s.AttrOr("src", "")
		})
	})

	if next_path == "" {
		return blackbeard.Video{Request: blackbeard.Request{Url: url}, Format: "mp4"}
	}

	next_path = wcofunRootUrl + next_path

	// Get api url
	request.Url = next_path

	apiPath := ""
	blackbeard.ScrapePage(request, "body", func(i int, s *goquery.Selection) {
		_url := regexp.MustCompile(`.*\$\.getJSON\("(.+?)",.+\).*`).FindAllStringSubmatch(s.Text(), -1)
		if len(_url) < 1 {
			println("failed to get url")
			return
		}
		apiPath = _url[0][1]
	})

	apiPath = wcofunRootUrl + apiPath

	// Request the api
	request.Url = apiPath
	request.Headers["Referer"] = wcofunRootUrl
	request.Headers["authority"] = "www.wcofun.com"
	request.Headers["pragma"] = "no-cache"
	request.Headers["cache-control"] = "no-cache"
	request.Headers["x-requested-with"] = "XMLHttpRequest"
	request.Headers["sec-gpc"] = "1"
	request.Headers["sec-fetch-site"] = "same-origin"
	request.Headers["sec-fetch-mode"] = "cors"
	request.Headers["sec-fetch-dest"] = "empty"
	request.Headers["accept-language"] = "en-US,en;q=0.9"

	data := WcofunCdnResponse{}
	blackbeard.GetJson(request, &data)
	if data.Cdn != "" && data.Enc != "" {
		url = data.Cdn + "/getvid?evid=" + data.Enc
	}

	videoRequest := blackbeard.Request{Url: url, Headers: wcofunUserAgent}
	episode.Video = blackbeard.Video{Request: videoRequest, Format: "mp4", Name: blackbeard.SanitizeFilename(episode.Title) + ".mp4"}
	return episode.Video
}
