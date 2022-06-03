// Tools to use while scrapping

package blackbeard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cavaliergopher/grab/v3"
)

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	res := map[K]V{}
	for _, m := range maps {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

// Perform a request using standart http
func Perform(request Request) (*http.Response, bool) {
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
		return nil, false
	}

	if res.StatusCode != 200 {
		return nil, false
	}
	return res, true
}

func ScrapePage(request Request, selector string, handler func(int, *goquery.Selection)) {
	var body io.Reader

	// Load the HTML document
	if request.Curl {
		_body, ok := curl(request)
		if !ok {
			println("Could not load page ", request.Url)
			return
		}
		// Convert to io.Reader
		body = strings.NewReader(_body)
	} else {
		res, ok := Perform(request)
		defer res.Body.Close()

		if !ok {
			println("Could not load page ", request.Url)
			return
		}
		body = res.Body
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		println("Could not parse page ", request.Url)
		return
	}

	// Iterate over selector matches
	doc.Find(selector).Each(handler)
}

// Parses a string into a goquery selection object and call handler on it
func Soup(text string, selector string, handler func(int, *goquery.Selection)) {
	body := strings.NewReader(text)
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		println("Could not parse text")
		return
	}
	// Iterate over selector matches
	doc.Find(selector).Each(handler)
}

// Get Json request
func GetJson[T any](request Request, data T) T {
	request.Headers["accept"] = "application/json, text/javascript, */*; q=0.01"
	body, ok := curl(request)
	if !ok {
		println("Could not load page ", request.Url)
		return data
	}
	err := json.Unmarshal([]byte(body), data)
	if err != nil {
		println("Could not parse json ")
		println(body)
		println(err.Error())
		return data
	}
	return data
}

// Downloads a video
func (video Video) Download() bool {
	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(".", video.Url)

	for key, value := range video.Headers {
		req.HTTPRequest.Header.Set(key, value)
	}

	// start download
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size(),
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		println("Download failed: %v\n", err.Error())
		return false
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)
	return true
}

func SanitizeFilename(name string) string {
  name = strings.Replace(name, ":", "", -1)
  name = strings.Replace(name, "?", "", -1)
  name = strings.Replace(name, "=", "", -1)
  name = strings.Replace(name, "&", "", -1)
  name = strings.Replace(name, "/", "", -1)
  name = strings.Replace(name, "\\", "", -1)
  name = strings.Replace(name, "*", "", -1)
  name = strings.Replace(name, "\"", "", -1)
  name = strings.Replace(name, "<", "", -1)
  name = strings.Replace(name, ">", "", -1)
  name = strings.Replace(name, "|", "", -1)
  name = strings.Replace(name, "!", "", -1)
  name = strings.Replace(name, "`", "", -1)
  name = strings.Replace(name, "~", "", -1)
  name = strings.Replace(name, ",", "", -1)
  name = strings.Replace(name, "'", "", -1)
  name = strings.Replace(name, ".", "", -1)
  name = strings.Replace(name, ";", "", -1)
  name = strings.Replace(name, " ", "-", -1)
  return name
}
