package main

import (
	"fmt"
	"testing"
	"blackbeard/providers"
)

func TestSum(t *testing.T) {
	t.Run("9 anime test", func(t *testing.T) {
		a := providers.GetProviders()["9anime"]
		shows := a.SearchShows("attack on titan")
		fmt.Println(shows[1])
		episodes := a.SearchEpisodes(&shows[1], "")
		fmt.Println(episodes)
	})
}
