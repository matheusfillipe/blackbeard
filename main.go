// Entry point, argparsing for either CLI or API

package main

import (
	"fmt"

	"github.com/matheusfillipe/blackbeard/providers"
)

func main() {
	a := providers.GetProviders()["wcofun"]
	shows := a.SearchShows("attack on titan")
	episodes := a.GetEpisodes(&shows[0], "")
	video := a.GetVideo(&episodes[0])
	fmt.Print("Downloading: ", video.Url)
	video.Download()
}
