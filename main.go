// Entry point, argparsing for either CLI or API

package main

import (
	"fmt"
	"blackbeard/providers"
)

func main() {
		a := providers.GetProviders()["wcofun"]
		shows := a.SearchShows("attack on titan")
		fmt.Println(shows[1])
		episodes := a.SearchEpisodes(&shows[1], "")
		fmt.Println(episodes)
}
