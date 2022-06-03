// Entry point, argparsing for either CLI or API

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/matheusfillipe/blackbeard/providers"
	"github.com/c-bata/go-prompt"
)


func completer(d prompt.Document) []prompt.Suggest {
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
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Provider %s",
				providerNames[i])
		}))

	if err != nil {
		log.Fatal(err)
	}

	provider := providers[providerNames[idx]]

  fmt.Println("Choose show/anime to search for: ")
  t := prompt.Input("> ", completer)
	shows := provider.SearchShows(t)

	idx, err = fuzzyfinder.Find(
		shows,
		func(i int) string {
			return shows[i].Title
		})

  show := shows[idx]
	episodes := provider.GetEpisodes(&show, "")

	idx, err = fuzzyfinder.Find(
		episodes,
		func(i int) string {
			return fmt.Sprintf("%v > %v", episodes[i].Number, episodes[i].Title)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Show %s",
				episodes[i].Title)
		}))

	if err != nil {
		log.Fatal(err)
	} 

  episode := episodes[idx]
	fmt.Printf("Selected: %v\n", episode.Title)

	video := provider.GetVideo(&episode)
	fmt.Print("Downloading: ", video.Url)
	video.Download()
}
