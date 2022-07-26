package main

import (
	"github.com/matheusfillipe/blackbeard/providers"
	"testing"
)

func TestSoap2day(t *testing.T) {
	t.Run("Wcofun test", func(t *testing.T) {
		a := providers.GetProviders()["soap2day"]
		shows := a.SearchShows("attack on titan")
		if len(shows) < 1 {
			t.Error("No shows found")
		}

		// TODO fix soap2day episodes
		// episodes := a.GetEpisodes(&shows[0])
		// if len(episodes) < 1 {
		// 	t.Error("No episodes found")
		// }

		// video := a.GetVideo(&episodes[0])
		// fmt.Println(video.Request.Url)
	})
}
