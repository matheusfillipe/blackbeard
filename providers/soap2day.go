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

var soapUserAgent = map[string]string{"User-Agent": "User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:100.0) Gecko/20100101 Firefox/100.0"}

const soapRootUrl = "https://ww1.ssoap2day.to"

type Soap2day struct{}

type CdnResponse struct {
	Cdn    string `json:"cdn"`
	Enc    string `json:"enc"`
	Server string `json:"server"`
	Hd     string `json:"hd"`
}

func (a Soap2day) Info() blackbeard.ProviderInfo {
	return blackbeard.ProviderInfo{
		Name:        "soap2day",
		Url:         "https://ssoap2day.to/",
		Description: "Soap2day is a website with a vast number of movies to watch on soap2day. This online platform is specifically designed to meet all your movie cravings at Soap2day.",
	}
}

func (a Soap2day) SearchShows(query string) []blackbeard.Show {
	url := soapRootUrl + "/index.php?do=search"

	// Find shows
	var shows []blackbeard.Show

	request := blackbeard.Request{
		Url:     url,
		Method:  "POST",
		Headers: soapUserAgent,
		Curl:    true,
		Body: map[string]string{
			"do":           "search",
			"subaction":    "search",
			"search_start": "0",
			"full_search":  "0",
			"result_from":  "1",
			"story":        "sonic",
		},
	}


	blackbeard.ScrapePage(request, "div.thumbnail.text-center", func(i int, s *goquery.Selection) {
		href := s.Find("div:nth-child(2) > a").AttrOr("href", "")
		title := s.Find("div:nth-child(2) > a").AttrOr("title", "No Title")
		strings.TrimSuffix(title, " Soap2day")

		metadata := blackbeard.Metadata{}
		request = blackbeard.Request{
			Url:     href,
			Method:  "POST",
			Headers: soapUserAgent,
			Curl:    true,
		}
		blackbeard.ScrapePage(request, "body", func(i int, s *goquery.Selection) {
			metadata.ThumbnailUrl = s.Find(
				".hidden-lg > div:nth-child(1) > div:nth-child(1) > img:nth-child(1)",
			).AttrOr("src", "")
			metadata.Description = strings.TrimSpace(s.Find(
				".col-md-7",
			).Text())
			metadata.Description = strings.ReplaceAll(metadata.Description, "\n\n\n", "\n")
			metadata.Description = strings.ReplaceAll(metadata.Description, "\n\n", "\n")
		})
		shows = append(shows, blackbeard.Show{Url: href, Title: title, Metadata: metadata})
	})
	return shows
}

func (a Soap2day) GetEpisodes(show *blackbeard.Show) []blackbeard.Episode {
	url := show.Url
	request := blackbeard.Request{
		Url:     url,
		Headers: soapUserAgent,
		Curl:    true,
	}

	blackbeard.ScrapePage(request, ".player-iframelist > li", func(i int, s *goquery.Selection) {
		title := strconv.Itoa(i + 1)
		href := s.AttrOr("data-playerlink", "")
		show.Episodes = append(show.Episodes, blackbeard.Episode{Title: title, Url: href, Number: i})
	})

	return show.Episodes
}

func (a Soap2day) GetVideo(episode *blackbeard.Episode) blackbeard.Video {
	url := episode.Url
	request := blackbeard.Request{
		Url:     url,
		Headers: soapUserAgent,
		Curl:    true,
	}

	next_path := ""
	blackbeard.ScrapePage(request, "body > div:nth-child(3) > div.twelve.columns > div > div.fourteen.columns > div:nth-child(7) > script:nth-child(2)", func(i int, s *goquery.Selection) {
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

	next_path = soapRootUrl + next_path

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

	apiPath = soapRootUrl + apiPath

	// Request the api
	request.Url = apiPath
	request.Headers["Referer"] = soapRootUrl
	request.Headers["authority"] = "www.wcofun.com"
	request.Headers["pragma"] = "no-cache"
	request.Headers["cache-control"] = "no-cache"
	request.Headers["x-requested-with"] = "XMLHttpRequest"
	request.Headers["sec-gpc"] = "1"
	request.Headers["sec-fetch-site"] = "same-origin"
	request.Headers["sec-fetch-mode"] = "cors"
	request.Headers["sec-fetch-dest"] = "empty"
	request.Headers["accept-language"] = "en-US,en;q=0.9"

	data := CdnResponse{}
	blackbeard.GetJson(request, &data)
	if data.Cdn != "" && data.Enc != "" {
		url = data.Cdn + "/getvid?evid=" + data.Enc
	}

	videoRequest := blackbeard.Request{Url: url, Headers: soapUserAgent}
	episode.Video = blackbeard.Video{Request: videoRequest, Format: "mp4", Name: blackbeard.SanitizeFilename(episode.Title) + ".mp4"}
	return episode.Video
}
