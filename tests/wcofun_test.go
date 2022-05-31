package main

import (
	"fmt"
	"testing"
	"blackbeard/providers"
)

func TestWcofun(t *testing.T) {
	t.Run("Wcofun test", func(t *testing.T) {
		a := providers.GetProviders()["wcofun"]
		shows := a.SearchShows("attack on titan")
		fmt.Println(shows[1])
		episodes := a.SearchEpisodes(&shows[1], "")
		fmt.Println(episodes)
	})
}
