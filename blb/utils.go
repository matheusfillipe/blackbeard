// Tools to use while scrapping or everywhere
// Just things I wrote and find useful

package blackbeard

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/cavaliergopher/grab/v3"
	"github.com/kennygrant/sanitize"
	runewidth "github.com/mattn/go-runewidth"
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

// hacky debug log that writes to /tmp/debug.txt
// tail -f /tmp/debug.txt
func DebugLog(vars ...any) {
	f, err := os.OpenFile("/tmp/debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	defer f.Close()

	if err != nil {
		log.Fatal(err)
	}
	for _, i := range vars {
		f.WriteString(fmt.Sprintf("%+v", i))
	}
	f.WriteString("\n")
}

// Hacky breakpoint that prints a message
func Breakpoint(vars ...any) {
	for _, i := range vars {
		fmt.Printf("%+v ", i)
	}
	fmt.Print("\n")
	fmt.Println("Press return to continue...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
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

// Return list of keys of a map
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for name := range m {
		keys[i] = name
		i++
	}
	return keys
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

		// // When in pain
		// println("------------------------------------------------------")
		// println(_body)
		// println("------------------------------------------------------")
		// fmt.Printf("%+v\n", request)

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

// Downloads a video to the given directory
// linepos is the position to print the download progress line in
func (video Video) Download(dir string, linepos int) bool {
	// create client
	client := grab.NewClient()
	req, err := grab.NewRequest(".", video.Request.Url)
	if err != nil {
		return false
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req = req.WithContext(ctx)

	for key, value := range video.Request.Headers {
		req.HTTPRequest.Header.Set(key, value)
	}

	// start download
	name := SanitizeFilename(video.Name)
	name = dir + "/" + name
	req.Filename = filepath.FromSlash(name)
	resp := client.Do(req)
	// fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

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

			fmt.Printf("\033[H\033[%dB\033[K", linepos+1)
			fmt.Printf("%d > Downloading to %s: %.2f%s - %.2fMB/%.2fMB (%.2f%%)\n",
				linepos+1,
				resp.Filename,
				speed,
				unit,
				float64(resp.BytesComplete())/1024/1024,
				float64(resp.Size())/1024/1024,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	fmt.Printf("%d > \033[H\033[%dB\033[K", linepos+1, linepos+1)

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Printf("Download failed: %v\n", err.Error())
		return false
	}
	fmt.Printf("%d > Finished %s: %.2fMB/%.2fMB (100%%)\n",
		linepos+1,
		resp.Filename,
		float64(resp.BytesComplete())/1024/1024,
		float64(resp.Size())/1024/1024)

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
func WrapString(s string, wantedWidth uint) string {
	// Initialize a buffer with a slightly larger size to account for breaks
	init := make([]byte, 0, len(s))
	buf := bytes.NewBuffer(init)

	var current uint
	var wordBuf, spaceBuf bytes.Buffer
	var wordBufLen, spaceBufLen uint
	lim := Max(uint(runewidth.StringWidth(s))/wantedWidth, 1) // Number of lines
	lim = uint(len(s)) / lim                                  // Number of characters per line

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
func WrapStringReguardlessly(s string, wantedWidth int) string {
	// Initialize a buffer with a slightly larger size to account for breaks
	init := make([]byte, 0, len(s))
	buf := bytes.NewBuffer(init)

	var count int

	for _, char := range s {
		if count >= wantedWidth {
			buf.WriteRune('\n')
			count = 0
		} else if char == '\n' {
			buf.WriteRune(char)
			count = 0
		} else {
			buf.WriteRune(char)
			count += runewidth.RuneWidth(char)
		}
	}

	return buf.String()

}

// Finds the maximum value among the arguments
func Max[T Number](vars ...T) T {
	max := vars[0]

	for _, i := range vars {
		if max < i {
			max = i
		}
	}

	return max
}

// Finds the minimum value among the arguments
func Min[T Number](vars ...T) T {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}

// Sums all arguments
func Sum[T Number](vars ...T) T {
	var result T
	for _, value := range vars {
		result += value
	}
	return result
}

// Run function for each element array, modifying it
func Map[T Number, O Number](vars []T, f func(v T) O) []O {
	var result []O
	for _, value := range vars {
		result = append(result, f(value))
	}
	return result
}

// Check if a value is the default
func IsDefault[T Number | string | bool](v T) bool {
	switch r := any(v).(type) {
	case int, int8, int16, int32, int64:
		return r == 0
	case float64, float32:
		return r == 0.0
	case bool:
		return r == false
	case string:
		return r == ""
	default:
		return false
	}
}

// Check if an array contains value
func Contains[T comparable](array []T, value T) bool {
	for _, i := range array {
		if i == value {
			return true
		}
	}
	return false
}

// Find index in array
func IndexOf[T comparable](array []T, value T) int {
	for i, v := range array {
		if v == value {
			return i
		}
	}
	return -1
}

// Repeat String
func Repeat(s string, times int) string {
	res := ""
	for i := 0; i < times; i++ {
		res += s
	}
	return res
}

// Run a function and return nil, false if it timeouts. otherwise returns f(), true
// timeout in seconds
func Timeout[R any](timeout int, f func() R) (R, bool) {
	c := make(chan int)
	resc := make(chan R)
	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		c <- 1
	}()
	go func() {
		resc <- f()
	}()

	select {
	case res := <-resc:
		return res, true
	case <-c:
		var res R
		return res, false
	}
}
