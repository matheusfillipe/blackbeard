// Get show and episode info from https://www.episodate.com/api/

package blackbeard

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const epd_api = "https://www.episodate.com/api"

type epdShow struct {
	Id                   int    `json:"id"`
	Name                 string `json:"name"`
	Permalink            string `json:"permalink"`
	Start_date           string `json:"start_date"`
	End_date             string `json:"end_date"`
	Country              string `json:"country"`
	Network              string `json:"network"`
	Status               string `json:"status"`
	Image_thumbnail_path string `json:"image_thumbnail_path"`
}

type epdShowRes struct {
	Tv_shows []epdShow `json:"tv_shows"`
}

type epdEpisodes struct {
	Season   int    `json:"season"`
	Episode  int    `json:"episode"`
	Name     string `json:"name"`
	Air_date string `json:"air_date"`
}

type epdShowDetails struct {
	Id          int           `json:"id"`
	Name        string        `json:"name"`
	Url         string        `json:"url"`
	Description string        `json:"description"`
	Start_date  string        `json:"start_date"`
	End_date    string        `json:"end_date"`
	Country     string        `json:"country"`
	Runtime     int           `json:"runtime"`
	Image_path  string        `json:"image_path"`
	Rating      json.Number   `json:"rating"`
	Genres      []string      `json:"genres"`
	Episodes    []epdEpisodes `json:"episodes"`
}

type epdTvShow struct {
	TvShow epdShowDetails `json:"tvShow"`
}

func getShowDetails(show *Show) epdShowDetails {
	name := strings.TrimSuffix(show.Title, " English Subbed")
	showDetails := epdShowDetails{}
	showRes := epdShowRes{}
	request := Request{
		Url: epd_api + "/search?q=" + url.QueryEscape(name) + "&page=1",
	}
	GetJson(request, &showRes)
	if len(showRes.Tv_shows) < 1 {
		// println("No episode metadata found for \"" + show.Title + "\"")
		return showDetails
	}
	showId := showRes.Tv_shows[0].Id

	tvShow := epdTvShow{}
	request.Url = epd_api + "/show-details?q=" + fmt.Sprintf("%d", showId)
	GetJson(request, &tvShow)
	showDetails = tvShow.TvShow
	return showDetails
}

func populateEpisodes(show *Show, showDetails epdShowDetails) {
	for i := 0; i < len(showDetails.Episodes) && i < len(show.Episodes); i++ {
		epDetails := showDetails.Episodes[i]
		show.Episodes[i].Metadata = Metadata{
			Description: fmt.Sprintf("Season: %d\nName: %s\nAiring Date: %s", epDetails.Season, epDetails.Name, epDetails.Air_date),
		}
	}
}

// Gets show description from episodate api
// Fills in the show metadata
func EpisodatePopulateShowMetadata(show *Show) {
	showDetails := getShowDetails(show)
	show.Metadata.Description = showDetails.Description
	show.Metadata.ThumbnailUrl = showDetails.Image_path
}

// Gets episode description from episodate api
// Fills in the episodes metadata of the show.Episodes array
func EpisodatePopulateEpisodesMetadata(show *Show) {
	showDetails := getShowDetails(show)
	populateEpisodes(show, showDetails)
}
