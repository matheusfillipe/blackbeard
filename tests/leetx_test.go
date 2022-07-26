package main

import (
	"fmt"
	"github.com/matheusfillipe/blackbeard/providers"
	"testing"
)

func TestLeetx(t *testing.T) {
	t.Run("leetx test", func(t *testing.T) {
		return
		a := providers.GetProviders()["1337x"]
		shows := a.SearchShows("attack on titan")
		fmt.Println("Shows: ", shows[0])
		episodes := a.GetEpisodes(&shows[0])
		fmt.Println(episodes)
	})
}
