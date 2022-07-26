package main

import (
	"fmt"
	"github.com/matheusfillipe/blackbeard/providers"
	"testing"
)

func TestLeetx(t *testing.T) {
	t.Run("leetx text", func(t *testing.T) {
		a := providers.GetProviders()["1337x"]
		shows := a.SearchShows("attack on titan")
		fmt.Println("Shows: ", shows[1])
		episodes := a.GetEpisodes(&shows[1])
		fmt.Println(episodes)
	})
}
