// Entry point, argparsing for either CLI or API

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/matheusfillipe/blackbeard/providers"
)

func completer(d prompt.Document) []prompt.Suggest {
	// TODO read from cache
	s := []prompt.Suggest{
		{Text: "attack on titan", Description: "Very nice one"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	apiMode := flag.Bool("api", false, "Start api")
	apiPort := flag.Int("port", 8080, "Port to bind to")
	apiHost := flag.String("host", "0.0.0.0", "Host to bind to")
	flag.Parse()
	if *apiMode {
		startApiServer(*apiHost, *apiPort)
		return
	}
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

	fmt.Println("Choose show/anime to search for: ")
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
			return fmt.Sprintf("Provider: %s\nShow: %s\n\nDescription: %s\n\n\n%s",
				strings.ToUpper(providerName),
				shows[i].Title,
				shows[i].Metadata.Description,
				shows[i].Metadata.ThumbnailUrl,
			)
		}))

	show := shows[idx]
	episodes := provider.GetEpisodes(&show, "")

	idx, err = fuzzyfinder.Find(
		episodes,
		func(i int) string {
			return fmt.Sprintf("%v > %v", episodes[i].Number + 1, episodes[i].Title)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Provider: %s\nShow: %s\nEpisode n. %d\n\nDescription: %s",
				strings.ToUpper(providerName),
				show.Title,
				episodes[i].Number+1,
				episodes[i].Metadata.Description,
			)
		}))

	if err != nil {
		log.Fatal(err)
	}

	episode := episodes[idx]
	fmt.Printf("Selected: %v\n", episode.Title)

	video := provider.GetVideo(&episode)
	fmt.Print("Downloading: ", video.Request.Url)
	video.Download()
}
