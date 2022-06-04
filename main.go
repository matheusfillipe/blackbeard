// Entry point, argparsing for either CLI or API

package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/ktr0731/go-fuzzyfinder"
	blb "github.com/matheusfillipe/blackbeard/blb"
	"github.com/matheusfillipe/blackbeard/providers"
)

func completer(d prompt.Document) []prompt.Suggest {
	// TODO read from cache
	s := []prompt.Suggest{
		{Text: "attack on titan", Description: "Very nice one"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func downloadCli() {
	providers := providers.GetProviders()
	providerNames := make([]string, len(providers))
	i := 0
	for name := range providers {
		providerNames[i] = name
		i++
	}

	idx, err := fuzzyfinder.Find(
		providerNames,
		func(i int) string {
			return providerNames[i]
		})

	if err != nil {
		log.Fatal(err)
	}

	providerName := providerNames[idx]
	provider := providers[providerName]

	fmt.Println("Search show/anime: ")
	t := prompt.Input("> ", completer)
	shows := provider.SearchShows(t)

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
	episodes := provider.GetEpisodes(&show, "")

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

	// TODO multitask downloads
	fmt.Println("...")
	for _, idx := range idxs {
		idx := idx
		episode := episodes[idx]
		video := provider.GetVideo(&episode)
		video.Download()
	}
}

func apiConnect(host string, port int) {
	println("Attempting connection to blackbeard api at " + host + ":" + strconv.Itoa(port))
}

var Version = "development"
var BuildDate = "development"

func main() {
	// API opts
	apiMode := flag.Bool("api", false, "Start a blackbeard api")
	apiPort := flag.Int("port", 8080, "Port to bind to if api or to connect to if client. Default: 8080")
	apiHost := flag.String("host", "0.0.0.0", "Host to bind to if api or to connect to if client. Default: 0.0.0.0")

	// Client opts
	clientMode := flag.Bool("connect", false, "Start a client that connects to a blackbeard api")

	version := flag.Bool("version", false, "Prints the version then exits")

	flag.Parse()

	if *version {
		fmt.Println("Blackbeard")
		fmt.Println("Version: ", Version)
		fmt.Println("Date: ", BuildDate)
		return
	}

	if *apiMode && *clientMode {
		log.Fatal("Cannot start api and client at the same time")
		return
	}

	if *apiMode {
		startApiServer(*apiHost, *apiPort)
		return
	}

	if *clientMode {
		apiConnect(*apiHost, *apiPort)
		return
	}

	// Interactive cli
	downloadCli()
}
