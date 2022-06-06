// Main interfaces for blackbeard providers, shows and requests

package blackbeard

// Non essential information about shows, movies or episodes.
type Metadata struct {
	Description  string
	ThumbnailUrl string
}

// An episode of a tv show
type Episode struct {
	// Title of the episode
	Title    string
	// Unique Number of the episode that follows the order of the series
	Number   int
  // Url to the episode in the providers website
	Url      string
  // Video struct for this episode
	Video    Video
  // Metadata for this episode
	Metadata Metadata
}

// Can represent a Tv Show or a Movie
type Show struct {
  // Title of the show
	Title    string
  // Url to the show in the providers website
	Url      string
  // List of episodes for this show
	Episodes []Episode
  // If set to true it is a movie provider, meaning there are no episodes and shows
  // are movies. By default it will be a show, meaning it has episodes.
	IsMovie bool
  // Metadata for this show
	Metadata Metadata
}

// Packages a request to the providers api to fetch the video for an episode or movie
// Also information about the file to be downloaded
type Video struct {
  // Name of the video file to be downloaded
	Name     string
  // Format of the video file
	Format   string
  // Request to the providers api to fetch the video
	Request  Request
  // Metadata for the video
	Metadata Metadata
}

// Information for providers, similar to Metadata for episodes
type ProviderInfo struct {
  // Name of the provider
	Name   string
  // Url to the providers website
  Url string
  // Description for the provider
  Description string
}

// Interface for video providers
type VideoProvider interface {
  // Get the list of shows for this provider.
  // If the provider is a movie provider, this will alerady be the movie list
	SearchShows(string) []Show
  // Get the list of episodes for a show and populates Show.Episodes.
  // If show.IsMovie then should return a single episode in the list, others
  // will be ignored.
	GetEpisodes(*Show) []Episode
  // Get the video for an episode and populates Episode.Video
	GetVideo(*Episode) Video
  // Get the provider information.
	Info() ProviderInfo
}

type Request struct {
  // url for the request
	Url     string
  // method for the request
	Method  string
  // headers for the request
	Headers map[string]string
  // body for the request
  Body    map[string]string
  // If set to true will use curl instead of go's stdlib http module
  // Useful in the cases there is a basic cloudflare protection and
  // curl-impersonate can be used
  Curl    bool
}

// Create a new request from an existing one, appending to the url
func (r Request) New(path string) Request {
	return Request{
		Url:     r.Url + path,
		Method:  r.Method,
		Headers: r.Headers,
		Body:    r.Body,
		Curl:    r.Curl,
	}
}
