// Wrapper around go-curl. Only caring for GET and POST.

package blackbeard

import (
	"fmt"
	"net/url"
	"strings"

	gocurl "github.com/andelf/go-curl"
)

// Performs a request using libcurl
func curl[R Request | string](_request R) (string, bool) {
	var body string
	easy := gocurl.EasyInit()
	defer easy.Cleanup()

	var request Request
	switch r := any(_request).(type) {
	case Request:
		request = r
	case string:
		request = Request{Url: r}
	}

	easy.Setopt(gocurl.OPT_URL, request.Url)
	easy.Setopt(gocurl.OPT_VERBOSE, false)

	easy.Setopt(gocurl.OPT_WRITEFUNCTION, func(buf []byte, userdata interface{}) bool {
		body += string(buf)
		return true
	})

	if request.Method == "POST" {
		setup_post(easy, request)
	}

	var headers []string
	for key, value := range request.Headers {
		headers = append(headers, fmt.Sprintf("%s: %s", key, value))

	}
	easy.Setopt(gocurl.OPT_HTTPHEADER, headers)

	if err := easy.Perform(); err != nil {
		println("ERROR: ", err.Error())
		return "", false
	}
	return body, true
}

func setup_post(easy *gocurl.CURL, request Request) {
	var sent = false

	easy.Setopt(gocurl.OPT_POST, true)

	post_data := ""
	for key, value := range request.Body {
		post_data += fmt.Sprintf("%s=%v&", key, url.QueryEscape(value))
	}

	if post_data != "" {
		post_data = strings.TrimSuffix(post_data, "&")
		easy.Setopt(gocurl.OPT_READFUNCTION,
			func(ptr []byte, userdata interface{}) int {
				// WARNING: never use append()
				if !sent {
					sent = true
					ret := copy(ptr, post_data)
					return ret
				}
				return 0 // sent ok
			})
	}

	easy.Setopt(gocurl.OPT_POSTFIELDSIZE, len(post_data))
}
