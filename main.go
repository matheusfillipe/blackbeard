// Entry point, argparsing for either CLI or API

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/ktr0731/go-fuzzyfinder"
	blb "github.com/matheusfillipe/blackbeard/blb"
	"github.com/matheusfillipe/blackbeard/providers"
)

var Version = "development"
var BuildDate = "development"

const DEFAULT_PORT = 8080

func completer(d prompt.Document) []prompt.Suggest {
	// TODO read from cache
	s := []prompt.Suggest{
		{Text: "attack on titan", Description: "Very nice one"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

type TuiFlowTemplate interface {
	getProviders() map[string]blb.VideoProvider
	setProvider(provider blb.VideoProvider, name string)
	searchShows(t string) []blb.Show
	getEpisodes(show blb.Show) []blb.Episode
	getVideo(episode blb.Episode) blb.Video
}

type localFlow struct {
	provider blb.VideoProvider
}

func (flow localFlow) getProviders() map[string]blb.VideoProvider {
	return providers.GetProviders()
}

func (flow *localFlow) setProvider(provider blb.VideoProvider, name string) {
	flow.provider = provider
}

func (flow localFlow) searchShows(t string) []blb.Show {
	return flow.provider.SearchShows(t)
}

func (flow localFlow) getEpisodes(show blb.Show) []blb.Episode {
	return flow.provider.GetEpisodes(&show)
}

func (flow localFlow) getVideo(episode blb.Episode) blb.Video {
	return flow.provider.GetVideo(&episode)
}

type apiFlow struct {
	provider    blb.VideoProvider
	baseRequest blb.Request
}

type apiProvider struct {
	Name        string
	BaseRequest blb.Request
}

func (a apiProvider) SearchShows(query string) []blb.Show {
	path := fmt.Sprintf("search?provider=%s&q=%s", a.Name, url.QueryEscape(query))
	data := struct {
		Shows []blb.Show `json:"shows"`
	}{}
	blb.GetJson(a.BaseRequest.New(path), &data)
	return data.Shows
}

func (a apiProvider) GetEpisodes(show *blb.Show) []blb.Episode {
	showurl := show.Url
	path := fmt.Sprintf("episodes?provider=%s&showurl=%s", a.Name, url.QueryEscape(showurl))
	data := struct {
		Episodes []blb.Episode `json:"episodes"`
	}{}
	blb.GetJson(a.BaseRequest.New(path), &data)
	return data.Episodes
}

func (a apiProvider) GetVideo(episode *blb.Episode) blb.Video {
	epurl := episode.Url
	path := fmt.Sprintf("video?provider=%s&epurl=%s", a.Name, url.QueryEscape(epurl))
	data := blb.Video{}
	blb.GetJson(a.BaseRequest.New(path), &data)
	return data
}

func (flow apiFlow) getProviders() map[string]blb.VideoProvider {
	providers := struct {
		Providers []string `json:"providers"`
	}{}
	request := flow.baseRequest.New("providers")
	blb.GetJson(request, &providers)

	resp := make(map[string]blb.VideoProvider)
	for _, value := range providers.Providers {
		resp[value] = apiProvider{}
	}
	return resp
}

func (flow *apiFlow) setProvider(provider blb.VideoProvider, name string) {
	prov := provider.(apiProvider)
	prov.BaseRequest = flow.baseRequest
	prov.Name = name
	flow.provider = prov
}

func (flow apiFlow) searchShows(t string) []blb.Show {
	return flow.provider.SearchShows(t)
}

func (flow apiFlow) getEpisodes(show blb.Show) []blb.Episode {
	return flow.provider.GetEpisodes(&show)
}

func (flow apiFlow) getVideo(episode blb.Episode) blb.Video {
	return flow.provider.GetVideo(&episode)
}

func downloadTuiFlow(flow TuiFlowTemplate) {
	providers := flow.getProviders()
	providerNames := blb.Keys(providers)

	idx, err := fuzzyfinder.Find(
		providerNames,
		func(i int) string {
			return providerNames[i]
		})

	if err != nil {
		log.Fatal(err)
	}

	providerName := providerNames[idx]
	flow.setProvider(providers[providerName], providerName)

	fmt.Println("Search show/anime: ")
	t := prompt.Input("> ", completer)
	if t == "" {
		log.Fatal("No search query")
	}
	shows := flow.searchShows(t)

	// TODO put this in another function and reuse in apiClient
	idx, err = fuzzyfinder.Find(
		shows,
		func(i int) string {
			return shows[i].Title
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			// TODO fix this wrapping
			w /= 2
			w -= 10
			return fmt.Sprintf("Provider: %s\nShow: %s\n\nDescription: %s\n\n\n%s",
				strings.ToUpper(providerName),
				blb.WrapString(shows[i].Title, uint(w)),
				blb.WrapString(shows[i].Metadata.Description, uint(w)),
				blb.WrapStringReguardlessly(shows[i].Metadata.ThumbnailUrl, uint(w)),
			)
		}))

	show := shows[idx]
	episodes := flow.getEpisodes(show)

	// TODO put this in another function and reuse in apiClient
	idxs, err2 := fuzzyfinder.FindMulti(
		episodes,
		func(i int) string {
			return fmt.Sprintf("%v > %v", episodes[i].Number+1, episodes[i].Title)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			// TODO fix this wrapping
			w /= 2
			w -= 10
			return fmt.Sprintf("Provider: %s\nShow: %s\nEpisode n. %d\n\nDescription: %s",
				strings.ToUpper(providerName),
				blb.WrapString(show.Title, uint(w)),
				episodes[i].Number+1,
				blb.WrapString(episodes[i].Metadata.Description, uint(w)),
			)
		}))

	if err2 != nil {
		log.Fatal(err)
	}

	// TODO multitask, parallel downloads
	fmt.Println("...")
	for _, idx := range idxs {
		idx := idx
		episode := episodes[idx]
		video := flow.getVideo(episode)
		video.Download()
	}
}

func apiConnect(url string) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	fmt.Printf("Attempting connection to blackbeard api at %q\n", url)

	// Check if there is a valid reply
	request := blb.Request{Url: url + "version"}
	res, ok := blb.Perform(request)
	if !ok {
		log.Fatal("Connection failed")
	}

	body := res.Body
	defer body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)

	if strings.Contains(buf.String(), "version") {
		println("Connection successful")
	} else {
		log.Fatal("Connection failed")
	}

	flow := apiFlow{}
	flow.baseRequest = blb.Request{Url: url}
	downloadTuiFlow(&flow)
}

func main() {
	defaultPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		defaultPort = DEFAULT_PORT
	}

	// API opts
	apiMode := flag.Bool("api", false, "Start a blackbeard api")
	apiPort := flag.Int("port", defaultPort, "Port to bind to if api. Will also read 'PORT' from env. Default: 8080")
	apiHost := flag.String("host", "0.0.0.0", "Host to bind to if api. Default: 0.0.0.0")

	// Client opts
	connectAddr := flag.String("connect", "0.0.0.0:8080", "Start a client that connects to a blackbeard api with the given address.")

	version := flag.Bool("version", false, "Prints the version then exits")

	flag.Parse()

	if *version {
		fmt.Println("Blackbeard")
		fmt.Println("Version: ", Version)
		fmt.Println("Date: ", BuildDate)
		return
	}

	if *apiMode && connectAddr == nil {
		log.Fatal("Cannot start api and client at the same time")
		return
	}

	if *apiMode {
		startApiServer(*apiHost, *apiPort)
		return
	}

	if connectAddr != nil {
		apiConnect(*connectAddr)
		return
	}

	// Interactive cli
	downloadTuiFlow(&localFlow{})
}
