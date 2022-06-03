package main

import (
	"fmt"
	"testing"
	"blackbeard/providers"
)

func TestNineAnime(t *testing.T) {
	t.Run("9 anime test", func(t *testing.T) {
		a := providers.GetProviders()["9anime"]
		shows := a.SearchShows("attack on titan")
		fmt.Println("Shows: ", shows[1])
		episodes := a.GetEpisodes(&shows[1], "")
		fmt.Println(episodes)
	})
}
