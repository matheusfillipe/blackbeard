package main

import (
	"blackbeard/providers"
	"blackbeard/blb"
	"fmt"
	"testing"
)

func TestWcofun(t *testing.T) {
	t.Run("Wcofun test", func(t *testing.T) {
		a := providers.GetProviders()["wcofun"]
		shows := a.SearchShows("attack on titan")
		if len(shows) < 1 {
			t.Error("No shows found")
		}

		episodes := a.SearchEpisodes(&shows[0], "")
		if len(episodes) < 1 {
			t.Error("No episodes found")
		}

		episode := blackbeard.Episode{Url: "https://www.wcofun.com/attack-on-titan-episode-1-english-dubbed-3"}
		video := a.GetVideo(&episode)
		fmt.Println(video.Url)
	})
}
