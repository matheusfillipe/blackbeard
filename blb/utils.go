// Tools to use while scrapping or everywhere

package blackbeard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/cavaliergopher/grab/v3"
	"github.com/kennygrant/sanitize"
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
	// Check if request.Headers map is nil before assigning to it
	if request.Headers == nil {
		request.Headers = map[string]string{}
	}
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
	req, _ := grab.NewRequest(".", video.Request.Url)
  ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
  req = req.WithContext(ctx)

	for key, value := range video.Request.Headers {
		req.HTTPRequest.Header.Set(key, value)
	}

	// start download
	name := SanitizeFilename(video.Name)
	req.Filename = name
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

	c := make(chan os.Signal, 1)
	wait := make(chan struct{})
	signal.Notify(c, os.Interrupt)

	fmt.Println("")
Loop:
	for {
		select {
		case <-t.C:
			speed := resp.BytesPerSecond() / 1024 // in KB/s
			unit := "KB/s"
			speedMb := speed / 1024 // in MB/s

			// If speed is greater than 1MB/s, print it in MB/s
			if speedMb > 1 {
				speed = speedMb
				unit = "MB/s"
			}

			fmt.Print("\033[1A\033[K")
			fmt.Printf("Downloading to %s: %.2f%s - %.2fMB/%.2fMB (%.2f%%)\n",
				resp.Filename,
				speed,
				unit,
				float64(resp.BytesComplete())/1024/1024,
				float64(resp.Size())/1024/1024,
				100*resp.Progress())

		case <-c:
			// manual cancel here
			fmt.Print("Cancelling")

			// print some dots to indicate waiting period after calling cancel
			go func() {
				for {
					select {
					case <-wait:
						return
					default:
						fmt.Print(".")
						time.Sleep(time.Second)
					}
				}
			}()

			cancel()
			break Loop

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

func SanitizeFilename(filename string) string {
	return sanitize.Path(filename)
}



// TAKEN FROM https://github.com/mitchellh/go-wordwrap
const nbsp = 0xA0

// WrapString wraps the given string within lim width in characters.
//
// Wrapping is currently naive and only happens at white-space. A future
// version of the library will implement smarter wrapping. This means that
// pathological cases can dramatically reach past the limit, such as a very
// long word.
func WrapString(s string, lim uint) string {
	// Initialize a buffer with a slightly larger size to account for breaks
	init := make([]byte, 0, len(s))
	buf := bytes.NewBuffer(init)

	var current uint
	var wordBuf, spaceBuf bytes.Buffer
	var wordBufLen, spaceBufLen uint

	for _, char := range s {
		if char == '\n' {
			if wordBuf.Len() == 0 {
				if current+spaceBufLen > lim {
					current = 0
				} else {
					current += spaceBufLen
					spaceBuf.WriteTo(buf)
				}
				spaceBuf.Reset()
				spaceBufLen = 0
			} else {
				current += spaceBufLen + wordBufLen
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				spaceBufLen = 0
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
				wordBufLen = 0
			}
			buf.WriteRune(char)
			current = 0
		} else if unicode.IsSpace(char) && char != nbsp {
			if spaceBuf.Len() == 0 || wordBuf.Len() > 0 {
				current += spaceBufLen + wordBufLen
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				spaceBufLen = 0
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
				wordBufLen = 0
			}

			spaceBuf.WriteRune(char)
			spaceBufLen++
		} else {
			wordBuf.WriteRune(char)
			wordBufLen++

			if current+wordBufLen+spaceBufLen > lim && wordBufLen < lim {
				buf.WriteRune('\n')
				current = 0
				spaceBuf.Reset()
				spaceBufLen = 0
			}
		}
	}

	if wordBuf.Len() == 0 {
		if current+spaceBufLen <= lim {
			spaceBuf.WriteTo(buf)
		}
	} else {
		spaceBuf.WriteTo(buf)
		wordBuf.WriteTo(buf)
	}

	return buf.String()
}

// Wrap reguardles of spaces
func WrapStringReguardlessly(s string, width uint) string {
  // Initialize a buffer with a slightly larger size to account for breaks
  init := make([]byte, 0, len(s))
  buf := bytes.NewBuffer(init)

  var count uint

  for _, char := range s {
    if count + 1 > uint(width) {
      buf.WriteRune('\n')
      count = 0
    } else if char == '\n' {
      buf.WriteRune(char)
      count = 0
    } else {
      buf.WriteRune(char)
      count++
    }
  }

  return buf.String()

}
