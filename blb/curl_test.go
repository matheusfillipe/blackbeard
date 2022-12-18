package blackbeard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				if r.URL.Path == "/gettest" {
					w.Header().Add("Content-Type", "application/text")
					time.Sleep(100 * time.Millisecond)
					for i := 1; i <= 100; i++ {
						w.Write([]byte("hello world\n"))
					}
					time.Sleep(100 * time.Millisecond)
					for i := 1; i <= 100; i++ {
						w.Write([]byte("hello world\n"))
					}
				}
			} else if r.Method == "POST" {
				w.Header().Add("Content-Type", "application/text")
				// Return the user agent
				w.Write([]byte(r.Header.Get("User-Agent")))
				w.Write([]byte("\n"))
				// Then return the body
				w.Write([]byte(r.FormValue("Ping")))
			}
		}),
	)
}

func TestCurl(t *testing.T) {
	t.Run("Get request", func(t *testing.T) {
		ts := mockServer()
		defer ts.Close()
		println(ts.URL)
		body, ok, _ := Curl(ts.URL + "/gettest")

		if !ok {
			t.Error("Curl GET failed. Not ok")
		}

		if len(body) != len("hello world\n")*200 {
			t.Error("Curl GET failed. Body length is not 200")
		}

		if body != strings.Repeat("hello world\n", 200) {
			t.Error("Curl GET failed. Body is not 200 times hello world")
		}

	})

	t.Run("Post request", func(t *testing.T) {
		ts := mockServer()
		defer ts.Close()
		ua := map[string]string{"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:96.0) Gecko/20100101 Firefox/96.0"}
		request := Request{Url: ts.URL, Method: "POST", Headers: ua, Body: map[string]string{"Ping": "Pong"}}
		body, ok, _ := Curl(request)
		respList := strings.Split(body, "\n")

		if !ok {
			t.Error("Curl POST failed. Not ok")
		}

		if respList[0] != ua["User-Agent"] {
			t.Error("Curl POST failed. User-Agent not returned")
		}

		if string(respList[1]) != "Pong" {
			t.Error("Curl POST failed. Pong not returned")
		}

	})
}
