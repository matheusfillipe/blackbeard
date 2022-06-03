package main

import (
	"blackbeard/providers"
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

		episodes := a.GetEpisodes(&shows[0], "")
		if len(episodes) < 1 {
			t.Error("No episodes found")
		}

		video := a.GetVideo(&episodes[0])
		fmt.Println(video.Url)
	})
}
