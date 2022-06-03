// Entry point, argparsing for either CLI or API

package main

import (
	"fmt"
	"blackbeard/providers"
)

func main() {
		a := providers.GetProviders()["wcofun"]
		shows := a.SearchShows("attack on titan")
		fmt.Println("SHOWS: ", shows)
		episodes := a.SearchEpisodes(&shows[0], "")
		fmt.Println("EPS: ", episodes)
}
